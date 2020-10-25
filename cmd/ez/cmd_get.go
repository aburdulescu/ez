package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"

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
	trackerClient := ezt.NewClient(trackerURL)
	rsp, err := trackerClient.Get(ezt.GetRequest{id})
	if err != nil {
		log.Println(err)
		return err
	}
	var d Downloader
	if err := d.Run(id, rsp.IFile, rsp.Peers); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

const MAX_CHUNKS uint64 = 8

type Downloader struct {
	id       string
	f        *os.File
	connPool *ConnPool
	peers    []string
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

type ConnPoolDialFunc func(addr string) (*SeederClient, error)

type ConnPool struct {
	mu   sync.RWMutex
	data map[string][]*SeederClient

	dialFunc ConnPoolDialFunc
}

func NewConnPool(peers []string, dialFunc ConnPoolDialFunc) (*ConnPool, []string, error) {
	var goodPeers []string
	data := make(map[string][]*SeederClient)
	for _, peer := range peers {
		c, err := dialFunc(peer)
		if err != nil {
			log.Println(err)
			continue
		}
		data[peer] = append(data[peer], c)
		goodPeers = append(goodPeers, peer)
	}
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("no peers available")
	}
	pool := &ConnPool{
		data:     data,
		dialFunc: dialFunc,
	}
	return pool, goodPeers, nil
}

func (p *ConnPool) Get(id string, addr string) (*SeederClient, error) {
	p.mu.Lock()
	clients, ok := p.data[addr]
	if ok && len(clients) != 0 {
		client := clients[0]
		clients = clients[1:]
		if len(clients) == 0 {
			delete(p.data, addr)
		} else {
			p.data[addr] = clients
		}
		p.mu.Unlock()
		return client, nil
	} else {
		p.mu.Unlock()
		client, err := p.dialFunc(addr)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if err := client.Connect(id); err != nil {
			client.Close()
			return nil, err
		}
		return client, nil
	}
}

func (p *ConnPool) Put(addr string, client *SeederClient) {
	p.mu.Lock()
	p.data[addr] = append(p.data[addr], client)
	p.mu.Unlock()
}

func (p *ConnPool) Len() int {
	p.mu.RLock()
	l := len(p.data)
	p.mu.RUnlock()
	return l
}

func (p *ConnPool) Connect(id string) error {
	p.mu.RLock()
	for _, clients := range p.data {
		for _, client := range clients {
			if err := client.Connect(id); err != nil {
				// TODO: remove conn if connect fails
				return err
			}
		}
	}
	p.mu.RUnlock()
	return nil
}

func (p *ConnPool) Disconnect() {
	p.mu.RLock()
	for _, clients := range p.data {
		for _, client := range clients {
			client.Disconnect()
		}
	}
	p.mu.RUnlock()
}

func (p *ConnPool) Release() {
	p.mu.Lock()
	for _, clients := range p.data {
		for i := range clients {
			clients[i].Close()
		}
	}
	p.data = make(map[string][]*SeederClient)
	p.mu.Unlock()
}
