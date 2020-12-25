package main

import (
	"bytes"
	"encoding/json"

	"github.com/aburdulescu/ez/cmn"
	"github.com/aburdulescu/ez/ezt"
	bolt "go.etcd.io/bbolt"
)

var (
	IfilesBucket    = []byte("ifiles")
	ChecksumsBucket = []byte("checksums")
)

type DB struct {
	db *bolt.DB
}

func NewDB(path string) (*DB, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(IfilesBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(ChecksumsBucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db DB) Close() error {
	return db.db.Close()
}

func (db DB) Add(id string, ifile ezt.IFile, checksums []cmn.Checksum) error {
	ifileBuf := new(bytes.Buffer)
	if err := json.NewEncoder(ifileBuf).Encode(&ifile); err != nil {
		return err
	}
	checksumsBuf := new(bytes.Buffer)
	if err := json.NewEncoder(checksumsBuf).Encode(&checksums); err != nil {
		return err
	}
	return db.db.Update(func(tx *bolt.Tx) error {
		ifilesBucket := tx.Bucket(IfilesBucket)
		if err := ifilesBucket.Put([]byte(id), ifileBuf.Bytes()); err != nil {
			return err
		}
		checksumsBucket := tx.Bucket(ChecksumsBucket)
		if err := checksumsBucket.Put([]byte(id), checksumsBuf.Bytes()); err != nil {
			return err
		}
		return nil
	})
}

func (db DB) Delete(id string) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		ifilesBucket := tx.Bucket(IfilesBucket)
		if err := ifilesBucket.Delete([]byte(id)); err != nil {
			return err
		}
		checksumsBucket := tx.Bucket(ChecksumsBucket)
		if err := checksumsBucket.Delete([]byte(id)); err != nil {
			return err
		}
		return nil
	})
}

func (db DB) GetIFile(id string) (ezt.IFile, error) {
	b, err := db.get(IfilesBucket, []byte(id))
	if err != nil {
		return ezt.IFile{}, err
	}
	if err != nil {
		return ezt.IFile{}, err
	}
	var ifile ezt.IFile
	if err := json.Unmarshal(b, &ifile); err != nil {
		return ezt.IFile{}, err
	}
	return ifile, nil
}

func (db DB) GetChecksums(id string) ([]cmn.Checksum, error) {
	b, err := db.get(ChecksumsBucket, []byte(id))
	if err != nil {
		return nil, err
	}
	var checksums []cmn.Checksum
	if err := json.Unmarshal(b, &checksums); err != nil {
		return nil, err
	}
	return checksums, nil
}

func (db DB) GetFiles() ([]ezt.File, error) {
	var files []ezt.File
	err := db.db.View(func(tx *bolt.Tx) error {
		bck := tx.Bucket(IfilesBucket)
		c := bck.Cursor()
		defer c.Delete()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			id := string(k)
			var ifile ezt.IFile
			if err := json.Unmarshal(v, &ifile); err != nil {
				return err
			}
			files = append(files, ezt.File{Id: id, IFile: ifile})
		}
		return nil
	})
	return files, err
}

func (db DB) get(bucket, key []byte) ([]byte, error) {
	var b []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		bck := tx.Bucket(bucket)
		v := bck.Get(key)
		b = make([]byte, len(v))
		copy(b, v)
		return nil
	})
	return b, err
}
