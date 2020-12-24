package main

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
	db      *DB
	fileIds map[string]string
}

func NewWatcher(db *DB) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		watcher: watcher,
		db:      db,
		fileIds: make(map[string]string),
	}
	files, err := db.GetAll()
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		path := filepath.Join(file.IFile.Dir, file.IFile.Name)
		if err := w.Add(path, file.Id); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func (w Watcher) Add(path, id string) error {
	if err := w.watcher.Add(path); err != nil {
		return err
	}
	w.fileIds[path] = id
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
				id, ok := w.fileIds[event.Name]
				if !ok {
					log.Printf("%s not found", event.Name)
					continue
				}
				if err := w.db.Delete(id); err != nil {
					log.Println(err)
					continue
				}
			case (event.Op & fsnotify.Write) != 0:
				log.Println("remove file, process it and add it to db")
			default:
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}
