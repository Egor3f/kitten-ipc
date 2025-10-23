package kittenipc

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/samber/mo"
)

const ipcSocketArg = "--ipc-socket"

type StdioMode int

type ipcMode int

const (
	modeParent ipcMode = 1
	modeChild  ipcMode = 2
)

type KittenIPC struct {
	mode     ipcMode
	cmd      *exec.Cmd
	localApi any

	socketPath string
	listener   net.Listener
	conn       net.Conn
	errCh      chan error

	nextId       int64
	pendingCalls map[int64]chan callResult
	mu           sync.Mutex
}

func NewParent(cmd *exec.Cmd, localApi any) (*KittenIPC, error) {
	k := KittenIPC{
		mode:         modeParent,
		cmd:          cmd,
		localApi:     localApi,
		pendingCalls: make(map[int64]chan callResult),
		errCh:        make(chan error, 1),
	}

	k.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("kitten-ipc-%d.sock", os.Getpid()))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if slices.Contains(cmd.Args, ipcSocketArg) {
		return nil, fmt.Errorf("you should not use `%s` argument in your command", ipcSocketArg)
	}
	cmd.Args = append(cmd.Args, ipcSocketArg, k.socketPath)

	k.errCh = make(chan error, 1)

	return &k, nil
}

func NewChild(localApi any) (*KittenIPC, error) {
	k := KittenIPC{
		mode:         modeChild,
		localApi:     localApi,
		pendingCalls: make(map[int64]chan callResult),
		errCh:        make(chan error, 1),
	}

	socketPath := flag.String("ipc-socket", "", "Path to IPC socket")
	flag.Parse()

	if *socketPath == "" {
		return nil, fmt.Errorf("ipc socket path is missing")
	}
	k.socketPath = *socketPath

	return &k, nil
}

func (k *KittenIPC) Start() error {
	if k.mode == modeParent {
		return k.startParent()
	}
	return k.startChild()
}

func (k *KittenIPC) startParent() error {
	_ = os.Remove(k.socketPath)
	listener, err := net.Listen("unix", k.socketPath)
	if err != nil {
		return fmt.Errorf("listen unix socket: %w", err)
	}
	k.listener = listener
	defer k.closeSock()

	err = k.cmd.Start()
	if err != nil {
		return fmt.Errorf("cmd start: %w", err)
	}

	if err := k.acceptConn(); err != nil {
		return fmt.Errorf("accept connection: %w", err)
	}

	return nil
}

func (k *KittenIPC) startChild() error {
	conn, err := net.Dial("unix", k.socketPath)
	if err != nil {
		return fmt.Errorf("connect to parent socket: %w", err)
	}
	k.conn = conn
	k.startRcvData()
	return nil
}

func (k *KittenIPC) acceptConn() error {
	const acceptTimeout = time.Second * 10

	res := make(chan mo.Result[net.Conn], 1)
	go func() {
		conn, err := k.listener.Accept()
		if err != nil {
			res <- mo.Err[net.Conn](err)
		} else {
			res <- mo.Ok[net.Conn](conn)
		}
		close(res)
	}()

	select {
	case <-time.After(acceptTimeout):
		_ = k.cmd.Process.Kill()
		return fmt.Errorf("accept timeout")
	case res := <-res:
		if res.IsError() {
			_ = k.cmd.Process.Kill()
			return fmt.Errorf("accept: %w", res.Error())
		}
		k.conn = res.MustGet()
		k.startRcvData()
	}
	return nil
}

type MsgType int

const (
	MsgCall     MsgType = 1
	MsgResponse MsgType = 2
)

type Message struct {
	Type   MsgType `json:"type"`
	Id     int64   `json:"id"`
	Method string  `json:"method"`
	Params []any   `json:"params"`
	Result []any   `json:"result"`
	Error  string  `json:"error"`
}

type callResult struct {
	result []any
	err    error
}

func (k *KittenIPC) startRcvData() {
	scn := bufio.NewScanner(k.conn)
	for scn.Scan() {
		var msg Message
		if err := json.Unmarshal(scn.Bytes(), &msg); err != nil {
			k.raiseErr(fmt.Errorf("unmarshal message: %w", err))
			break
		}
		k.processMsg(msg)
	}
	if err := scn.Err(); err != nil {
		k.raiseErr(err)
	}
}

