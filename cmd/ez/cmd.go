package main

import (
	"github.com/aburdulescu/ez/cadet"
)

var (
	root = &cadet.Command{
		Use:   "ez",
		Short: "Easy to use p2p file transfer tool for your local network",
	}

	commands = []*cadet.Command{
		&cadet.Command{
			Use:   "ls",
			Short: "List files",
			Run:   onLs,
		},
		&cadet.Command{
			Use:   "get",
			Short: "Download a file",
			Run:   onGet,
		},
		&cadet.Command{
			Use:   "tracker",
			Short: "Set/get tracker address",
			Run:   onTracker,
		},
	}
)
