package golang

const ipcSocketArg = "--ipc-socket"
const maxMessageLength = 1 << 30 // 1 GB
const defaultAcceptTimeout = 10 // seconds

type MsgType int

type Vals []any

const (
	MsgCall     MsgType = 1
	MsgResponse MsgType = 2
)

type Message struct {
	Type   MsgType `json:"type"`
	Id     int64   `json:"id"`
	Method string  `json:"method"`
	Args   Vals    `json:"args"`
	Result Vals    `json:"result"`
	Error  string  `json:"error"`
}
