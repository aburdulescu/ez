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

const (
	TRACKER_HOST = "localhost"
	TRACKER_ADDR = "http://" + TRACKER_HOST + ":23230/"
)

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

func main() {
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
	rsp, err := http.Get(TRACKER_ADDR + "?hash=all")
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
	rsp, err := http.Get(TRACKER_ADDR + "?hash=" + id)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	r := &ezt.GetResult{}
	if err := json.NewDecoder(rsp.Body).Decode(r); err != nil {
		return err
	}
	// peersLen := uint64(len(r.Peers))
	nchunks := uint64(r.IFile.Size / chunks.CHUNK_SIZE)
	if r.IFile.Size%chunks.CHUNK_SIZE != 0 {
		nchunks++
	}
	// if nchunks <= peersLen {
	// 	// select nchunks number of peers if available
	// 	// ex: nchunks=10, npeers=20 => select 10 peers from the list and get from each one chunk
	// 	nchunksPerPeer := 1
	// 	log.Printf("nchunks=%d, chunksPerPeer=%d\n", nchunks, nchunksPerPeer)
	// } else {
	// 	// split the chunks between the available peers
	// 	// ex: nchunks=41, npeers=3 => split chunks between the 3 peers: peer1=13, peer2=13, peer3=15
	// 	nchunksPerPeer := nchunks / peersLen
	// 	remainder := nchunks % peersLen
	// 	log.Printf("nchunks=%d, chunksPerPeer=%d, remainder=%d\n", nchunks, nchunksPerPeer, remainder)
	// 	if remainder != 0 {
	// 		// add remainder chunks to one(or more) peers
	// 	}
	// }
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
