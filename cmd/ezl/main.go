package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aburdulescu/go-ez/chunks"
	"github.com/aburdulescu/go-ez/cli"
	"github.com/aburdulescu/go-ez/ezt"
	"github.com/aburdulescu/go-ez/hash"

	badger "github.com/dgraph-io/badger/v2"
)

var c = cli.New(os.Args[0], []cli.Cmd{
	cli.Cmd{
		Name:    "ls",
		Desc:    "List available files",
		Handler: onLs,
	},
	cli.Cmd{
		Name:    "add",
		Desc:    "Add a file",
		Handler: onAdd,
	},
	cli.Cmd{
		Name:    "rm",
		Desc:    "Remove a file",
		Handler: onRm,
	},
	cli.Cmd{
		Name:    "sync",
		Desc:    "Send all available files to the tracker",
		Handler: onSync,
	},
})

var db *badger.DB

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		handleErr(fmt.Errorf("command not provided"))
	}
	var err error
	db, err = badger.Open(badger.DefaultOptions("db").WithLogger(nil))
	if err != nil {
		handleErr(err)
	}
	defer db.Close()
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
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				kstr := string(k)
				if strings.HasSuffix(kstr, "ifile") {
					var i ezt.IFile
					if err := json.Unmarshal(v, &i); err != nil {
						return err
					}
					fmt.Printf("%s %v\n", k, i)
				} else if strings.HasSuffix(kstr, "chunks") {
					var c []hash.Hash
					if err := json.Unmarshal(v, &c); err != nil {
						return err
					}
					fmt.Printf("%s %v\n", k, len(c))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func onAdd(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("filepath wasn't provided")
	}
	path := args[0]
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	i, err := ezt.NewIFile(f, path)
	if err != nil {
		return err
	}
	chunks, err := chunks.FromFile(f, i.Size)
	if err != nil {
		return err
	}
	h, err := hash.New(chunks)
	if err != nil {
		return err
	}
	ifileBuf := new(bytes.Buffer)
	if err := json.NewEncoder(ifileBuf).Encode(&i); err != nil {
		return err
	}
	chunksBuf := new(bytes.Buffer)
	if err := json.NewEncoder(chunksBuf).Encode(&chunks); err != nil {
		return err
	}
	k := hash.HASH_ALG + "-" + h.String()
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(k+".ifile"), ifileBuf.Bytes())
		if err != nil {
			return err
		}
		err = txn.Set([]byte(k+".chunks"), chunksBuf.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Println(k)
	params := ezt.PostParams{
		Files: []ezt.File{
			ezt.File{Hash: k, IFile: i},
		},
		Addr: "localhost:22334",
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(params); err != nil {
		return err
	}
	rsp, err := http.Post("http://localhost:8080/", "application/json", buf)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func onRm(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
	err := db.Update(func(txn *badger.Txn) error {
		ifileKey := id + ".ifile"
		if err := txn.Delete([]byte(ifileKey)); err != nil {
			return err
		}
		chunksKey := id + ".chunks"
		if err := txn.Delete([]byte(chunksKey)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", "http://localhost:8080/?hash="+id+"&addr=localhost:22334", nil)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func onSync(args ...string) error {
	var files []ezt.File
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				kstr := string(k)
				if strings.HasSuffix(kstr, "ifile") {
					var i ezt.IFile
					if err := json.Unmarshal(v, &i); err != nil {
						return err
					}
					id := strings.Split(kstr, ".")[0]
					files = append(files, ezt.File{Hash: id, IFile: i})
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	params := ezt.PostParams{
		Files: files,
		Addr:  "localhost:22334",
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(params); err != nil {
		return err
	}
	rsp, err := http.Post("http://localhost:8080/", "application/json", buf)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}
