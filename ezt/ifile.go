package ezt

import (
	"fmt"
	"os"
	"path/filepath"
)

type IFile struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
	Size int64  `json:"size"`
}

func NewIFile(f *os.File, path string) (IFile, error) {
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
	ifile := IFile{
		Name: fi.Name(),
		Size: fi.Size(),
		Dir:  abspath,
	}
	return ifile, nil
}

func (l IFile) Equals(r IFile) bool {
	return (l.Name == r.Name && l.Size == r.Size && l.Dir == r.Dir)
}
