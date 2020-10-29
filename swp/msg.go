package swp

type MsgType uint8

const (
	CONNECT MsgType = iota
	DISCONNECT
	GETCHUNK
	ACK
	CHUNKHASH
	PIECE
)

type Msg interface {
	Type() MsgType
}

type Connect struct {
	Id string
}

func (r Connect) Type() MsgType {
	return CONNECT
}

type Disconnect struct {
}

func (r Disconnect) Type() MsgType {
	return DISCONNECT
}

type Getchunk struct {
	Index uint64
}

func (r Getchunk) Type() MsgType {
	return GETCHUNK
}

type Ack struct {
}

func (r Ack) Type() MsgType {
	return ACK
}

type Chunkhash struct {
	NPieces uint64
	Hash    []byte
}

func (r Chunkhash) Type() MsgType {
	return CHUNKHASH
}

type Piece struct {
	Piece []byte
}

func (r Piece) Type() MsgType {
	return PIECE
}
