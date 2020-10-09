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
}

func NewLocalServer() LocalServer {
	return LocalServer{}
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
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		return http.StatusBadRequest, errors.New("missing 'hash' parameter")
	}
	return http.StatusOK, nil
}

func (s LocalServer) handlePost(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	return http.StatusOK, nil
}

func (s LocalServer) handleDelete(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		return http.StatusBadRequest, errors.New("missing 'hash' parameter")
	}
	return http.StatusOK, nil
}

func respond(w http.ResponseWriter, data interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	buf.Truncate(buf.Len() - 1)
	if _, err := io.Copy(w, &buf); err != nil {
		log.Println("respond:", err)
	}
}
