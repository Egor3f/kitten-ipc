package kittenipc

import (
	"bufio"
	"encoding/json"
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

	"github.com/samber/mo"
)

const ipcSocketArg = "--ipc-socket"

type StdioMode int

type MsgType int

type Vals []any

const (
	MsgCall     MsgType = 1
	MsgResponse MsgType = 2
)

type Message struct {
	Type   MsgType `json:"type"`
	Id     int64   `json:"id"`
	Method string  `json:"method"`
	Params Vals    `json:"params"`
	Result Vals    `json:"result"`
	Error  string  `json:"error"`
}

type Callable interface {
	Call(method string, params ...any) (Vals, error)
}

type ipcCommon struct {
	localApi     any
	socketPath   string
	conn         net.Conn
	errCh        chan error
	nextId       int64
	pendingCalls map[int64]chan mo.Result[Vals]
	mu           sync.Mutex
}

func (ipc *ipcCommon) readConn() {
	scn := bufio.NewScanner(ipc.conn)
	for scn.Scan() {
		var msg Message
		if err := json.Unmarshal(scn.Bytes(), &msg); err != nil {
			ipc.raiseErr(fmt.Errorf("unmarshal message: %w", err))
			break
		}
		ipc.processMsg(msg)
	}
	if err := scn.Err(); err != nil {
		ipc.raiseErr(err)
	}
}

func (ipc *ipcCommon) processMsg(msg Message) {
	switch msg.Type {
	case MsgCall:
		ipc.handleCall(msg)
	case MsgResponse:
		ipc.handleResponse(msg)
	}
}

