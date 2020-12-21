package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aburdulescu/ez/ezt"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ezl",
		Short: "Manage seeder file database",
	}
	lsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List files",
		RunE:  onLs,
	}
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a file",
		RunE:  onAdd,
	}
	rmCmd = &cobra.Command{
		Use:   "rm",
		Short: "Remove a file",
		RunE:  onRm,
	}
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Send all files to the tracker",
		RunE:  onSync,
	}
)

func init() {
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(syncCmd)
}

func onLs(cmd *cobra.Command, args []string) error {
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
		fmt.Printf("%s\t%s/%s\t%d\n", f.Id, f.IFile.Dir, f.IFile.Name, f.IFile.Size)
	}
	return nil
}

func onAdd(cmd *cobra.Command, args []string) error {
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

func onRm(cmd *cobra.Command, args []string) error {
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

func onSync(cmd *cobra.Command, args []string) error {
	rsp, err := http.Get("http://localhost:22202/sync")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	return nil
}
