package main

import (
	"log"

	"github.com/aburdulescu/ez/cmn"
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
	p := cmn.NewPrinter()
	defer p.Flush()
	p.Printf("ID\tFilename\tSize\n")
	for _, f := range rsp.Files {
		p.Printf("%s\t%s\t%d\n", f.Id, f.Name, f.Size)
	}
	return nil
}
