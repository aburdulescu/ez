package main

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/aburdulescu/ez/ezt"
	"github.com/aburdulescu/ez/hash"
	badger "github.com/dgraph-io/badger/v2"
)

type DB struct {
	db *badger.DB
}

func NewDB(path string) (DB, error) {
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return DB{}, err
	}
	return DB{db}, nil
}

func (db DB) Close() {
	db.db.Close()
}

func (db DB) List() ([]ezt.File, error) {
	var files []ezt.File
	err := db.db.View(func(txn *badger.Txn) error {
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
					hash := strings.Split(kstr, ".")[0]
					files = append(files, ezt.File{Hash: hash, IFile: i})
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return files, err
}

type entry struct {
	k []byte
	v []byte
}

func (db DB) add(entries []*badger.Entry) error {
	return db.db.Update(func(txn *badger.Txn) error {
		for _, entry := range entries {
			err := txn.SetEntry(entry)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db DB) Add(h hash.Hash, i ezt.IFile, chunks []hash.Hash) error {
	ifileBuf := new(bytes.Buffer)
	if err := json.NewEncoder(ifileBuf).Encode(&i); err != nil {
		return err
	}
	chunksBuf := new(bytes.Buffer)
	if err := json.NewEncoder(chunksBuf).Encode(&chunks); err != nil {
		return err
	}
	k := hash.ALG + "-" + h.String()
	entries := []*badger.Entry{
		badger.NewEntry([]byte(k+".ifile"), ifileBuf.Bytes()),
		badger.NewEntry([]byte(k+".chunks"), chunksBuf.Bytes()),
	}
	return db.add(entries)
}
