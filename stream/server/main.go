package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/aburdulescu/go-ez/stream/rpc"
	"google.golang.org/protobuf/proto"
)

const PIECE_SIZE = 8192
const CHUNK_SIZE = 1024 * PIECE_SIZE

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	ln, err := net.Listen("tcp", ":28080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	remAddr := conn.RemoteAddr().String()
	b := make([]byte, 8192)
	for {
		n, err := conn.Read(b)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			return
		}
		npieces := uint64(1024)
		rsp := &rpc.Response{Npieces: npieces}
		rspBuf, err := proto.Marshal(rsp)
		if err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			return
		}
		if _, err := conn.Write(rspBuf); err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			return
		}
		req := &rpc.Request{}
		if err := proto.Unmarshal(b[:n], req); err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			return
		}
		if err := handleRequest(conn, req); err != nil {
			log.Printf("%s: error: %v\n", remAddr, err)
			return
		}
	}
}

func handleRequest(conn net.Conn, req *rpc.Request) error {
	chunk, err := readChunk("f100MB", req.GetIndex()-1)
	if err != nil {
		return err
	}
	npieces := len(chunk) / PIECE_SIZE
	for i := 0; i < npieces; i++ {
		piece := chunk[i*PIECE_SIZE : (i+1)*PIECE_SIZE]
		rsp := &rpc.Piece{Piece: piece}
		if err := sendPiece(conn, rsp); err != nil {
			return err
		}
	}
	if len(chunk)%PIECE_SIZE != 0 {
		piece := chunk[(len(chunk)-1)*PIECE_SIZE:]
		rsp := &rpc.Piece{Piece: piece}
		if err := sendPiece(conn, rsp); err != nil {
			return err
		}
	}
	rsp := &rpc.Piece{Piece: nil}
	if err := sendPiece(conn, rsp); err != nil {
		return err
	}
	return nil
}

func readChunk(path string, i uint64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := io.NewSectionReader(f, int64(CHUNK_SIZE*i), CHUNK_SIZE)
	b := make([]byte, CHUNK_SIZE)
	n, err := r.Read(b)
	if err != io.EOF && err != nil {
		return nil, err
	}
	return b[:n], nil
}

func sendPiece(conn net.Conn, rsp *rpc.Piece) error {
	b, err := proto.Marshal(rsp)
	if err != nil {
		return err
	}
	if _, err := conn.Write(b); err != nil {
		return err
	}
	return nil
}
