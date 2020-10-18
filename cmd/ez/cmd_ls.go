package main

import (
	"fmt"
	"log"

	"github.com/aburdulescu/ez/ezt"
	"github.com/spf13/cobra"
)

func onLs(cmd *cobra.Command, args []string) error {
	trackerURL, err := getTrackerURL()
	if err != nil {
		return err
	}
	trackerClient := ezt.NewClient(trackerURL)
	rsp, err := trackerClient.GetAll()
	if err != nil {
		log.Println(err)
		return err
	}
	for _, f := range rsp.Files {
		fmt.Printf("%s\t\t%s\t\t%d\n", f.Hash, f.Name, f.Size)
	}
	return nil
}
