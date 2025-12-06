package api

import "efprojects.com/kitten-ipc/types"

// todo check TInt size < 64
// todo check not float

type Val struct {
	Name     string
	Type     types.ValType
	Children []Val
}

type Method struct {
	Name   string
	Params []Val
	Ret    []Val
}

type Endpoint struct {
	Name    string
	Methods []Method
}

type Api struct {
	Endpoints []Endpoint
}
