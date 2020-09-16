package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/aburdulescu/go-ez/chunks"
	"github.com/aburdulescu/go-ez/cli"
	"github.com/aburdulescu/go-ez/ezt"
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
	rsp, err := http.Get("http://localhost:8080/?hash=all")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	var files []ezt.GetAllResult
	if err := json.NewDecoder(rsp.Body).Decode(&files); err != nil {
		return err
	}
	for _, f := range files {
		fmt.Println(f)
	}
	return nil
}

func onGet(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
	rsp, err := http.Get("http://localhost:8080/?hash=" + id)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	r := &ezt.GetResult{}
	if err := json.NewDecoder(rsp.Body).Decode(r); err != nil {
		return err
	}
	fmt.Println(r)

	peersLen := uint64(len(r.Peers))
	nchunks := uint64(math.Ceil(float64(r.IFile.Size) / float64(chunks.CHUNK_SIZE)))
	if nchunks <= peersLen {
		// select nchunks number of peers if available
		// ex: nchunks=10, npeers=20 => select 10 peers from the list and get from each one chunk
		nchunksPerPeer := 1
		fmt.Printf("nchunks=%d, chunksPerPeer=%d\n", nchunks, nchunksPerPeer)
	} else {
		// split the chunks between the available peers
		// ex: nchunks=41, npeers=3 => split chunks between the 3 peers: peer1=13, peer2=13, peer3=15
		nchunksPerPeer := nchunks / peersLen
		remainder := nchunks % peersLen
		fmt.Printf("nchunks=%d, chunksPerPeer=%d, remainder=%d\n", nchunks, nchunksPerPeer, remainder)
		if remainder != 0 {
			// add remainder chunks to one(or more) peers
		}
	}

	if err := fetch(r.IFile.Name, id); err != nil {
		return err
	}

	return nil
}

func fetch(name string, id string) error {
	var client Client
	if err := client.Dial(":8081"); err != nil {
		return err
	}
	defer client.Close()
	if err := client.Connect(id); err != nil {
		return err
	}
	chunkHash, ch, err := client.Getchunk(0)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", chunkHash)
	buf := new(bytes.Buffer)
	for part := range ch {
		if part.err != nil {
			return err
		}
		_, err := buf.Write(part.piece)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, buf); err != nil {
		return err
	}
	return nil
}
