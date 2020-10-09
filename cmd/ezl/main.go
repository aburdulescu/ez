package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aburdulescu/ez/cli"
	"github.com/aburdulescu/ez/ezt"

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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) < 1 {
		return fmt.Errorf("command not provided")
	}
	db, err := badger.Open(badger.DefaultOptions("./db").WithLogger(nil))
	if err != nil {
		return err
	}
	defer db.Close()
	name := args[0]
	args = args[1:]
	if err := c.Handle(name, args); err != nil {
		return err
	}
	return nil
}
func onLs(args ...string) error {
	rsp, err := http.Get("http://localhost:22202")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	var files []ezt.IFile
	if err := json.NewDecoder(rsp.Body).Decode(files); err != nil {
		return err
	}
	for _, f := range files {
		fmt.Printf("%s %s %d\n", f.Dir, f.Name, f.Size)
	}
	return nil
}

func onAdd(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("filepath wasn't provided")
	}
	path := args[0]
	data := struct {
		Filepath string `json:"filepath"`
	}{
		Filepath: path,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return err
	}
	rsp, err := http.Post("http://localhost:22202", "application/json", &buf)
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
	req, err := http.NewRequest("DELETE", cfg.TrackerURL+"?hash="+id+"&addr="+cfg.SeederAddr, nil)
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

// TODO: first get the file from the tracker, compare with local ones and send the diff to tracker
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
		Addr:  cfg.SeederAddr,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(params); err != nil {
		return err
	}
	rsp, err := http.Post(cfg.TrackerURL, "application/json", buf)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}
