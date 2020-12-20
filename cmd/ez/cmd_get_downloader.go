package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezt"
	pb "github.com/cheggaaa/pb/v3"
)

const MAX_CHUNKS uint64 = 8

type Downloader struct {
	id       string
	f        *os.File
	connPool *ConnPool
	peers    []string
	pb       *pb.ProgressBar
}

type Chunk struct {
	err   error
	index uint64
	buf   *bytes.Buffer
}

func (d *Downloader) Run(id string, ifile ezt.IFile, peers []string) error {
	d.id = id

	connPool, goodPeers, err := NewConnPool(peers, DialSeederClient)
	if err != nil {
		return err
	}
	if connPool.Len() == 0 {
		return fmt.Errorf("no peers available")
	}
	d.connPool = connPool
	d.peers = goodPeers
	defer d.connPool.Release()
	if err := d.connPool.Connect(id); err != nil {
		return err
	}
	defer d.connPool.Disconnect()

	f, err := os.Create(ifile.Name)
	if err != nil {
		log.Println(err)
		return err
	}
	if err := f.Truncate(ifile.Size); err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	d.f = f

	d.pb = pb.New64(ifile.Size)
	d.pb.Set(pb.Bytes, true)
	d.pb.Set(pb.SIBytesPrefix, true)
	d.pb.Start()
	defer d.pb.Finish()

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
		peer := d.peers[peerIndex]
		client, err := d.connPool.Get(d.id, peer)
		if err != nil {
			// TODO: don't skip it, maybe try with other peer?
			continue
		}
		go d.fetch(peer, client, index, result)
		peerCount++
	}
	for i := uint64(0); i < (end - start); i++ {
		chunk := <-result
		if chunk.err != nil {
			log.Println(chunk.err)
			// TODO: retry chunk download
			continue
		}
		off := int64(chunk.index * chunks.CHUNK_SIZE)
		n, err := d.f.WriteAt(chunk.buf.Bytes(), off)
		if err != nil {
			log.Println(err, n)
			return err
		}
		d.pb.Add(n)
		ReleaseChunk(chunk.buf.Bytes())
	}
	return nil
}

func (d Downloader) fetch(peer string, client *SeederClient, index uint64, result chan Chunk) {
	buf, err := client.Getchunk(index)
	if err != nil {
		result <- Chunk{err, index, nil}
		return
	}
	d.connPool.Put(peer, client)
	result <- Chunk{nil, index, buf}
}
