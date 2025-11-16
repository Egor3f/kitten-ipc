package golang

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/samber/mo"
)

const ipcSocketArg = "--ipc-socket"
const maxMessageLength = 1 * 1024 * 1024 * 1024 // 1 gigabyte

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
	Args   Vals    `json:"args"`
	Result Vals    `json:"result"`
	Error  string  `json:"error"`
}

type Callable interface {
	Call(method string, params ...any) (Vals, error)
}

type ipcCommon struct {
	localApis       map[string]any
	socketPath      string
	conn            net.Conn
	errCh           chan error
	nextId          int64
	pendingCalls    map[int64]chan mo.Result[Vals]
	processingCalls atomic.Int64
	stopRequested   atomic.Bool
	mu              sync.Mutex
}

func (ipc *ipcCommon) readConn() {
	scn := bufio.NewScanner(ipc.conn)
	scn.Buffer(nil, maxMessageLength)
	for scn.Scan() {
		var msg Message
		a := scn.Bytes()
		_ = a
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
	if ipc.stopRequested.Load() {
		return
	}

	ipc.processingCalls.Add(1)
	defer ipc.processingCalls.Add(-1)

	method, err := ipc.findMethod(msg.Method)
	if err != nil {
		ipc.sendResponse(msg.Id, nil, fmt.Errorf("find method: %w", err))
		return
	}

	argsCount := method.Type().NumIn()
	if len(msg.Args) != argsCount {
		ipc.sendResponse(msg.Id, nil, fmt.Errorf("args count mismatch: expected %d, got %d", argsCount, len(msg.Args)))
		return
	}

	var args []reflect.Value
	for i, arg := range msg.Args {
		paramType := method.Type().In(i)
		argType := reflect.TypeOf(arg)
		arg = ipc.convType(paramType, argType, arg)
		args = append(args, reflect.ValueOf(arg))
	}

	allResultVals := method.Call(args)
	retResultVals := allResultVals[0 : len(allResultVals)-1]
	errResultVals := allResultVals[len(allResultVals)-1]

	var results []any
	for _, resVal := range retResultVals {
		results = append(results, resVal.Interface())
	}

	var resErr error
	if !errResultVals.IsNil() {
		resErr = errResultVals.Interface().(error)
	}

	ipc.sendResponse(msg.Id, results, resErr)
}

func (ipc *ipcCommon) convType(needType reflect.Type, gotType reflect.Type, arg any) any {
	// JSON decodes any number to float64. If we need int, we should check and convert
	if needType.Kind() == reflect.Int && gotType.Kind() == reflect.Float64 {
		floatArg := arg.(float64)
		if float64(int64(floatArg)) == floatArg && !needType.OverflowInt(int64(floatArg)) {
			arg = int(floatArg)
		}
	}
	return arg
}

func (ipc *ipcCommon) findMethod(methodName string) (reflect.Value, error) {
	parts := strings.Split(methodName, ".")
	if len(parts) != 2 {
		return reflect.Value{}, fmt.Errorf("invalid method: %s", methodName)
	}

	endpointName, methodName := parts[0], parts[1]

	localApi, ok := ipc.localApis[endpointName]
	if !ok {
		return reflect.Value{}, fmt.Errorf("endpoint not found: %s", endpointName)
	}

	method := reflect.ValueOf(localApi).MethodByName(methodName)
	if !method.IsValid() {
		return reflect.Value{}, fmt.Errorf("method not found: %s", methodName)
	}

	return method, nil
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
	if ipc.stopRequested.Load() {
		return nil, fmt.Errorf("ipc is stopping")
	}

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
		Args:   params,
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

func (ipc *ipcCommon) closeConn() {
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

func NewParent(cmd *exec.Cmd, localApis ...any) (*ParentIPC, error) {
	p := ParentIPC{
		ipcCommon: &ipcCommon{
			localApis:    mapTypeNames(localApis),
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

func (p *ParentIPC) Stop() error {
	if len(p.pendingCalls) > 0 {
		return fmt.Errorf("there are calls pending")
	}
	if p.processingCalls.Load() > 0 {
		return fmt.Errorf("there are calls processing")
	}
	p.stopRequested.Store(true)
	if err := p.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return fmt.Errorf("send SIGTERM: %w", err)
	}
	return p.Wait()
}

func (p *ParentIPC) Wait() error {
	waitErrCh := make(chan error, 1)

	go func() {
		waitErrCh <- p.cmd.Wait()
	}()

	var retErr error
	select {
	case err := <-p.errCh:
		retErr = fmt.Errorf("ipc internal error: %w", err)
	case err := <-waitErrCh:
		if err != nil {
			var exitErr *exec.ExitError
			if ok := errors.As(err, &exitErr); ok {
				if !exitErr.Success() {
					ws, ok := exitErr.Sys().(syscall.WaitStatus)
					if !(ok && ws.Signaled() && ws.Signal() == syscall.SIGINT && p.stopRequested.Load()) {
						retErr = fmt.Errorf("cmd wait: %w", err)
					}
				}
			} else {
				retErr = fmt.Errorf("cmd wait: %w", err)
			}
		}
	}

	p.closeConn()

	return retErr
}

type ChildIPC struct {
	*ipcCommon
}

func NewChild(localApis ...any) (*ChildIPC, error) {
	c := ChildIPC{
		ipcCommon: &ipcCommon{
			localApis:    mapTypeNames(localApis),
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

func mapTypeNames(types []any) map[string]any {
	result := make(map[string]any)
	for _, t := range types {
		typeName := reflect.TypeOf(t).Elem().Name()
		result[typeName] = t
	}
	return result
}
