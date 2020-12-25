package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type LocalServer struct {
	db          *DB
	watcherChan chan<- WatcherEntry
}

func NewLocalServer(db *DB, watcherChan chan<- WatcherEntry) LocalServer {
	s := LocalServer{db: db, watcherChan: watcherChan}
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
	files, err := s.db.GetFiles()
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
	id, err := AddFile(s.db, path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	s.watcherChan <- WatcherEntry{
		Op:   AddOp,
		Path: path,
		Id:   id,
	}
	return http.StatusOK, nil
}

func (s LocalServer) handleRm(w http.ResponseWriter, r *http.Request) (int, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return http.StatusBadRequest, errors.New("missing 'id' parameter")
	}
	if err := RemoveFile(s.db, id); err != nil {
		return http.StatusInternalServerError, err
	}
	s.watcherChan <- WatcherEntry{
		Op: RmOp,
		Id: id,
	}
	return http.StatusOK, nil
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
