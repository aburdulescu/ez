package main

import (
	"os"

	"github.com/aburdulescu/ez/ezt"
)

func AddFile(db *DB, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	i, err := ezt.NewIFile(f, path)
	if err != nil {
		return "", err
	}
	checksums, err := ProcessFile(f, i.Size)
	if err != nil {
		return "", err
	}
	id := NewID(checksums)
	if err != nil {
		return "", err
	}
	if err := db.Add(id, i, checksums); err != nil {
		return "", err
	}
	trackerClient := ezt.NewClient(trackerURL)
	req := ezt.AddRequest{
		Files: []ezt.File{
			ezt.File{Id: id, IFile: i},
		},
		Addr: seedAddr,
	}
	if err := trackerClient.Add(req); err != nil {
		logger.Println(err)
	}
	return id, nil
}

func RemoveFile(db *DB, id string) error {
	if err := db.Delete(id); err != nil {
		return err
	}
	trackerClient := ezt.NewClient(trackerURL)
	req := ezt.RemoveRequest{
		Id: id, Addr: seedAddr,
	}
	if err := trackerClient.Remove(req); err != nil {
		logger.Println(err)
	}
	return nil
}
