package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aburdulescu/ez/ezt"
)

type Server struct {
	c *KV
}

func main() {
	s := Server{NewKV()}
	http.HandleFunc("/", s.handleRequest)
	log.Fatal(http.ListenAndServe(":23230", nil))
}

func (s Server) handleRequest(w http.ResponseWriter, r *http.Request) {
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

func (s Server) handleGet(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		return http.StatusBadRequest, errors.New("missing 'hash' parameter")
	}
	if hash == "all" {
		files := s.c.GetAll()
		respond(w, &files)
	} else {
		v, err := s.c.Get(hash)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		respond(w, &v)
	}
	return http.StatusOK, nil
}

func (s Server) handlePost(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	var params ezt.PostParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not decode body: %v", err.Error())
	}
	for i := range params.Files {
		s.c.Add(params.Files[i].Hash, params.Files[i].IFile, params.Addr)
	}
	return http.StatusOK, nil
}

func (s Server) handleDelete(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		return http.StatusBadRequest, errors.New("missing 'hash' parameter")
	}
	addr := r.URL.Query().Get("addr")
	if hash == "" {
		return http.StatusBadRequest, errors.New("missing 'addr' parameter")
	}
	if err := s.c.Del(hash, addr); err != nil {
		return http.StatusInternalServerError, err
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
