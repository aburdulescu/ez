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
					id := strings.Split(kstr, ".")[0]
					files = append(files, ezt.File{Id: id, IFile: i})
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

func (db DB) GetIFile(id string) (ezt.IFile, error) {
	v, err := db.get(id + ".ifile")
	if err != nil {
		return ezt.IFile{}, err
	}
	var ifile ezt.IFile
	if err := json.Unmarshal(v, &ifile); err != nil {
		return ezt.IFile{}, err
	}
	return ifile, nil
}

func (db DB) GetChecksums(id string) ([]hash.Checksum, error) {
	v, err := db.get(id + ".checksums")
	if err != nil {
		return nil, err
	}
	var checksums []hash.Checksum
	if err := json.Unmarshal(v, &checksums); err != nil {
		return nil, err
	}
	return checksums, nil
}

func (db DB) Add(id string, ifile ezt.IFile, checksums []hash.Checksum) error {
	ifileBuf := new(bytes.Buffer)
	if err := json.NewEncoder(ifileBuf).Encode(&ifile); err != nil {
		return err
	}
	checksumsBuf := new(bytes.Buffer)
	if err := json.NewEncoder(checksumsBuf).Encode(&checksums); err != nil {
		return err
	}
	entries := []*badger.Entry{
		badger.NewEntry([]byte(id+".ifile"), ifileBuf.Bytes()),
		badger.NewEntry([]byte(id+".checksums"), checksumsBuf.Bytes()),
	}
	return db.add(entries)
}

func (db DB) Delete(id string) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		ifileKey := id + ".ifile"
		if err := txn.Delete([]byte(ifileKey)); err != nil {
			return err
		}
		checksumsKey := id + ".checksums"
		if err := txn.Delete([]byte(checksumsKey)); err != nil {
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

func (db DB) get(id string) ([]byte, error) {
	var v []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
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
