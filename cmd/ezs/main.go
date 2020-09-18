package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	badger "github.com/dgraph-io/badger/v2"
)

type Config struct {
	ListenAddr string `json:"listenaddr"`
	DBPath     string `json:"dbpath"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "path to the configuration file")
	flag.Parse()
	f, err := os.Open(configPath)
	if err != nil {
		handleErr(err)
	}
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		handleErr(err)
	}
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		handleErr(err)
	}
	opts := badger.DefaultOptions(cfg.DBPath).WithLogger(nil).WithReadOnly(true).WithBypassLockGuard(true)
	db, err := badger.Open(opts)
	if err != nil {
		handleErr(err)
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

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func handleCtrlC(c chan os.Signal, db *badger.DB) {
	<-c
	db.Close()
	os.Exit(0)
}
