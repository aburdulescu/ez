package main

import (
	"log"
	"net"

	"github.com/aburdulescu/ez/ezt"
)

const (
	maxDatagramSize = 8192
)

type HandlerFunc func(*net.UDPConn, *net.UDPAddr, []byte)

type TrackerProbeServer struct {
	conn *net.UDPConn
	db   DB
}

func NewTrackerProbeServer(addr string, db DB) (TrackerProbeServer, error) {
	a, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return TrackerProbeServer{}, err
	}
	c, err := net.ListenMulticastUDP("udp4", nil, a) // TODO: don't use nil for interface
	if err != nil {
		return TrackerProbeServer{}, err
	}
	s := TrackerProbeServer{conn: c, db: db}
	if err := updateTracker(s.db); err != nil {
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
		log.Println("tracker sent probe")
		if err := updateTracker(s.db); err != nil {
			log.Println(err)
			return
		}
	}
}

// TODO: first get the file from the tracker, compare with local ones and send the diff to tracker
func updateTracker(db DB) error {
	files, err := db.GetAll()
	if err != nil {
		return err
	}
	trackerClient := ezt.NewClient(trackerURL)
	req := ezt.AddRequest{
		Files: files,
		Addr:  seedAddr,
	}
	if err := trackerClient.Add(req); err != nil {
		return err
	}
	return nil
}
