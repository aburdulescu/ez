package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	badger "github.com/dgraph-io/badger/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var dbPath string
var seedAddr string
var trackerAddr string
var trackerURL string

func run() error {
	flag.StringVar(&dbPath, "dbpath", "./db", "path where the database is stored")
	flag.StringVar(&seedAddr, "seedaddr", "", "address to used by peers")
	flag.StringVar(&trackerAddr, "trackeraddr", "", "tracker address")
	flag.Parse()

	if seedAddr == "" {
		return fmt.Errorf("seedaddr is empty")
	}
	if trackerAddr == "" {
		return fmt.Errorf("trackeraddr is empty")
	}

	seedAddr = seedAddr + ":22201"
	trackerURL = "http://" + trackerAddr + ":22200" + "/"

	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)

	ln, err := net.Listen("tcp", ":22201")
	if err != nil {
		return err
	}

	opts := badger.DefaultOptions(dbPath).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}

	trackerProbeServer, err := NewTrackerProbeServer("239.23.23.0:22203", db)
	if err != nil {
		return err
	}
	go trackerProbeServer.ListenAndServe()

	go NewLocalServer(db).Run()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(c, db)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		c := Client{
			db:   db,
			conn: conn,
		}
		go c.run()
	}
}

func handleCtrlC(c chan os.Signal, db *badger.DB) {
	<-c
	db.Close()
	os.Exit(0)
}
