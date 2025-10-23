package kittenipc

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/samber/mo"
)

type StdioMode int

type Config struct {
}

type KittenIPC struct {
	cmd *exec.Cmd
	cfg Config

	socketPath string
	listener   net.Listener
	conn       net.Conn
	errCh      chan error
}

func New(cmd *exec.Cmd, api any, cfg Config) (*KittenIPC, error) {
	k := KittenIPC{
		cmd: cmd,
		cfg: cfg,
	}

	k.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("kitten-ipc-%d.sock", os.Getpid()))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	const ipcSocketArg = "--ipc-socket"
	if slices.Contains(cmd.Args, ipcSocketArg) {
		return nil, fmt.Errorf("you should not use `%s` argument in your command", ipcSocketArg)
	}
	cmd.Args = append(cmd.Args, ipcSocketArg, k.socketPath)

	k.errCh = make(chan error, 1)

	return &k, nil
}

func (k *KittenIPC) Start() error {
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

}

func (k *KittenIPC) Call() {

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
