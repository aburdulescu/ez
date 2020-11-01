package main

import (
	"fmt"
	"log"

	"github.com/aburdulescu/ez/ezt"
	"github.com/spf13/cobra"
)

func onGet(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("id wasn't provided")
	}
	id := args[0]
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
