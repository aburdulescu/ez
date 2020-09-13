package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type IFile struct {
	Hash string
	Name string
	Dir  string
	Size int64
}

func NewIFile(path string) (IFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return IFile{}, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return IFile{}, err
	}
	if fi.IsDir() {
		return IFile{}, fmt.Errorf("%s is a directory", path)
	}
	var abspath string
	if filepath.IsAbs(path) {
		abspath = filepath.Dir(path)
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			return IFile{}, err
		}
		abspath = filepath.Join(pwd, filepath.Dir(path))
	}
	chunks, err := ChunksFromFile(f, fi.Size())
	if err != nil {
		return IFile{}, err
	}
	h, err := NewHash(chunks)
	if err != nil {
		return IFile{}, err
	}
	ifile := IFile{
		Name: fi.Name(),
		Size: fi.Size(),
		Dir:  abspath,
		Hash: h.String(),
	}
	return ifile, nil
}
