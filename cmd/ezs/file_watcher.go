package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
}

func NewWatcher() (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{w}, nil
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
				log.Println("process file and add it to db")
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