func (k *KittenIPC) processMsg(msg Message) {
	switch msg.Type {
	case MsgCall:
		k.handleCall(msg)
	case MsgResponse:
		k.handleResponse(msg)
	}
}

func (k *KittenIPC) handleCall(msg Message) {

	if k.localApi == nil {
		k.sendResponse(msg.Id, nil, fmt.Errorf("remote side does not accept ipc calls"))
	}
	localApi := reflect.ValueOf(k.localApi)

	method := localApi.MethodByName(msg.Method)
	if !method.IsValid() {
		k.sendResponse(msg.Id, nil, fmt.Errorf("method not found: %s", msg.Method))
		return
	}

	methodType := method.Type()
	argsCount := methodType.NumIn()

	if len(msg.Params) != argsCount {
		k.sendResponse(msg.Id, nil, fmt.Errorf("argument count mismatch: expected %d, got %d", argsCount, len(msg.Params)))
		return
	}

	var args []reflect.Value
	for _, param := range msg.Params {
		args = append(args, reflect.ValueOf(param))
	}

	results := method.Call(args)
	resVals := results[0 : len(results)-1]
	resErr := results[len(results)-1]

	var res []any
	for _, resVal := range resVals {
		res = append(res, resVal)
	}

	k.sendResponse(msg.Id, res, resErr.Interface().(error))
}

func (k *KittenIPC) handleResponse(msg Message) {

	k.mu.Lock()
	ch, ok := k.pendingCalls[msg.Id]
	if ok {
		delete(k.pendingCalls, msg.Id)
	}
	k.mu.Unlock()

	if !ok {
		k.raiseErr(fmt.Errorf("received response for unknown call id: %d", msg.Id))
		return
	}

	var err error
	if msg.Error != "" {
		err = fmt.Errorf("remote error: %s", msg.Error)
	}

	ch <- callResult{result: msg.Result, err: err}
	close(ch)
}

func (k *KittenIPC) sendResponse(id int64, result []any, err error) {
	msg := Message{
		Type:   MsgResponse,
		Id:     id,
		Result: result,
	}

	if err != nil {
		msg.Error = err.Error()
	}

	if err := k.sendMsg(msg); err != nil {
		k.raiseErr(fmt.Errorf("send response for id=%d: %w", id, err))
	}
}

func (k *KittenIPC) sendMsg(msg Message) error {

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	data = append(data, '\n')

	if _, err := k.conn.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (k *KittenIPC) Call(method string, params ...any) ([]any, error) {
	k.mu.Lock()
	id := k.nextId
	k.nextId++
	resChan := make(chan callResult, 1)
	k.pendingCalls[id] = resChan
	k.mu.Unlock()

	msg := Message{
		Type:   MsgCall,
		Id:     id,
		Method: method,
		Params: params,
	}

	if err := k.sendMsg(msg); err != nil {
		k.mu.Lock()
		delete(k.pendingCalls, id)
		k.mu.Unlock()
		return nil, fmt.Errorf("send call: %w", err)
	}

	result := <-resChan
	return result.result, result.err
}

func (k *KittenIPC) raiseErr(err error) {
	select {
	case k.errCh <- err:
	default:
	}
}

func (k *KittenIPC) closeSock() error {
	if err := k.listener.Close(); err != nil {
		return fmt.Errorf("close socket listener: %w", err)
	}
	return nil
}

func (k *KittenIPC) Wait() error {
	if k.mode == modeParent {
		return k.waitParent()
	}
	return k.waitChild()
}

func (k *KittenIPC) waitParent() error {
	waitErrCh := make(chan error, 1)

	go func() {
		waitErrCh <- k.cmd.Wait()
	}()

	select {
	case err := <-k.errCh:
		runtimeErr := fmt.Errorf("runtime error: %w", err)
		killErr := k.cmd.Process.Kill()
		return mergeErr(runtimeErr, killErr)
	case err := <-waitErrCh:
		if err != nil {
			return fmt.Errorf("cmd wait: %w", err)
		}
	}

	return nil
}

func (k *KittenIPC) waitChild() error {
	err := <-k.errCh
	if err != nil {
		return fmt.Errorf("ipc error: %w", err)
	}
	return nil
}

func mergeErr(errs ...error) (ret error) {
	for _, err := range errs {
		if err != nil {
			if ret == nil {
				ret = err
			} else {
				ret = fmt.Errorf("%w; %w", ret, err)
			}
		}
	}
	return
}
