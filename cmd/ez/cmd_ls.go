package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aburdulescu/ez/ezt"
	"github.com/spf13/cobra"
)

func onLs(cmd *cobra.Command, args []string) error {
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
