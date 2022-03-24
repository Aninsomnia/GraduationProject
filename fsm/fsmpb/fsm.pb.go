package fsmpb

type MessageType int32

const (
	MsgBeat MessageType = 0
)

type Message struct {
	Type MessageType
	To   uint64
	From uint64
}
