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
	f       *os.File
	clients []Client
}

type Chunk struct {
	err   error
	index uint64
	buf   *bytes.Buffer
}

func (d *Downloader) Run(id string, ifile ezt.IFile, peers []string) error {
	f, err := os.Create(ifile.Name)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	d.f = f
	d.clients = make([]Client, len(peers))
	for i := range peers {
		var client Client
		if err := client.Dial(peers[i]); err != nil {
			log.Println(err)
			return err
		}
		defer client.Close()
		if err := client.Connect(id); err != nil {
			log.Println(err)
			return err
		}
		d.clients[i] = client
	}
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
	clientCount := 0
	result := make(chan Chunk)
	for index := start; index < end; index++ {
		peerIndex := clientCount % len(d.clients)
		go fetch(d.clients[peerIndex], index, result)
		clientCount++
	}
	chunks := make([]Chunk, end-start)
	for j := uint64(0); j < (end - start); j++ {
		chunk := <-result
		chunks[chunk.index] = chunk
		log.Println("recv:", chunk.index)
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
	}
	return nil
}

func fetch(c Client, index uint64, result chan Chunk) {
	log.Println("fetch:", index)
	buf, err := c.Getchunk(index)
	if err != nil {
		result <- Chunk{err, index, nil}
		return
	}
	result <- Chunk{nil, index, buf}
}
