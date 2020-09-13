package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	badger "github.com/dgraph-io/badger/v2"
)

var cli = NewCLI(os.Args[0], []Cmd{
	Cmd{
		name:    "ls",
		desc:    "List available files",
		handler: onLs,
	},
	Cmd{
		name:    "add",
		desc:    "Add a file",
		handler: onAdd,
	},
	Cmd{
		name:    "rm",
		desc:    "Remove a file",
		handler: onRm,
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
	if err := cli.Handle(name, args); err != nil {
		handleErr(err)
	}
}

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	cli.Usage()
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
				var i IFile
				if err := json.Unmarshal(v, &i); err != nil {
					return err
				}
				fmt.Printf("%s %v\n", k, i)
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
	i, err := NewIFile(f, path)
	if err != nil {
		return err
	}
	chunks, err := ChunksFromFile(f, i.Size)
	if err != nil {
		return err
	}
	h, err := NewHash(chunks)
	if err != nil {
		return err
	}
	fmt.Printf("%s %v\n", h, i)
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&i); err != nil {
		return err
	}
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(h.String()), b.Bytes())
	})
	return err
}

func onRm(args ...string) error {
	fmt.Println("rm")
	return nil
}
