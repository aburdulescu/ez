package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aburdulescu/go-ez/chunks"
	"github.com/aburdulescu/go-ez/cli"
	"github.com/aburdulescu/go-ez/ezt"
	"github.com/aburdulescu/go-ez/hash"
)

type Config struct {
	TrackerURL string `json:"trackerUrl"`
}

var c = cli.New(os.Args[0], []cli.Cmd{
	cli.Cmd{
		Name:    "ls",
		Desc:    "List available files",
		Handler: onLs,
	},
	cli.Cmd{
		Name:    "get",
		Desc:    "Download a file",
		Handler: onGet,
	},
})

var cfg Config

func main() {
	f, err := os.Open("ez.json")
	if err != nil {
		handleErr(err)
	}
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		handleErr(err)
	}
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	args := os.Args[1:]
	if len(args) < 1 {
		handleErr(fmt.Errorf("command not provided"))
	}
	name := args[0]
	args = args[1:]
	if err := c.Handle(name, args); err != nil {
		handleErr(err)
	}
}

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	c.Usage()
	os.Exit(1)
}

func onLs(args ...string) error {
	rsp, err := http.Get(cfg.TrackerURL + "?hash=all")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	var files []ezt.GetAllResult
	if err := json.NewDecoder(rsp.Body).Decode(&files); err != nil {
		return err
	}
	for _, f := range files {
		fmt.Printf("%s\t\t%s\t\t%d\n", f.Hash, f.Name, f.Size)
	}
	return nil
}

func onGet(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
	rsp, err := http.Get(cfg.TrackerURL + "?hash=" + id)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	r := &ezt.GetResult{}
	if err := json.NewDecoder(rsp.Body).Decode(r); err != nil {
		return err
	}
	dist := distributeChunks(r.IFile.Size, r.Peers)
	for k, v := range dist {
		log.Println(k, v)
	}
	peerAddr := r.Peers[0]
	var client Client
	if err := client.Dial(peerAddr); err != nil {
		return err
	}
	defer client.Close()
	if err := client.Connect(id); err != nil {
		return err
	}
	fileBuf := new(bytes.Buffer)
	nchunks := uint64(r.IFile.Size / chunks.CHUNK_SIZE)
	if r.IFile.Size%chunks.CHUNK_SIZE != 0 {
		nchunks++
	}
	for i := uint64(0); i < nchunks; i++ {
		buf, err := fetchChunk(client, i)
		if err != nil {
			return err
		}
		if _, err := io.Copy(fileBuf, buf); err != nil {
			return err
		}
	}
	if int64(fileBuf.Len()) != r.IFile.Size {
		return fmt.Errorf("downloaded file has different size than expected: expected %d, got %d", r.IFile.Size, fileBuf.Len())
	}
	// f, err := os.Create(r.IFile.Name)
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()
	// if _, err := io.Copy(f, buf); err != nil {
	// 	return err
	// }
	return nil
}

func distributeChunks(fileSize int64, peers []string) map[string][]uint64 {
	peersLen := uint64(len(peers))
	nchunks := uint64(fileSize / chunks.CHUNK_SIZE)
	if fileSize%chunks.CHUNK_SIZE != 0 {
		nchunks++
	}
	if nchunks <= peersLen {
		// select nchunks number of peers if available
		// ex: nchunks=10, npeers=20 => select 10 peers from the list and get from each one chunk
		dist := make(map[string][]uint64, nchunks)
		for i := uint64(0); i < nchunks; i++ {
			dist[peers[i]] = []uint64{i}
		}
		return dist
	} else {
		// split the chunks between the available peers
		// ex: nchunks=41, npeers=3 => split chunks between the 3 peers: peer1=13, peer2=13, peer3=15
		nchunksPerPeer := nchunks / peersLen
		dist := make(map[string][]uint64, peersLen)
		for i := uint64(0); i < peersLen; i++ {
			indexes := make([]uint64, nchunksPerPeer)
			n := 0
			for j := i * nchunksPerPeer; j < (i+1)*nchunksPerPeer; j++ {
				indexes[n] = j
				n++
			}
			dist[peers[i]] = indexes
		}
		remainder := nchunks % peersLen
		if remainder != 0 {
			// add remainder chunks to one(or more) peers
			peerIdx := peersLen - 1
			for i := nchunks - remainder; i < nchunks; i++ {
				dist[peers[peerIdx]] = append(dist[peers[peerIdx]], i)
				peerIdx--
			}
		}
		return dist
	}
}

type ChunkData struct {
	err   error
	index uint64
	buf   *bytes.Buffer
}

type FetchResult struct {
	err    error
	chunks []ChunkData
}

func fetch(id, addr string, indexes []uint64, r chan FetchResult) {
	var client Client
	if err := client.Dial(addr); err != nil {
		r <- FetchResult{err, nil}
		return
	}
	defer client.Close()
	if err := client.Connect(id); err != nil {
		r <- FetchResult{err, nil}
		return
	}
	var chunkData []ChunkData
	resultBuf := new(bytes.Buffer)
	for _, v := range indexes {
		buf, err := fetchChunk(client, v)
		if err != nil {
			chunkData = append(chunkData, ChunkData{err, v, nil})
			continue
		}
		if _, err := io.Copy(resultBuf, buf); err != nil {
			chunkData = append(chunkData, ChunkData{err, v, nil})
			return
		}
		chunkData = append(chunkData, ChunkData{nil, v, buf})
	}
	r <- FetchResult{nil, chunkData}
}

func fetchChunk(client Client, i uint64) (*bytes.Buffer, error) {
	chunkHash, ch, err := client.Getchunk(i)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	for part := range ch {
		if part.err != nil {
			return nil, err
		}
		if _, err := buf.Write(part.piece); err != nil {
			return nil, err
		}
	}
	calcChunkHash, err := hash.FromChunk(buf.Bytes())
	if err != nil {
		return nil, err
	}
	if !calcChunkHash.Equals(chunkHash) {
		// TODO: don't return err, retry download from other peer(or maybe the same peer?)
		return nil, fmt.Errorf("hash of chunk %d differs from hash provided by peer", i)
	}
	return buf, nil
}
