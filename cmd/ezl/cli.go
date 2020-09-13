package main

import "fmt"

type Cmd struct {
	desc    string
	handler func(args ...string) error
}

type CLI struct {
	name string
	cmds map[string]Cmd
}

func NewCLI(name string, cmds map[string]Cmd) CLI {
	c := CLI{name: name, cmds: cmds}
	c.cmds["help"] = Cmd{
		desc:    "Print this message",
		handler: c.usage,
	}
	return c
}

func (c CLI) Handle(name string, args []string) error {
	cmd, ok := c.cmds[name]
	if !ok {
		return fmt.Errorf("unknownd command '%s'", name)
	}
	return cmd.handler(args...)
}

func (c CLI) Usage() {
	fmt.Printf("%s COMMAND", c.name)
	fmt.Println("Commands:")
}

func (c CLI) usage(args ...string) error {
	c.Usage()
	return nil
}
