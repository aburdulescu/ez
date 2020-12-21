package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aburdulescu/ez/chunks"
	"github.com/aburdulescu/ez/ezt"
	"github.com/aburdulescu/ez/hash"
)

type LocalServer struct {
	db DB
}

func NewLocalServer(db DB) LocalServer {
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
	files, err := s.db.GetAll()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if err := respond(w, files); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
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
	checksums, err := chunks.FromFile(f, i.Size)
	if err != nil {
		return err
	}
	id := hash.NewID(checksums)
	if err != nil {
		return err
	}
	if err := s.db.Add(id, i, checksums); err != nil {
		return err
	}
	trackerClient := ezt.NewClient(trackerURL)
	req := ezt.AddRequest{
		Files: []ezt.File{
			ezt.File{Id: id, IFile: i},
		},
		Addr: seedAddr,
	}
	if err := trackerClient.Add(req); err != nil {
		return err
	}
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
	if err := s.db.Delete(hash); err != nil {
		return err
	}
	trackerClient := ezt.NewClient(trackerURL)
	req := ezt.RemoveRequest{
		Id: hash, Addr: seedAddr,
	}
	if err := trackerClient.Remove(req); err != nil {
		return err
	}
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