func (ipc *ipcCommon) sendMsg(msg Message) error {

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	data = append(data, '\n')

	if _, err := ipc.conn.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (ipc *ipcCommon) handleCall(msg Message) {

	if ipc.localApi == nil {
		ipc.sendResponse(msg.Id, nil, fmt.Errorf("remote side does not accept ipc calls"))
	}
	localApi := reflect.ValueOf(ipc.localApi)

	method := localApi.MethodByName(msg.Method)
	if !method.IsValid() {
		ipc.sendResponse(msg.Id, nil, fmt.Errorf("method not found: %s", msg.Method))
		return
	}

	methodType := method.Type()
	argsCount := methodType.NumIn()

	if len(msg.Params) != argsCount {
		ipc.sendResponse(msg.Id, nil, fmt.Errorf("argument count mismatch: expected %d, got %d", argsCount, len(msg.Params)))
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

	ipc.sendResponse(msg.Id, res, resErr.Interface().(error))
}

func (ipc *ipcCommon) sendResponse(id int64, result []any, err error) {
	msg := Message{
		Type:   MsgResponse,
		Id:     id,
		Result: result,
	}

	if err != nil {
		msg.Error = err.Error()
	}

	if err := ipc.sendMsg(msg); err != nil {
		ipc.raiseErr(fmt.Errorf("send response for id=%d: %w", id, err))
	}
}

func (ipc *ipcCommon) handleResponse(msg Message) {
	ipc.mu.Lock()
	ch, ok := ipc.pendingCalls[msg.Id]
	if ok {
		delete(ipc.pendingCalls, msg.Id)
	}
	ipc.mu.Unlock()

	if !ok {
		ipc.raiseErr(fmt.Errorf("received response for unknown call id: %d", msg.Id))
		return
	}

	var res mo.Result[Vals]
	if msg.Error == "" {
		res = mo.Ok[Vals](msg.Result)
	} else {
		res = mo.Err[Vals](fmt.Errorf("remote error: %s", msg.Error))
	}
	ch <- res
	close(ch)
}

func (ipc *ipcCommon) Call(method string, params ...any) (Vals, error) {
	ipc.mu.Lock()
	id := ipc.nextId
	ipc.nextId++
	resChan := make(chan mo.Result[Vals], 1)
	ipc.pendingCalls[id] = resChan
	ipc.mu.Unlock()

	msg := Message{
		Type:   MsgCall,
		Id:     id,
		Method: method,
		Params: params,
	}

	if err := ipc.sendMsg(msg); err != nil {
		ipc.mu.Lock()
		delete(ipc.pendingCalls, id)
		ipc.mu.Unlock()
		return nil, fmt.Errorf("send call: %w", err)
	}

	result := <-resChan
	return result.Get()
}

func (ipc *ipcCommon) raiseErr(err error) {
	select {
	case ipc.errCh <- err:
	default:
	}
}

func (ipc *ipcCommon) cleanup() {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()
	_ = ipc.conn.Close()
	for _, call := range ipc.pendingCalls {
		call <- mo.Err[Vals](fmt.Errorf("call cancelled due to ipc termination"))
	}
}

type ParentIPC struct {
	*ipcCommon
	cmd      *exec.Cmd
	listener net.Listener
}

func NewParent(cmd *exec.Cmd, localApi any) (*ParentIPC, error) {
	p := ParentIPC{
		ipcCommon: &ipcCommon{
			localApi:     localApi,
			pendingCalls: make(map[int64]chan mo.Result[Vals]),
			errCh:        make(chan error, 1),
			socketPath:   filepath.Join(os.TempDir(), fmt.Sprintf("kitten-ipc-%d.sock", os.Getpid())),
		},
		cmd: cmd,
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if slices.Contains(cmd.Args, ipcSocketArg) {
		return nil, fmt.Errorf("you should not use `%s` argument in your command", ipcSocketArg)
	}
	cmd.Args = append(cmd.Args, ipcSocketArg, p.socketPath)

	p.errCh = make(chan error, 1)

	return &p, nil
}

func (p *ParentIPC) Start() error {
	_ = os.Remove(p.socketPath)
	listener, err := net.Listen("unix", p.socketPath)
	if err != nil {
		return fmt.Errorf("listen unix socket: %w", err)
	}
	p.listener = listener
	defer p.listener.Close()

	err = p.cmd.Start()
	if err != nil {
		return fmt.Errorf("cmd start: %w", err)
	}

	return p.acceptConn()
}

func (p *ParentIPC) acceptConn() error {
	const acceptTimeout = time.Second * 10

	res := make(chan mo.Result[net.Conn], 1)
	go func() {
		conn, err := p.listener.Accept()
		if err != nil {
			res <- mo.Err[net.Conn](err)
		} else {
			res <- mo.Ok[net.Conn](conn)
		}
		close(res)
	}()

	select {
	case <-time.After(acceptTimeout):
		_ = p.cmd.Process.Kill()
		return fmt.Errorf("accept timeout")
	case res := <-res:
		if res.IsError() {
			_ = p.cmd.Process.Kill()
			return fmt.Errorf("accept: %w", res.Error())
		}
		p.conn = res.MustGet()
		go p.readConn()
	}
	return nil
}

func (p *ParentIPC) Wait() (retErr error) {
	waitErrCh := make(chan error, 1)

	go func() {
		waitErrCh <- p.cmd.Wait()
	}()

	select {
	case err := <-p.errCh:
		retErr = fmt.Errorf("ipc internal error: %w", err)
	case err := <-waitErrCh:
		if err != nil {
			retErr = fmt.Errorf("cmd wait: %w", err)
		}
	}

	p.cleanup()

	return
}

type ChildIPC struct {
	*ipcCommon
}

func NewChild(localApi any) (*ChildIPC, error) {
	c := ChildIPC{
		ipcCommon: &ipcCommon{
			localApi:     localApi,
			pendingCalls: make(map[int64]chan mo.Result[Vals]),
			errCh:        make(chan error, 1),
		},
	}

	socketPath := flag.String("ipc-socket", "", "Path to IPC socket")
	flag.Parse()

	if *socketPath == "" {
		return nil, fmt.Errorf("ipc socket path is missing")
	}
	c.socketPath = *socketPath

	return &c, nil
}

func (c *ChildIPC) Start() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("connect to parent socket: %w", err)
	}
	c.conn = conn
	c.readConn()
	return nil
}

func (c *ChildIPC) Wait() error {
	err := <-c.errCh
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
