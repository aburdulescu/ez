package main

import (
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezs"
	"google.golang.org/protobuf/proto"
)

var chunkPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, chunks.CHUNK_SIZE)
	},
}

type SeederServer struct {
	ln net.Listener
	db DB
}

type SeederServerReqHandler struct {
	id   string
	conn net.Conn
	db   DB
}

func NewSeederServer(db DB) (SeederServer, error) {
	ln, err := net.Listen("tcp", ":22201")
	if err != nil {
		return SeederServer{}, err
	}
	s := SeederServer{
		ln: ln, db: db,
	}
	return s, nil
}

func (s SeederServer) ListenAndServe() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		h := SeederServerReqHandler{
			db:   s.db,
			conn: conn,
		}
		go h.run()
	}
}

func (h SeederServerReqHandler) run() {
	defer h.conn.Close()
	remAddr := h.conn.RemoteAddr().String()
	for {
		req, err := h.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			continue
		}
		reqType := req.GetType()
		switch reqType {
		case ezs.RequestType_CONNECT:
			h.id = req.GetId()
			if err := h.handleConnect(); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_DISCONNECT:
			h.id = ""
			if err := h.handleDisconnect(); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_GETCHUNK:
			if err := h.handleGetchunk(req.GetIndex()); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_GETPIECE:
			if err := h.handleGetpiece(req.GetIndex()); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		default:
			log.Printf("%s: error: unknown request type %v\n", remAddr, reqType)
		}
	}
}

func (h SeederServerReqHandler) Send(rsp ezs.Response) error {
	b, err := proto.Marshal(&rsp)
	if err != nil {
		log.Println(err)
		return err
	}
	if err := ezs.Write(h.conn, b); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (h SeederServerReqHandler) Recv() (*ezs.Request, error) {
	b, err := ezs.Read(h.conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req := &ezs.Request{}
	if err := proto.Unmarshal(b, req); err != nil {
		log.Println(err)
		return nil, err
	}
	ezs.ReleaseMsg(b)
	return req, nil
}

func (h SeederServerReqHandler) handleConnect() error {
	rsp := ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	if err := h.Send(rsp); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (h SeederServerReqHandler) handleDisconnect() error {
	rsp := ezs.Response{
		Type:    ezs.ResponseType_ACK,
		Payload: &ezs.Response_Dummy{},
	}
	if err := h.Send(rsp); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (h SeederServerReqHandler) handleGetchunk(index uint64) error {
	chunkHashes, err := h.db.GetChunkHashes(string(h.id))
	if err != nil {
		log.Println(err)
		return err
	}
	ifile, err := h.db.GetIFile(string(h.id))
	if err != nil {
		log.Println(err)
		return err
	}
	chunkBuf, n, err := readChunk(filepath.Join(ifile.Dir, ifile.Name), index)
	if err != nil {
		log.Println(err)
		return err
	}
	defer chunkPool.Put(chunkBuf)
	chunk := chunkBuf[:n]
	npieces := uint64(len(chunk) / chunks.PIECE_SIZE)
	remainder := uint64(0)
	if len(chunk)%chunks.PIECE_SIZE != 0 {
		npieces++
		remainder = 1
	}
	rsp := ezs.Response{
		Type: ezs.ResponseType_CHUNKHASH,
		Payload: &ezs.Response_Chunkhash{
			&ezs.Chunkhash{
				Hash:    []byte(chunkHashes[index]),
				Npieces: npieces,
			},
		},
	}
	if err := h.Send(rsp); err != nil {
		log.Println(err)
		return err
	}
	for i := uint64(0); i < npieces-remainder; i++ {
		piece := chunk[i*chunks.PIECE_SIZE : (i+1)*chunks.PIECE_SIZE]
		rsp := ezs.Piece{Piece: piece}
		b, err := proto.Marshal(&rsp)
		if err != nil {
			log.Println(err)
			return err
		}
		if err := ezs.Write(h.conn, b); err != nil {
			log.Println(err)
			return err
		}
	}
	if remainder != 0 {
		piece := chunk[(len(chunk)-1)*chunks.PIECE_SIZE:]
		rsp := &ezs.Piece{Piece: piece}
		b, err := proto.Marshal(rsp)
		if err != nil {
			log.Println(err)
			return err
		}
		if err := ezs.Write(h.conn, b); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func readChunk(path string, i uint64) ([]byte, int, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}
	defer f.Close()
	r := io.NewSectionReader(f, int64(chunks.CHUNK_SIZE*i), chunks.CHUNK_SIZE)
	b := chunkPool.Get().([]byte)
	n, err := r.Read(b)
	if err != io.EOF && err != nil {
		log.Println(n, err)
		return nil, 0, err
	}
	return b, n, nil
}

func (h SeederServerReqHandler) handleGetpiece(index uint64) error {
	rsp := ezs.Response{
		Type:    ezs.ResponseType_PIECE,
		Payload: &ezs.Response_Piece{[]byte("piece")},
	}
	if err := h.Send(rsp); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
