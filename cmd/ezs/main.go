package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	badger "github.com/dgraph-io/badger/v2"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	ln, err := net.Listen("tcp", ":23231")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	opts := badger.DefaultOptions("../ezl/db").WithLogger(nil).WithReadOnly(true).WithBypassLockGuard(true)
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
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
