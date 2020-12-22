package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/aburdulescu/ez/ezt"
)

type Server struct {
	c *KV
}

func main() {
	s := Server{NewKV()}
	http.HandleFunc("/", s.handleRequest)
	go sendProbeToSeeders()
	log.Fatal(http.ListenAndServe(":22200", nil))
}

func sendProbeToSeeders() {
	time.Sleep(200 * time.Millisecond)
	c, err := NewTrackerProbeClient("239.23.23.0:22203")
	if err != nil {
		log.Println(err)
	}
	b := []byte("hello")
	if _, err := c.Write(b); err != nil {
		log.Println(err)
	}
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
	id := r.URL.Query().Get("id")
	if id == "" {
		return http.StatusBadRequest, errors.New("missing 'id' parameter")
	}
	if id == "all" {
		files := s.c.GetAll()
		respond(w, &files)
	} else {
		v, err := s.c.Get(id)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		respond(w, &v)
	}
	return http.StatusOK, nil
}

func (s Server) handlePost(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	var req ezt.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not decode body: %v", err.Error())
	}
	log.Println("post data:", req)
	for i := range req.Files {
		s.c.Add(req.Files[i].Id, req.Files[i].IFile, req.Addr)
	}
	return http.StatusOK, nil
}

func (s Server) handleDelete(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()
	id := r.URL.Query().Get("id")
	if id == "" {
		return http.StatusBadRequest, errors.New("missing 'id' parameter")
	}
	addr := r.URL.Query().Get("addr")
	if id == "" {
		return http.StatusBadRequest, errors.New("missing 'addr' parameter")
	}
	if err := s.c.Del(id, addr); err != nil {
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
