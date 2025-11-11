package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	kittenipc "efprojects.com/kitten-ipc"
)

// kittenipc:api
type GoIpcApi struct {
}

func (api GoIpcApi) Div(a int, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("zero division")
	}
	return a / b, nil
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	localApi := GoIpcApi{}

	cmdStr := fmt.Sprintf("node %s", path.Join(cwd, "..", "ts/index.js"))
	cmd := exec.Command(cmdStr)

	ipc, err := kittenipc.NewParent(cmd, &localApi)
	if err != nil {
		log.Panic(err)
	}

	if err := ipc.Start(); err != nil {
		log.Panic(err)
	}

	remoteApi := TsIpcApi{Ipc: ipc}
	res, err := remoteApi.Div(10, 2)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("remote call result = %v", res)

	if err := ipc.Wait(); err != nil {
		log.Panic(err)
	}
}
