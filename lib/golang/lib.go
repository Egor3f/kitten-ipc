package kittenipc

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

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
	}

	return nil
}

func (k *KittenIPC) closeSock() error {
	if err := k.listener.Close(); err != nil {
		return fmt.Errorf("close socket listener: %w", err)
	}
	return nil
}

func (k *KittenIPC) Wait() error {
	if err := k.cmd.Wait(); err != nil {
		return fmt.Errorf("cmd wait: %w", err)
	}

	if err := k.closeSock(); err != nil {
		return fmt.Errorf("closeSock: %w", err)
	}

	return nil
}
