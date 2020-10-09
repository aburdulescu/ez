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
	return LocalServer{db: db}
}

func (s LocalServer) handlerRequest(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	var err error
	var status int
	switch r.Method {
	case "GET":
		status, err = s.handleGet(w, r)
	case "POST":
		status, err = s.handlePost(w, r)
	case "DELETE":
		status, err = s.handleDelete(w, r)
	default:
		err = errors.New("unknown HTTP method")
		status = http.StatusBadRequest
	}
	if err != nil {
		http.Error(w, err.Error(), status)
	}
}

func (s LocalServer) handleGet(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
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

func (s LocalServer) handlePost(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	data := struct {
		Filepath string `json:"filepath"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return http.StatusBadRequest, err
	}
	if err := s.addFile(data.Filepath); err != nil {
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
	rsp, err := http.Post(trackerAddr, "application/json", buf)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func (s LocalServer) handleDelete(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		return http.StatusBadRequest, errors.New("missing 'hash' parameter")
	}
	return http.StatusOK, nil
}

func respond(w http.ResponseWriter, data interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	buf.Truncate(buf.Len() - 1)
	if _, err := io.Copy(w, &buf); err != nil {
		return err
	}
	return nil
}
