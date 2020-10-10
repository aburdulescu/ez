package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/aburdulescu/ez/cli"
	"github.com/aburdulescu/ez/ezt"
	// "github.com/pkg/profile"
)

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
	cli.Cmd{
		Name:    "tracker",
		Desc:    "Set tracker address",
		Handler: onTracker,
	},
})

func main() {
	// defer profile.Start(profile.ProfilePath("."), profile.CPUProfile).Stop()
	// defer profile.Start(profile.ProfilePath("."), profile.MemProfile, profile.MemProfileRate(1)).Stop()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)
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

func setTracker(addr string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	path := filepath.Join(usr.HomeDir, ".ez")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(addr); err != nil {
		return err
	}
	return nil
}

func getTracker() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	f, err := os.Open(filepath.Join(usr.HomeDir, ".ez"))
	if os.IsNotExist(err) {
		return "", fmt.Errorf("tracker address is not set")
	}
	if err != nil {
		return "", err
	}
	defer f.Close()
	b := make([]byte, 256)
	n, err := f.Read(b)
	if err != nil {
		return "", err
	}
	return string(b[:n]), nil
}

func getTrackerURL() (string, error) {
	tracker, err := getTracker()
	if err != nil {
		return "", err
	}
	trackerURL := "http://" + tracker + ":22200"
	return trackerURL, nil
}

func onLs(args ...string) error {
	trackerURL, err := getTrackerURL()
	if err != nil {
		return err
	}
	rsp, err := http.Get(trackerURL + "?hash=all")
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
	trackerURL, err := getTrackerURL()
	if err != nil {
		return err
	}
	rsp, err := http.Get(trackerURL + "?hash=" + id)
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

func onTracker(args ...string) error {
	if len(args) < 1 {
		tracker, err := getTracker()
		if err != nil {
			return err
		}
		fmt.Println(tracker)
	} else {
		addr := args[0]
		if err := setTracker(addr); err != nil {
			return err
		}
	}
	return nil
}
