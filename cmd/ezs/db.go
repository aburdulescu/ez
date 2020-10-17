package main

import (
	"bytes"
	"encoding/json"
	"log"
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

func (db DB) GetAll() ([]ezt.File, error) {
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

func (db DB) GetIFile(k string) (ezt.IFile, error) {
	v, err := db.get(k + ".ifile")
	if err != nil {
		return ezt.IFile{}, err
	}
	var ifile ezt.IFile
	if err := json.Unmarshal(v, &ifile); err != nil {
		return ezt.IFile{}, err
	}
	return ifile, nil
}

func (db DB) GetChunkHashes(k string) ([]hash.Hash, error) {
	v, err := db.get(k + ".chunks")
	if err != nil {
		return nil, err
	}
	var hashes []hash.Hash
	if err := json.Unmarshal(v, &hashes); err != nil {
		return nil, err
	}
	return hashes, nil
}

func (db DB) Add(h hash.Hash, i ezt.IFile, chunks []hash.Hash) (string, error) {
	ifileBuf := new(bytes.Buffer)
	if err := json.NewEncoder(ifileBuf).Encode(&i); err != nil {
		return "", err
	}
	chunksBuf := new(bytes.Buffer)
	if err := json.NewEncoder(chunksBuf).Encode(&chunks); err != nil {
		return "", err
	}
	k := hash.ALG + "-" + h.String()
	entries := []*badger.Entry{
		badger.NewEntry([]byte(k+".ifile"), ifileBuf.Bytes()),
		badger.NewEntry([]byte(k+".chunks"), chunksBuf.Bytes()),
	}
	return k, db.add(entries)
}

func (db DB) Delete(k string) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		ifileKey := k + ".ifile"
		if err := txn.Delete([]byte(ifileKey)); err != nil {
			return err
		}
		chunksKey := k + ".chunks"
		if err := txn.Delete([]byte(chunksKey)); err != nil {
			return err
		}
		return nil
	})
	return err
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

func (db DB) get(k string) ([]byte, error) {
	var v []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			log.Println(err)
			return err
		}
		v, err = item.ValueCopy(nil)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	return v, err
}
