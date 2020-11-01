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

func (t MsgType) String() string {
	switch t {
	case CONNECT:
		return "CONNECT"
	case DISCONNECT:
		return "DISCONNECT"
	case GETCHUNK:
		return "GETCHUNK"
	case ACK:
		return "ACK"
	case CHUNKHASH:
		return "CHUNKHASH"
	case PIECE:
		return "PIECE"
	default:
		return "UNKNOWN"
	}
}

const headerSize = 1

type Msg interface {
	Type() MsgType
	Size() int
}

type Connect struct {
	Id string
}

func (r Connect) Type() MsgType {
	return CONNECT
}

func (r Connect) Size() int {
	return headerSize + len([]byte(r.Id))
}

type Disconnect struct {
}

func (r Disconnect) Type() MsgType {
	return DISCONNECT
}

func (r Disconnect) Size() int {
	return headerSize
}

type Getchunk struct {
	Index uint64
}

func (r Getchunk) Type() MsgType {
	return GETCHUNK
}

func (r Getchunk) Size() int {
	return headerSize + 8
}

type Ack struct {
}

func (r Ack) Type() MsgType {
	return ACK
}

func (r Ack) Size() int {
	return headerSize
}

type Chunkhash struct {
	NPieces uint64
	Hash    []byte
}

func (r Chunkhash) Type() MsgType {
	return CHUNKHASH
}

func (r Chunkhash) Size() int {
	return headerSize + 8 + len(r.Hash)
}

type Piece struct {
	Piece []byte
}

func (r Piece) Type() MsgType {
	return PIECE
}

func (r Piece) Size() int {
	return headerSize + len(r.Piece)
}
