package types

type ValType int

const (
	TInt    ValType = 1
	TString ValType = 2
	TBool   ValType = 3
	TBlob   ValType = 4
	TArray  ValType = 5
)

func (v ValType) String() string {
	switch v {
	case TInt:
		return "int"
	case TString:
		return "string"
	case TBool:
		return "bool"
	case TBlob:
		return "blob"
	case TArray:
		return "array"
	default:
		panic("unreachable code")
	}
}
