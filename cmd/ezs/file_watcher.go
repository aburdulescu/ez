package main

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
	db      *DB
}

func NewWatcher(db *DB) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	files, err := db.GetAll()
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(file.ifile.Dir, file.ifile.Name)
		if err := w.Add(path); err != nil {
			return err
		}
	}
	return &Watcher{w, db}, nil
}

func (w Watcher) Add(path string) error {
	if err := w.watcher.Add(path); err != nil {
		return err
	}
	return nil
}

func (w Watcher) Run() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			log.Println("event:", event)
			switch {
			case (event.Op & (fsnotify.Remove | fsnotify.Rename)) != 0:
				log.Println("remove file from db")
			case (event.Op & fsnotify.Write) != 0:
				log.Println("remove file, process it and add it to db")
			default:
				log.Println("do nothing")
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
