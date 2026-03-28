package api

type ValType string

const (
	TNoType ValType = ""
	TInt    ValType = "int"
	TString ValType = "string"
	TBool   ValType = "bool"
	TBlob   ValType = "blob"
)

type Val struct {
	Name string
	Type ValType
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
