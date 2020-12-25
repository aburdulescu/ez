package main

import (
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/aburdulescu/ez/cmn"
	"github.com/aburdulescu/ez/ezt"
	"github.com/aburdulescu/ez/swp"
)

var chunkPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, cmn.ChunkSize)
	},
}

type SeederServer struct {
	ln net.Listener
	db *DB
}

type SeederServerReqHandler struct {
	id    string
	conn  net.Conn
	db    *DB
	f     *os.File
	ifile *ezt.IFile
}

func NewSeederServer(db *DB) (SeederServer, error) {
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
		msg, cleanup, err := swp.Recv(h.conn)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			continue
		}
		msgType := msg.Type()
		switch msgType {
		case swp.CONNECT:
			req := msg.(swp.Connect)
			log.Printf("%s: CONNECT %v\n", remAddr, req.Id)
			if err := h.handleConnect(req.Id); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case swp.DISCONNECT:
			log.Printf("%s: DISCONNECT\n", remAddr)
			if err := h.handleDisconnect(); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		case swp.GETCHUNK:
			req := msg.(swp.Getchunk)
			log.Printf("%s: GETCHUNK %v\n", remAddr, req.Index)
			if err := h.handleGetchunk(req.Index); err != nil {
				log.Printf("%s: error: %v\n", remAddr, err)
			}
		default:
			log.Printf("%s: error: unknown request type %v\n", remAddr, msgType)
		}
		cleanup()
	}
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
	if err := swp.Send(h.conn, swp.Ack{}); err != nil {
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
	if err := swp.Send(h.conn, swp.Ack{}); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (h SeederServerReqHandler) handleGetchunk(index uint64) error {
	checksums, err := h.db.GetChecksums(h.id)
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
	npieces := uint64(len(chunk) / cmn.PieceSize)
	remainder := uint64(0)
	if len(chunk)%cmn.PieceSize != 0 {
		npieces++
		remainder = 1
	}
	rsp := swp.Chunkinfo{
		NPieces:  npieces,
		Checksum: uint64(checksums[index]),
	}
	if err := swp.Send(h.conn, rsp); err != nil {
		log.Println(err)
		return err
	}
	i := uint64(0)
	for ; i < npieces-remainder; i++ {
		piece := chunk[i*cmn.PieceSize : (i+1)*cmn.PieceSize]
		if err := swp.Send(h.conn, swp.Piece{piece}); err != nil {
			log.Println(err)
			return err
		}
	}
	if remainder != 0 {
		log.Println(i * cmn.PieceSize)
		piece := chunk[i*cmn.PieceSize:]
		if err := swp.Send(h.conn, swp.Piece{piece}); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func readChunk(f *os.File, i uint64) ([]byte, int, error) {
	r := io.NewSectionReader(f, int64(cmn.ChunkSize*i), cmn.ChunkSize)
	b := chunkPool.Get().([]byte)
	n, err := r.Read(b)
	if err != io.EOF && err != nil {
		log.Println(n, err)
		return nil, 0, err
	}
	return b, n, nil
}
