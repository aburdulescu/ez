package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezt"
)

const MAX_CHUNKS uint64 = 8

type Downloader struct {
	id    string
	peers []string
	f     *os.File
}

type Chunk struct {
	err   error
	index uint64
	buf   *bytes.Buffer
}

func (d *Downloader) Run(id string, ifile ezt.IFile, peers []string) error {
	d.id = id
	d.peers = peers
	f, err := os.Create(ifile.Name)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	d.f = f
	nchunks := uint64(ifile.Size / chunks.CHUNK_SIZE)
	if ifile.Size%chunks.CHUNK_SIZE != 0 {
		nchunks++
	}
	n := nchunks / MAX_CHUNKS
	for i := uint64(0); i < n; i++ {
		start := i * MAX_CHUNKS
		end := start + MAX_CHUNKS
		if err := d.dwChunks(start, end); err != nil {
			log.Println(err)
			return err
		}
	}
	remainder := nchunks % MAX_CHUNKS
	if remainder == 0 {
		return nil
	}
	start := nchunks - remainder
	end := start + remainder
	if err := d.dwChunks(start, end); err != nil {
		log.Println(err)
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		log.Println(err)
		return err
	}
	if fi.Size() != ifile.Size {
		return fmt.Errorf("downloaded file has different size than expected: expected %d, got %d", ifile.Size, fi.Size())
	}
	return nil
}

func (d Downloader) dwChunks(start, end uint64) error {
	peerCount := 0
	result := make(chan Chunk)
	for index := start; index < end; index++ {
		peerIndex := peerCount % len(d.peers)
		go fetch(d.id, d.peers[peerIndex], index, result)
		peerCount++
	}
	chunks := make([]Chunk, end-start)
	if start == 0 {
		for i := uint64(0); i < (end - start); i++ {
			chunk := <-result
			chunks[chunk.index] = chunk
		}
	} else {
		for i := uint64(0); i < (end - start); i++ {
			chunk := <-result
			chunks[chunk.index%start] = chunk
		}
	}
	for _, chunk := range chunks {
		if chunk.err != nil {
			log.Println(chunk.err)
			return chunk.err
		}
		if _, err := io.Copy(d.f, chunk.buf); err != nil {
			log.Println(err)
			return err
		}
		ReleaseChunk(chunk.buf.Bytes())
	}
	return nil
}

func fetch(id string, addr string, index uint64, result chan Chunk) {
	var client Client
	if err := client.Dial(addr); err != nil {
		log.Println(err)
		return
	}
	defer client.Close()
	if err := client.Connect(id); err != nil {
		log.Println(err)
		return
	}
	buf, err := client.Getchunk(index)
	if err != nil {
		result <- Chunk{err, index, nil}
		return
	}
	result <- Chunk{nil, index, buf}
}
