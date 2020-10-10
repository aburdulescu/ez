package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aburdulescu/ez/cli"
	"github.com/aburdulescu/ez/ezt"
)

var c = cli.New(os.Args[0], []cli.Cmd{
	cli.Cmd{
		Name:    "ls",
		Desc:    "List available files",
		Handler: onLs,
	},
	cli.Cmd{
		Name:    "add",
		Desc:    "Add a file",
		Handler: onAdd,
	},
	cli.Cmd{
		Name:    "rm",
		Desc:    "Remove a file",
		Handler: onRm,
	},
	cli.Cmd{
		Name:    "sync",
		Desc:    "Send all available files to the tracker",
		Handler: onSync,
	},
})

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) < 1 {
		return fmt.Errorf("command not provided")
	}
	name := args[0]
	args = args[1:]
	if err := c.Handle(name, args); err != nil {
		return err
	}
	return nil
}
func onLs(args ...string) error {
	rsp, err := http.Get("http://localhost:22202/list")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	var files []ezt.File
	if err := json.NewDecoder(rsp.Body).Decode(&files); err != nil {
		return err
	}
	for _, f := range files {
		fmt.Printf("%s\t%s/%s\t%d\n", f.Hash, f.IFile.Dir, f.IFile.Name, f.IFile.Size)
	}
	return nil
}

func onAdd(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("filepath wasn't provided")
	}
	path := args[0]
	if !filepath.IsAbs(path) {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = filepath.Join(pwd, path)
	}
	rsp, err := http.Get("http://localhost:22202/add?path=" + path)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func onRm(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
	rsp, err := http.Get("http://localhost:22202/rm?hash=" + id)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func onSync(args ...string) error {
	rsp, err := http.Get("http://localhost:22202/sync")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}
