package main

import (
	"fmt"
	"os"
)

var cli = NewCLI(os.Args[0], []Cmd{
	Cmd{
		name:    "ls",
		desc:    "List available files",
		handler: onLs,
	},
	Cmd{
		name:    "add",
		desc:    "Add a file",
		handler: onAdd,
	},
	Cmd{
		name:    "rm",
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
	cli.Usage()
	os.Exit(1)
}

func onLs(args ...string) error {
	fmt.Println("ls")
	return nil
}

func onAdd(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("filepath wasn't provided")
	}
	path := args[0]
	i, err := NewIFile(path)
	if err != nil {
		return err
	}
	fmt.Println(i)
	return nil
}

func onRm(args ...string) error {
	fmt.Println("rm")
	return nil
}
