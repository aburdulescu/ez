package ezt

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type File struct {
	Hash  string `json:"hash"`
	IFile IFile  `json:"ifile"`
}

type AddRequest struct {
	Files []File `json:"files"`
	Addr  string `json:"addr"`
}

type RemoveRequest struct {
	Id   string `json:"id"`
	Addr string `josn:"addr"`
}

type GetAllItem struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type GetAllResponse struct {
	Files []GetAllItem `json:"files"`
}

type GetRequest struct {
	Id string `json:"id"`
}

type GetResponse struct {
	IFile IFile    `json:"ifile"`
	Peers []string `json:"peers"`
}

type Client struct {
	url string
}

func NewClient(url string) Client {
	return Client{url}
}

func (c Client) GetAll() (GetAllResponse, error) {
	rsp, err := http.Get(c.url + "?hash=all")
	if err != nil {
		log.Println(err)
		return GetAllResponse{}, err
	}
	defer rsp.Body.Close()
	var files []GetAllItem
	if err := json.NewDecoder(rsp.Body).Decode(&files); err != nil {
		log.Println(err)
		return GetAllResponse{}, err
	}
	return GetAllResponse{files}, nil
}

func (c Client) Get(req GetRequest) (GetResponse, error) {
	rsp, err := http.Get(c.url + "?hash=" + req.Id)
	if err != nil {
		log.Println(err)
		return GetResponse{}, err
	}
	defer rsp.Body.Close()
	var r GetResponse
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		log.Println(err)
		return GetResponse{}, err
	}
	return r, nil
}

func (c Client) Add(req AddRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&req); err != nil {
		return err
	}
	rsp, err := http.Post(c.url, "application/json", &buf)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func (c Client) Remove(req RemoveRequest) error {
	query := "?hash=" + req.Id + "&addr=" + req.Addr
	r, err := http.NewRequest("DELETE", c.url+query, nil)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	rsp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}
