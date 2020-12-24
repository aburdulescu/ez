package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aburdulescu/ez/ezt"
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
	if err := addFile(s.db, path); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func addFile(db DB, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	i, err := ezt.NewIFile(f, path)
	if err != nil {
		return err
	}
	checksums, err := ProcessFile(f, i.Size)
	if err != nil {
		return err
	}
	id := NewID(checksums)
	if err != nil {
		return err
	}
	if err := db.Add(id, i, checksums); err != nil {
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
	id := r.URL.Query().Get("id")
	if id == "" {
		return http.StatusBadRequest, errors.New("missing 'id' parameter")
	}
	if err := removeFile(s.db, id); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func removeFile(db DB, id string) error {
	if err := db.Delete(id); err != nil {
		return err
	}
	trackerClient := ezt.NewClient(trackerURL)
	req := ezt.RemoveRequest{
		Id: id, Addr: seedAddr,
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
