package api

// todo check TInt size < 64
// todo check not float

type ValType int

const (
	TInt    ValType = 1
	TString ValType = 2
	TBool   ValType = 3
	TBlob   ValType = 4
	TArray  ValType = 5
)

type Val struct {
	Name     string
	Type     ValType
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
