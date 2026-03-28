package golang

import (
	"context"
	"fmt"
	"net"
	"os"
)

type ChildIPC struct {
	*ipcCommon
}

func NewChild(localApis ...any) (*ChildIPC, error) {
	c := ChildIPC{
		ipcCommon: &ipcCommon{
			localApis:    mapTypeNames(localApis),
			pendingCalls: make(map[int64]*pendingCall),
			errCh:        make(chan error, 1),
			ctx:          context.Background(),
		},
	}

	socketPath := socketPathFromArgs()
	if socketPath == "" {
		return nil, fmt.Errorf("ipc socket path is missing")
	}
	c.socketPath = socketPath

	return &c, nil
}

func (c *ChildIPC) Start() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("connect to parent socket: %w", err)
	}
	c.conn = conn
	go c.readConn()
	return nil
}

func (c *ChildIPC) Wait() error {
	err := <-c.errCh
	if err != nil {
		return fmt.Errorf("ipc error: %w", err)
	}
	return nil
}

// socketPathFromArgs parses --ipc-socket from os.Args without calling flag.Parse(),
// which would interfere with the host application's flag handling.
func socketPathFromArgs() string {
	for i, arg := range os.Args {
		if arg == ipcSocketArg && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}
