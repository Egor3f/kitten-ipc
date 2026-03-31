package golang

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

type IpcCommon interface {
	Call(method string, params ...any) (Vals, error)
	ConvType(needType, gotType reflect.Type, arg any) any
}

type callResult struct {
	vals Vals
	err  error
}

type pendingCall struct {
	resultChan chan callResult
}

type Options struct {
	DebugMessages bool
}

type ipcCommon struct {
	localApis               map[string]any
	socketPath              string
	conn                    net.Conn
	errCh                   chan error
	nextId                  int64
	pendingCalls            map[int64]*pendingCall
	processingIncomingCalls atomic.Int64
	stopRequested           atomic.Bool
	mu                      sync.Mutex
	writeMu                 sync.Mutex
	ctx                     context.Context
	debugMessages           bool
}

func (ipc *ipcCommon) readConn() {
	scn := bufio.NewScanner(ipc.conn)
	scn.Buffer(nil, maxMessageLength)
	for scn.Scan() {
		var msg Message
		msgBytes := scn.Bytes()
		if ipc.debugMessages {
			log.Printf("[ipc recv] %s", string(msgBytes))
		}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			ipc.raiseErr(fmt.Errorf("unmarshal message: %w", err))
			break
		}
		ipc.handleIncomingMsg(msg)
	}
	if err := scn.Err(); err != nil {
		ipc.raiseErr(err)
	}
}

func (ipc *ipcCommon) handleIncomingMsg(msg Message) {
	switch msg.Type {
	case MsgCall:
		go ipc.handleIncomingCall(msg)
	case MsgResponse:
		ipc.handleOutgoingResponse(msg)
	}
}

func (ipc *ipcCommon) sendMsg(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	if ipc.debugMessages {
		log.Printf("[ipc send] %s", string(data))
	}
	data = append(data, '\n')

	ipc.writeMu.Lock()
	_, writeErr := ipc.conn.Write(data)
	ipc.writeMu.Unlock()
	if writeErr != nil {
		return fmt.Errorf("write message: %w", writeErr)
	}

	return nil
}

func (ipc *ipcCommon) handleIncomingCall(msg Message) {
	if ipc.stopRequested.Load() {
		return
	}

	ipc.processingIncomingCalls.Add(1)
	defer ipc.processingIncomingCalls.Add(-1)

	defer func() {
		if err := recover(); err != nil {
			ipc.sendResponse(msg.Id, nil, fmt.Errorf("handle call panicked: %s", err))
		}
	}()

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
		arg = ipc.ConvType(paramType, argType, arg)
		args = append(args, reflect.ValueOf(arg))
	}

	var errorType = reflect.TypeOf((*error)(nil)).Elem()

	allResultVals := method.Call(args)
	var retResultVals []reflect.Value
	var errResultVal reflect.Value
	if len(allResultVals) > 0 {
		if allResultVals[len(allResultVals)-1].Type().Implements(errorType) {
			retResultVals = allResultVals[0 : len(allResultVals)-1]
			errResultVal = allResultVals[len(allResultVals)-1]
		} else {
			retResultVals = allResultVals
		}
	}

	var results []any
	for _, resVal := range retResultVals {
		results = append(results, resVal.Interface())
	}

	var resultError error
	if errResultVal.IsValid() && !errResultVal.IsNil() {
		resultError = errResultVal.Interface().(error)
	}

	ipc.sendResponse(msg.Id, results, resultError)
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

func (ipc *ipcCommon) handleOutgoingResponse(msg Message) {
	ipc.mu.Lock()
	call, ok := ipc.pendingCalls[msg.Id]
	if ok {
		delete(ipc.pendingCalls, msg.Id)
	}
	ipc.mu.Unlock()

	if !ok {
		ipc.raiseErr(fmt.Errorf("received response for unknown call id: %d", msg.Id))
		return
	}

	var res callResult
	if msg.Error == "" {
		res = callResult{vals: msg.Result}
	} else {
		res = callResult{err: fmt.Errorf("remote error: %s", msg.Error)}
	}
	call.resultChan <- res
	close(call.resultChan)
}

func (ipc *ipcCommon) Call(method string, params ...any) (Vals, error) {
	if ipc.conn == nil {
		return nil, fmt.Errorf("ipc is not connected to remote process socket")
	}

	if ipc.stopRequested.Load() {
		return nil, fmt.Errorf("ipc is stopping")
	}

	ipc.mu.Lock()
	id := ipc.nextId
	ipc.nextId++
	call := &pendingCall{
		resultChan: make(chan callResult, 1),
	}
	ipc.pendingCalls[id] = call
	ipc.mu.Unlock()

	if params == nil {
		params = make([]any, 0)
	}

	for i := range params {
		params[i] = ipc.serialize(params[i])
	}

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

	select {
	case result := <-call.resultChan:
		return result.vals, result.err
	case <-ipc.ctx.Done():
		ipc.mu.Lock()
		delete(ipc.pendingCalls, id)
		ipc.mu.Unlock()
		return nil, ipc.ctx.Err()
	}
}

func (ipc *ipcCommon) raiseErr(err error) {
	select {
	case ipc.errCh <- err:
	default:
	}
}

func (ipc *ipcCommon) closeConn() {
	_ = ipc.conn.Close()
	ipc.mu.Lock()
	pending := ipc.pendingCalls
	ipc.pendingCalls = make(map[int64]*pendingCall)
	ipc.mu.Unlock()
	for _, call := range pending {
		call.resultChan <- callResult{err: fmt.Errorf("call cancelled due to ipc termination")}
		close(call.resultChan)
	}
}
