package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezt"
	"github.com/aburdulescu/ez/hash"
	badger "github.com/dgraph-io/badger/v2"
)

type LocalServer struct {
	db *badger.DB
}

func NewLocalServer(db *badger.DB) LocalServer {
	s := LocalServer{db: db}
	http.Handle("/list", errHandler(s.handleList))
	http.Handle("/add", errHandler(s.handleAdd))
	http.Handle("/rm", errHandler(s.handleRm))
	http.Handle("/sync", errHandler(s.handleSync))
	return s
}

func (s LocalServer) Run() {
	log.Fatal(http.ListenAndServe(":22202", nil))
}

func errHandler(f func(w http.ResponseWriter, r *http.Request) (int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		log.Println(r.Method, r.URL)
		status, err := f(w, r)
		if err != nil {
			http.Error(w, err.Error(), status)
		}
	}
}

func (s LocalServer) handleSync(w http.ResponseWriter, r *http.Request) (int, error) {
	if err := updateTracker(s.db); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (s LocalServer) handleList(w http.ResponseWriter, r *http.Request) (int, error) {
	files, err := s.getFiles()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if err := respond(w, files); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (s LocalServer) getFiles() ([]ezt.IFile, error) {
	var files []ezt.IFile
	err := s.db.View(func(txn *badger.Txn) error {
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
					files = append(files, i)
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

func (s LocalServer) handleAdd(w http.ResponseWriter, r *http.Request) (int, error) {
	path := r.URL.Query().Get("path")
	if path == "" {
		return http.StatusBadRequest, errors.New("missing 'path' parameter")
	}
	if err := s.addFile(path); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (s LocalServer) addFile(path string) error {
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
	h := hash.FromChunkHashes(chunks)
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
	k := hash.ALG + "-" + h.String()
	err = s.db.Update(func(txn *badger.Txn) error {
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
		Addr: seedAddr,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(params); err != nil {
		return err
	}
	rsp, err := http.Post(trackerURL, "application/json", buf)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func (s LocalServer) handleRm(w http.ResponseWriter, r *http.Request) (int, error) {
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		return http.StatusBadRequest, errors.New("missing 'hash' parameter")
	}
	if err := s.removeFile(hash); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (s LocalServer) removeFile(hash string) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		ifileKey := hash + ".ifile"
		if err := txn.Delete([]byte(ifileKey)); err != nil {
			return err
		}
		chunksKey := hash + ".chunks"
		if err := txn.Delete([]byte(chunksKey)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", trackerURL+"?hash="+hash+"&addr="+seedAddr+":22201", nil)
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

func respond(w http.ResponseWriter, data interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		return err
	}
	return nil
}
