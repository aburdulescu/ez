package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aburdulescu/ez/ezt"

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
	go func() {
		log.Println(http.ListenAndServe(":23232", nil))
	}()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	ln, err := net.Listen("tcp", ":23231")
	if err != nil {
		return err
	}
	opts := badger.DefaultOptions(dbPath).WithLogger(nil).WithReadOnly(true).WithBypassLockGuard(true)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	if err := updateTracker(db); err != nil {
		return err
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

func updateTracker(db *badger.DB) error {
	var files []ezt.File
	err := db.View(func(txn *badger.Txn) error {
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
	rsp, err := http.Post(cfg.TrackerURL, "application/json", buf)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func handleCtrlC(c chan os.Signal, db *badger.DB) {
	<-c
	db.Close()
	os.Exit(0)
}
