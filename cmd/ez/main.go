package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aburdulescu/go-ez/chunks"
	"github.com/aburdulescu/go-ez/cli"
	"github.com/aburdulescu/go-ez/ezt"
	// "github.com/pkg/profile"
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
	// defer profile.Start(profile.ProfilePath("."), profile.CPUProfile).Stop()
	// defer profile.Start(profile.ProfilePath("."), profile.MemProfile).Stop()
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
	dist := distributeChunks(r.IFile.Size, r.Peers)
	f, err := os.Create(r.IFile.Name)
	if err != nil {
		return err
	}
	defer f.Close()
	for addr, indexes := range dist {
		if err := fetchAndWriteChunks(addr, id, indexes, f); err != nil {
			return err
		}
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() != r.IFile.Size {
		return fmt.Errorf("downloaded file has different size than expected: expected %d, got %d", r.IFile.Size, fi.Size())
	}
	return nil
}

func fetchAndWriteChunks(addr string, id string, indexes []uint64, f *os.File) error {
	var client Client
	if err := client.Dial(addr); err != nil {
		log.Printf("%s: %v", addr, err)
		return err
	}
	defer client.Close()
	if err := client.Connect(id); err != nil {
		log.Printf("%s: %v", addr, err)
		return err
	}
	for _, index := range indexes {
		buf, err := client.Getchunk(index)
		if err != nil {
			log.Printf("%s: %d: %v", addr, index, err)
			return err
		}
		if _, err := io.Copy(f, buf); err != nil {
			log.Println(index, err)
			return err
		}
	}
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
			// add remainder chunks to the last peer
			peerIdx := peersLen - 1
			for i := nchunks - remainder; i < nchunks; i++ {
				dist[peers[peerIdx]] = append(dist[peers[peerIdx]], i)
			}
		}
		return dist
	}
}
