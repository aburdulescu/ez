package main

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

const (
	AddOp uint8 = iota
	RmOp
)

type WatcherEntry struct {
	Op   uint8
	Path string
	Id   string
}

type Watcher struct {
	watcher  *fsnotify.Watcher
	db       *DB
	fileIds  map[string]string
	incoming chan WatcherEntry
}

func NewWatcher(db *DB) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		watcher:  watcher,
		db:       db,
		fileIds:  make(map[string]string),
		incoming: make(chan WatcherEntry),
	}
	files, err := db.GetFiles()
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		path := filepath.Join(file.IFile.Dir, file.IFile.Name)
		if err := w.add(path, file.Id); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func (w Watcher) Channel() chan<- WatcherEntry {
	return w.incoming
}

func (w Watcher) add(path, id string) error {
	if err := w.watcher.Add(path); err != nil {
		return err
	}
	w.fileIds[path] = id
	return nil
}

func (w Watcher) remove(path string) error {
	if err := w.watcher.Remove(path); err != nil {
		return err
	}
	delete(w.fileIds, path)
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
				if err := RemoveFile(w.db, id); err != nil {
					log.Println(err)
					continue
				}
				if err := w.remove(event.Name); err != nil {
					log.Println(err)
					continue
				}
			case (event.Op & (fsnotify.Write | fsnotify.Chmod)) != 0:
				id, ok := w.fileIds[event.Name]
				if !ok {
					log.Printf("%s not found", event.Name)
					continue
				}
				if err := RemoveFile(w.db, id); err != nil {
					log.Println(err)
					continue
				}
				if err := w.remove(event.Name); err != nil {
					log.Println(err)
					continue
				}
				id, err := AddFile(w.db, event.Name)
				if err != nil {
					log.Println(err)
					continue
				}
				if err := w.add(event.Name, id); err != nil {
					log.Println(err)
					continue
				}
			default:
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		case e := <-w.incoming:
			log.Println(e)
			switch e.Op {
			case AddOp:
				if err := w.add(e.Path, e.Id); err != nil {
					log.Println(err)
					continue
				}
			case RmOp:
				if err := w.remove(e.Id); err != nil {
					log.Println(err)
					continue
				}
			default:
			}
		}
	}
}
