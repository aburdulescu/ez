package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/aburdulescu/ez/ezt"
	badger "github.com/dgraph-io/badger/v2"
)

const (
	maxDatagramSize = 8192
)

type HandlerFunc func(*net.UDPConn, *net.UDPAddr, []byte)

type TrackerProbeServer struct {
	conn *net.UDPConn
	db   *badger.DB
}

func NewTrackerProbeServer(addr string, db *badger.DB) (TrackerProbeServer, error) {
	a, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return TrackerProbeServer{}, err
	}
	c, err := net.ListenMulticastUDP("udp4", nil, a) // TODO: don't use nil for interface
	if err != nil {
		return TrackerProbeServer{}, err
	}
	s := TrackerProbeServer{conn: c, db: db}
	if err := s.updateTracker(); err != nil {
		log.Println(err)
	}
	return s, nil
}

func (s TrackerProbeServer) ListenAndServe() {
	s.conn.SetReadBuffer(maxDatagramSize)
	for {
		b := make([]byte, maxDatagramSize)
		if _, err := s.conn.Read(b); err != nil {
			log.Println(err)
			return
		}
		if err := s.updateTracker(); err != nil {
			log.Println(err)
			return
		}
	}
}

func (s TrackerProbeServer) updateTracker() error {
	var files []ezt.File
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
					id := strings.Split(kstr, ".")[0]
					files = append(files, ezt.File{Hash: id, IFile: i})
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	params := ezt.PostParams{
		Files: files,
		Addr:  seedAddr,
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
