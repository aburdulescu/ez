package main

import (
	"fmt"
	"log"
	"net"
	"os"

	badger "github.com/dgraph-io/badger/v2"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	db, err := badger.Open(badger.DefaultOptions("../ezl/db").WithLogger(nil))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		c := Client{
			conn: conn,
		}
		go c.run(db)
	}
}
