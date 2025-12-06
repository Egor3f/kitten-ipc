package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"time"

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

func (api GoIpcApi) XorData(data1 []byte, data2 []byte) ([]byte, error) {
	if len(data1) == 0 || len(data2) == 0 {
		return nil, fmt.Errorf("empty input data")
	}
	if len(data1) != len(data2) {
		return nil, fmt.Errorf("input data length mismatch")
	}
	result := make([]byte, len(data1))
	for i := 0; i < len(data1); i++ {
		result[i] = data1[i] ^ data2[i]
	}
	return result, nil
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	localApi := GoIpcApi{}

	cmd := exec.Command("node", path.Join(cwd, "ts/dist/index.js"))

	ipc, err := kittenipc.NewParent(cmd, &localApi)
	if err != nil {
		log.Panic(err)
	}

	if err := ipc.Start(); err != nil {
		log.Panic(err)
	}

	remoteApi := TsIpcApi{Ipc: ipc}
	resDiv, err := remoteApi.Div(10, 2)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("call result go->ts Div = %v", resDiv)

	data1 := slices.Repeat([]byte{0b10101010}, 10)
	data2 := slices.Repeat([]byte{0b11110000}, 10)

	resXor, err := remoteApi.XorData(data1, data2)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("call result go->ts XorData = %v", resXor)

	if err := ipc.Wait(1 * time.Second); err != nil {
		log.Panic(err)
	}
}
