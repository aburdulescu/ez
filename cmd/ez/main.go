package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aburdulescu/ez/cli"
	"github.com/aburdulescu/ez/ezt"
	// "github.com/pkg/profile"
)

type Config struct {
	TrackerURL string `json:"trackerUrl"`
}

var c = cli.New(os.Args[0], []cli.Cmd{
	cli.Cmd{
		Name:    "ls",
		Desc:    "List available files",
		Handler: onLs,
	},
	cli.Cmd{
		Name:    "get",
		Desc:    "Download a file",
		Handler: onGet,
	},
})

var cfg Config

func main() {
	// defer profile.Start(profile.ProfilePath("."), profile.CPUProfile).Stop()
	// defer profile.Start(profile.ProfilePath("."), profile.MemProfile).Stop()
	f, err := os.Open("ez.json")
	if err != nil {
		handleErr(err)
	}
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		handleErr(err)
	}
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
	args := os.Args[1:]
	if len(args) < 1 {
		handleErr(fmt.Errorf("command not provided"))
	}
	name := args[0]
	args = args[1:]
	if err := c.Handle(name, args); err != nil {
		handleErr(err)
	}
}

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	c.Usage()
	os.Exit(1)
}

func onLs(args ...string) error {
	rsp, err := http.Get(cfg.TrackerURL + "?hash=all")
	if err != nil {
		log.Println(err)
		return err
	}
	defer rsp.Body.Close()
	var files []ezt.GetAllResult
	if err := json.NewDecoder(rsp.Body).Decode(&files); err != nil {
		log.Println(err)
		return err
	}
	for _, f := range files {
		fmt.Printf("%s\t\t%s\t\t%d\n", f.Hash, f.Name, f.Size)
	}
	return nil
}

func onGet(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
	rsp, err := http.Get(cfg.TrackerURL + "?hash=" + id)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rsp.Body.Close()
	r := ezt.GetResult{}
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		log.Println(err)
		return err
	}
	var d Downloader
	if err := d.Run(id, r.IFile, r.Peers); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
