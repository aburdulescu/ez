package main

import (
	"github.com/aburdulescu/ez/cadet"
)

func main() {
	root := &cadet.Command{
		Use:   "cadet",
		Short: "Short desc of the tool.",
	}

	add := &cadet.Command{
		Use:     "add filepath",
		Short:   "Add a file.",
		Example: "cadet add foo/bar.txt",
		Run:     onAdd,
	}

	rm := &cadet.Command{
		Use:     "rm filepath",
		Short:   "Remove a file.",
		Example: "cadet rm foo/bar.txt",
		Run:     onAdd,
	}

	root.AddCommand(add, rm)

	root.Execute()
}

func onAdd(cmd *cadet.Command, args []string) error {
	return nil
}
