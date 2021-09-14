package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
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
var disableLog bool

var trackerURL string
var logger Logger

func run() error {
	flag.StringVar(&dbPath, "dbpath", "./seeder.db", "path where the database is stored")
	flag.StringVar(&seedAddr, "seedaddr", "", "address to be used by peers")
	flag.StringVar(&trackerAddr, "trackeraddr", "", "tracker address")
	flag.BoolVar(&disableLog, "disable-log", false, "disable logging")
	flag.Parse()

	if seedAddr == "" {
		return fmt.Errorf("seedaddr is empty")
	}
	if trackerAddr == "" {
		return fmt.Errorf("trackeraddr is empty")
	}

	if disableLog {
		logger = &NopLogger{
			log.New(os.Stderr, "", 0),
		}
	} else {
		logger = log.New(os.Stderr, "", log.Lshortfile|log.Ltime|log.Lmicroseconds|log.LUTC)
	}

	seedAddr = seedAddr + ":22201"
	trackerURL = "http://" + trackerAddr + ":22200" + "/"

	db, err := NewDB(dbPath)
	if err != nil {
		return err
	}

	seederServer, err := NewSeederServer(db)
	if err != nil {
		return err
	}
	go seederServer.ListenAndServe()

	fileWatcher, err := NewWatcher(db)
	if err != nil {
		return err
	}
	go fileWatcher.Run()

	go NewLocalServer(db, fileWatcher.Channel()).Run()

	trackerProbeServer, err := NewTrackerProbeServer("239.23.23.0:22203", db)
	if err != nil {
		return err
	}
	go trackerProbeServer.ListenAndServe()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	waitForSignal(c, db)

	return nil
}

func waitForSignal(c chan os.Signal, db *DB) {
	<-c
	db.Close()
}
