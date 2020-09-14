package cli

import "fmt"

type Cmd struct {
	Name    string
	Desc    string
	Handler func(args ...string) error
}

type CLI struct {
	name string
	cmds []Cmd
}

func New(name string, cmds []Cmd) CLI {
	c := CLI{name: name}
	c.cmds = append(c.cmds, Cmd{
		Name:    "help",
		Desc:    "Print help message",
		Handler: c.usage,
	})
	c.cmds = append(c.cmds, cmds...)
	return c
}

func (c CLI) Handle(name string, args []string) error {
	cmd, err := c.find(name)
	if err != nil {
		return err
	}
	return cmd.Handler(args...)
}

func (c CLI) Usage() {
	fmt.Printf("Usage: %s command\n", c.name)
	fmt.Println("Commands:")
	for _, cmd := range c.cmds {
		fmt.Printf("\t%s\t\t%s\n", cmd.Name, cmd.Desc)
	}
}

func (c CLI) find(name string) (Cmd, error) {
	for _, cmd := range c.cmds {
		if cmd.Name == name {
			return cmd, nil
		}
	}
	return Cmd{}, fmt.Errorf("unknownd command '%s'", name)
}

func (c *CLI) usage(args ...string) error {
	c.Usage()
	return nil
}
