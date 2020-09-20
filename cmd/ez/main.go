package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/pprof"

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
	// uncomment when profile is needed
	perfcpu, err := os.Create("perf.cpu")
	if err != nil {
		handleErr(err)
	}
	pprof.StartCPUProfile(perfcpu)
	defer pprof.StopCPUProfile()
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
	if err := download(id, r); err != nil {
		return err
	}
	return nil
}

func download(id string, r *ezt.GetResult) error {
	dist, nchunks := distributeChunks(r.IFile.Size, r.Peers)
	result := make(chan FetchResult)
	for addr, indexes := range dist {
		go fetch(id, addr, indexes, result)
	}
	chunkData := make(map[uint64]ChunkData, nchunks)
	for i := 0; i < len(dist); i++ {
		fetchRes := <-result
		if fetchRes.err != nil {
			continue
		}
		for k, v := range fetchRes.chunks {
			chunkData[k] = v
		}
	}
	buf := new(bytes.Buffer)
	buf.Grow(int(r.IFile.Size))
	for i := uint64(0); i < nchunks; i++ {
		d, ok := chunkData[i]
		if !ok {
			log.Printf("chunk index %d not found\n", i)
			continue
		}
		if d.err != nil {
			log.Println(i, d.err)
			continue
		}
		if _, err := io.Copy(buf, d.buf); err != nil {
			log.Println(i, d.err)
			continue
		}
	}
	if int64(buf.Len()) != r.IFile.Size {
		return fmt.Errorf("downloaded file has different size than expected: expected %d, got %d", r.IFile.Size, buf.Len())
	}
	f, err := os.Create(r.IFile.Name)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, buf); err != nil {
		return err
	}
	return nil
}

func distributeChunks(fileSize int64, peers []string) (map[string][]uint64, uint64) {
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
		return dist, nchunks
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
		return dist, nchunks
	}
}

type ChunkData struct {
	err error
	buf *bytes.Buffer
}

type FetchResult struct {
	err    error
	chunks map[uint64]ChunkData
}

func fetch(id, addr string, indexes []uint64, r chan FetchResult) {
	var client Client
	if err := client.Dial(addr); err != nil {
		log.Printf("%s: %v", addr, err)
		r <- FetchResult{err, nil}
		return
	}
	defer client.Close()
	if err := client.Connect(id); err != nil {
		log.Printf("%s: %v", addr, err)
		r <- FetchResult{err, nil}
		return
	}
	chunkData := make(map[uint64]ChunkData, len(indexes))
	for _, v := range indexes {
		buf, err := fetchChunk(client, v) // TODO: do this in a separate goroutine
		if err != nil {
			log.Printf("%s: %d: %v", addr, v, err)
			chunkData[v] = ChunkData{err, nil}
			continue
		}
		chunkData[v] = ChunkData{nil, buf}
	}
	r <- FetchResult{nil, chunkData}
}

func fetchChunk(client Client, i uint64) (*bytes.Buffer, error) {
	chunkHash, ch, err := client.Getchunk(i)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.Grow(chunks.CHUNK_SIZE)
	for part := range ch {
		if part.err != nil {
			return nil, err
		}
		if _, err := buf.Write(part.piece); err != nil {
			return nil, err
		}
	}
	calcChunkHash := hash.FromChunk(buf.Bytes())
	if !calcChunkHash.Equals(chunkHash) {
		// TODO: don't return err, retry download from other peer(or maybe the same peer?)
		return nil, fmt.Errorf("hash of chunk %d differs from hash provided by peer", i)
	}
	return buf, nil
}
