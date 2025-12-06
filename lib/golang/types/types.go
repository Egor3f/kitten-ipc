package types

type ValType string

const (
	TNoType ValType = "" // zero value constant for ValType (please don't use just "" as zero value!)
	TInt    ValType = "int"
	TString ValType = "string"
	TBool   ValType = "bool"
	TBlob   ValType = "blob"
	TArray  ValType = "array"
)
