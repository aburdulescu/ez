package main

import (
	"fmt"
	"os"

	"github.com/aburdulescu/go-ez/cli"
)

var c = cli.New(os.Args[0], []cli.Cmd{
	cli.Cmd{
		Name:    "ls",
		Desc:    "List available files",
		Handler: onLs,
	},
	cli.Cmd{
		Name:    "get",
		Desc:    "Download a file",
		Handler: onGet,
	},
})

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		handleErr(fmt.Errorf("command not provided"))
	}
	name := args[0]
	args = args[1:]
	if err := c.Handle(name, args); err != nil {
		handleErr(err)
	}
}

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	c.Usage()
	os.Exit(1)
}

func onLs(args ...string) error {
	return nil
}

func onGet(args ...string) error {
	return nil
}
