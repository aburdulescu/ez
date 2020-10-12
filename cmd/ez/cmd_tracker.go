package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

func onTracker(cmd *cobra.Command, args []string) error {
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
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getTrackerURL() (string, error) {
	tracker, err := getTracker()
	if err != nil {
		return "", err
	}
	trackerURL := "http://" + tracker + ":22200"
	return trackerURL, nil
}
