package main

import (
	"fmt"
	"log"

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
		Run:     onRm,
	}

	if err := root.AddCommand(add, rm); err != nil {
		log.Fatal(err)
	}

	root.Execute()
}

func onAdd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing filepath")
	}
	fmt.Println("add:", args[0])
	return nil
}

func onRm(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing filepath")
	}
	fmt.Println("remove:", args[0])
	return nil
}
