package golang

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"
)

type ParentIPC struct {
	*ipcCommon
	cmd      *exec.Cmd
	listener net.Listener
	cmdDone  chan struct{}
	cmdErr   error
}

func NewParent(cmd *exec.Cmd, localApis ...any) (*ParentIPC, error) {
	return NewParentWithContext(context.Background(), cmd, localApis...)
}

func NewParentWithContext(ctx context.Context, cmd *exec.Cmd, localApis ...any) (*ParentIPC, error) {
	p := ParentIPC{
		ipcCommon: &ipcCommon{
			localApis:    mapTypeNames(localApis),
			pendingCalls: make(map[int64]*pendingCall),
			errCh:        make(chan error, 1),
			socketPath:   filepath.Join(os.TempDir(), fmt.Sprintf("kitten-ipc-%d-%d.sock", os.Getpid(), rand.Int63())),
			ctx:          ctx,
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
	p.cmdDone = make(chan struct{})

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

	go func() {
		p.cmdErr = p.cmd.Wait()
		close(p.cmdDone)
	}()

	return p.acceptConn()
}

type connResult struct {
	conn net.Conn
	err  error
}

func (p *ParentIPC) acceptConn() error {
	res := make(chan connResult, 1)
	go func() {
		conn, err := p.listener.Accept()
		res <- connResult{conn: conn, err: err}
		close(res)
	}()

	select {
	case <-time.After(time.Duration(defaultAcceptTimeout) * time.Second):
		_ = p.cmd.Process.Kill()
		return fmt.Errorf("accept timeout")
	case <-p.cmdDone:
		return fmt.Errorf("cmd exited before accepting connection: %w", p.cmdErr)
	case r := <-res:
		if r.err != nil {
			_ = p.cmd.Process.Kill()
			return fmt.Errorf("accept: %w", r.err)
		}
		p.conn = r.conn
		go p.readConn()
	}
	return nil
}

func (p *ParentIPC) Stop() error {
	p.mu.Lock()
	hasPending := len(p.pendingCalls) > 0
	p.mu.Unlock()
	if hasPending {
		return fmt.Errorf("there are calls pending")
	}
	if p.processingIncomingCalls.Load() > 0 {
		return fmt.Errorf("there are calls processing")
	}
	p.stopRequested.Store(true)
	if err := p.cmd.Process.Signal(syscall.SIGINT); err != nil {
		return fmt.Errorf("send SIGINT: %w", err)
	}
	return p.Wait()
}

func (p *ParentIPC) Wait(timeout ...time.Duration) (retErr error) {
	const maxDuration = time.Duration(1<<63 - 1)
	_timeout := maxDuration
	if len(timeout) > 0 {
		_timeout = timeout[0]
	}

loop:
	for {
		select {
		case err := <-p.errCh:
			retErr = mergeErr(retErr, fmt.Errorf("ipc internal error: %w", err))
			break loop
		case <-p.cmdDone:
			err := p.cmdErr
			if err != nil {
				var exitErr *exec.ExitError
				if ok := errors.As(err, &exitErr); ok {
					if !exitErr.Success() {
						ws, ok := exitErr.Sys().(syscall.WaitStatus)
						if !(ok && ws.Signaled() && ws.Signal() == syscall.SIGINT && p.stopRequested.Load()) {
							retErr = mergeErr(retErr, fmt.Errorf("cmd wait: %w", err))
						}
					}
				} else {
					retErr = mergeErr(retErr, fmt.Errorf("cmd wait: %w", err))
				}
			}
			break loop
		case <-time.After(_timeout):
			p.stopRequested.Store(true)
			if err := p.cmd.Process.Signal(syscall.SIGINT); err != nil {
				retErr = mergeErr(retErr, fmt.Errorf("send SIGINT: %w", err))
			}
		}
	}

	p.closeConn()
	_ = os.Remove(p.socketPath)

	return retErr
}
