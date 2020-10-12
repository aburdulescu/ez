package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ez",
		Short: "Easy to use p2p file transfer tool for your local network",
	}
	lsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List files",
		RunE:  onLs,
	}
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Download a file",
		Args:  cobra.MinimumNArgs(1),
		RunE:  onGet,
	}
	trackerCmd = &cobra.Command{
		Use:   "tracker",
		Short: "Set/get tracker address",
		RunE:  onTracker,
	}
)

func init() {
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(trackerCmd)
}
