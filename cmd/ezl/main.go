package main

import (
	"fmt"
	"os"
)

var cli = NewCLI(os.Args[0], map[string]Cmd{
	"ls": Cmd{
		desc:    "List available files",
		handler: onLs,
	},
	"add": Cmd{
		desc:    "Add a file",
		handler: onAdd,
	},
	"rm": Cmd{
		desc:    "Remove a file",
		handler: onRm,
	},
})

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		handleErr(fmt.Errorf("command not provided"))
	}
	name := args[0]
	args = args[1:]
	if err := cli.Handle(name, args); err != nil {
		handleErr(err)
	}
}

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func onLs(args ...string) error {
	fmt.Println("ls")
	return nil
}

func onAdd(args ...string) error {
	fmt.Println("add")
	return nil
}

func onRm(args ...string) error {
	fmt.Println("rm")
	return nil
}
