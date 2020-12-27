package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aburdulescu/ez/cadet"
	"github.com/aburdulescu/ez/cmn"
	"github.com/aburdulescu/ez/ezt"
)

var (
	root = &cadet.Command{
		Use:   "ezl",
		Short: "Manage seeder file database",
	}
)
var commands = []*cadet.Command{
	&cadet.Command{
		Use:   "ls",
		Short: "List files",
		Run:   onLs,
	},
	&cadet.Command{
		Use:   "add",
		Short: "Add a file",
		Run:   onAdd,
	},
	&cadet.Command{
		Use:   "rm",
		Short: "Remove a file",
		Run:   onRm,
	},
	&cadet.Command{
		Use:   "sync",
		Short: "Send all files to the tracker",
		Run:   onSync,
	},
}

func main() {
	if err := root.AddCommand(commands...); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	root.Execute()
}

func init() {

}

func onLs(args []string) error {
	rsp, err := http.Get("http://localhost:22202/list")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	var files []ezt.File
	if err := json.NewDecoder(rsp.Body).Decode(&files); err != nil {
		return err
	}
	p := cmn.NewPrinter()
	defer p.Flush()
	p.Printf("ID\tPath\tSize\n")
	for _, f := range files {
		p.Printf("%s\t%s\t%d\n", f.Id, filepath.Join(f.IFile.Dir, f.IFile.Name), f.IFile.Size)
	}
	return nil
}

func onAdd(args []string) error {
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

func onRm(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
	rsp, err := http.Get("http://localhost:22202/rm?id=" + id)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}

func onSync(args []string) error {
	rsp, err := http.Get("http://localhost:22202/sync")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}
