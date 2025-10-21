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
type IpcApi struct {
}

func (api IpcApi) Div(a int, b int) (int, error) {
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

	api := IpcApi{}

	cmdStr := fmt.Sprintf("node %s", path.Join(cwd, "..", "ts/index.js"))
	cmd := exec.Command(cmdStr)

	kit, err := kittenipc.New(cmd, &api, kittenipc.Config{})
	if err != nil {
		log.Panic(err)
	}

	if err := kit.Start(); err != nil {
		log.Panic(err)
	}

	if err := kit.Wait(); err != nil {
		log.Panic(err)
	}
}
