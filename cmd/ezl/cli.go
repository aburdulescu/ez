package main

import "fmt"

type Cmd struct {
	name    string
	desc    string
	handler func(args ...string) error
}

type CLI struct {
	name string
	cmds []Cmd
}

func NewCLI(name string, cmds []Cmd) CLI {
	c := CLI{name: name}
	c.cmds = append(c.cmds, Cmd{
		name:    "help",
		desc:    "Print help message",
		handler: c.usage,
	})
	c.cmds = append(c.cmds, cmds...)
	return c
}

func (c CLI) Handle(name string, args []string) error {
	cmd, err := c.find(name)
	if err != nil {
		return err
	}
	return cmd.handler(args...)
}

func (c CLI) Usage() {
	fmt.Printf("Usage: %s command\n", c.name)
	fmt.Println("Commands:")
	for _, cmd := range c.cmds {
		fmt.Printf("\t%s\t\t%s\n", cmd.name, cmd.desc)
	}
}

func (c CLI) find(name string) (Cmd, error) {
	for _, cmd := range c.cmds {
		if cmd.name == name {
			return cmd, nil
		}
	}
	return Cmd{}, fmt.Errorf("unknownd command '%s'", name)
}

func (c *CLI) usage(args ...string) error {
	c.Usage()
	return nil
}
