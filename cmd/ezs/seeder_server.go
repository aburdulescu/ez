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
	"github.com/aburdulescu/ez/ezt"
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
	id    string
	conn  net.Conn
	db    DB
	f     *os.File
	ifile *ezt.IFile
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
			if err := h.handleConnect(req.GetId()); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_DISCONNECT:
			if err := h.handleDisconnect(); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case ezs.RequestType_GETCHUNK:
			if err := h.handleGetchunk(req.GetIndex()); err != nil {
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

func (h *SeederServerReqHandler) handleConnect(id string) error {
	h.id = id
	ifile, err := h.db.GetIFile(h.id)
	if err != nil {
		log.Println(err)
		return err
	}
	h.ifile = &ifile
	f, err := os.Open(filepath.Join(ifile.Dir, ifile.Name))
	if err != nil {
		log.Println(err)
		return err
	}
	h.f = f
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

func (h *SeederServerReqHandler) handleDisconnect() error {
	h.id = ""
	h.ifile = nil
	h.f.Close()
	h.f = nil
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
	chunkBuf, n, err := readChunk(h.f, index)
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
	i := uint64(0)
	for ; i < npieces-remainder; i++ {
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
		log.Println(i * chunks.PIECE_SIZE)
		piece := chunk[i*chunks.PIECE_SIZE:]
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

func readChunk(f *os.File, i uint64) ([]byte, int, error) {
	r := io.NewSectionReader(f, int64(chunks.CHUNK_SIZE*i), chunks.CHUNK_SIZE)
	b := chunkPool.Get().([]byte)
	n, err := r.Read(b)
	if err != io.EOF && err != nil {
		log.Println(n, err)
		return nil, 0, err
	}
	return b, n, nil
}
