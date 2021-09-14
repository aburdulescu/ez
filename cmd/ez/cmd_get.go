package main

import (
	"fmt"
	"log"

	"github.com/aburdulescu/ez/ezt"
)

var disableGetProgressBar bool = false

func onGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	var id string
	if len(args) > 1 {
		switch args[0] {
		case "--no-progress":
			disableGetProgressBar = true
		default:
			return fmt.Errorf("unknown flag %s", args[0])
		}
		id = args[1]
	} else {
		id = args[0]
	}
	trackerURL, err := getTrackerURL()
	if err != nil {
		return err
	}
	trackerClient := ezt.NewClient(trackerURL)
	rsp, err := trackerClient.Get(ezt.GetRequest{id})
	if err != nil {
		log.Println(err)
		return err
	}
	var d Downloader
	if err := d.Run(id, rsp.IFile, rsp.Peers); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
