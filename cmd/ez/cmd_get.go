package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezt"
	"github.com/spf13/cobra"
	// pb "github.com/cheggaaa/pb/v3" // TODO: use it for progress bar
)

func onGet(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
	trackerURL, err := getTrackerURL()
	if err != nil {
		return err
	}
	rsp, err := http.Get(trackerURL + "?hash=" + id)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rsp.Body.Close()
	r := ezt.GetResult{}
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		log.Println(err)
		return err
	}
	var d Downloader
	if err := d.Run(id, r.IFile, r.Peers); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

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
	d.peers = removeUnavailablePeers(peers)
	if len(d.peers) == 0 {
		return fmt.Errorf("no peers available")
	}
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

func removeUnavailablePeers(peers []string) []string {
	var goodPeers []string
	for _, peer := range peers {
		c, err := DialEzs(peer)
		if err != nil {
			log.Println(err)
			continue
		}
		c.Close()
		goodPeers = append(goodPeers, peer)
	}
	return goodPeers
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
	c, err := DialEzs(addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()
	if err := c.Connect(id); err != nil {
		log.Println(err)
		return
	}
	buf, err := c.Getchunk(index)
	if err != nil {
		result <- Chunk{err, index, nil}
		return
	}
	result <- Chunk{nil, index, buf}
}
