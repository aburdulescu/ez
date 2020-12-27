package cadet

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

type Command struct {
	// Use is the one-line usage message.
	// Recommended syntax is as follow:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Use string

	// Short is the short description shown in the 'help' output.
	Short string

	// Example is examples of how to use the command.
	Example string

	// Run: Typically the actual work function. Most commands will only implement this.
	Run func(cmd *Command, args []string) error

	// commands is the list of commands supported by this program.
	commands []*Command
	// parent is a parent command for this command.
	parent *Command

	// args is actual args parsed from flags.
	args []string

	// Max lengths of commands' string lengths for use in padding.
	commandsMaxNameLen int
}

// Root finds root command.
func (c *Command) Root() *Command {
	if c.HasParent() {
		return c.Parent().Root()
	}
	return c
}

// Execute uses the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches for commands.
func (c *Command) Execute() {
	// Regardless of what command execute is called on, run on Root only
	if c.HasParent() {
		c.Root().Execute()
		return
	}

	cmd, args, err := c.Find(os.Args[1:])
	if err != nil {
		// If found parse to a subcommand and then failed, talk about the subcommand
		if cmd != nil {
			c = cmd
		}
		c.HandlerErr(err)
		return
	}

	if !cmd.Runnable() {
		cmd.Usage()
		return
	}

	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help", "help":
			cmd.Usage()
			return
		}
	}

	if err = cmd.Run(cmd, args); err != nil {
		cmd.HandlerErr(err)
		return
	}
}

// AddCommand adds one or more commands to this parent command.
func (c *Command) AddCommand(cmds ...*Command) {
	for i, x := range cmds {
		if cmds[i] == c {
			panic("Command can't be a child of itself")
		}
		cmds[i].parent = c
		nameLen := len(x.Name())
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
		c.commands = append(c.commands, x)
	}
}

// Commands returns a slice of child commands.
func (c *Command) Commands() []*Command {
	return c.commands
}

func (c *Command) findNext(next string) *Command {
	for _, cmd := range c.commands {
		if cmd.Name() == next {
			return cmd
		}
	}
	return nil
}

// argsMinusFirstX removes only the first x from args.  Otherwise, commands that look like
// openshift admin policy add-role-to-user admin my-user, lose the admin argument (arg[4]).
func argsMinusFirstX(args []string, x string) []string {
	for i, y := range args {
		if x == y {
			ret := []string{}
			ret = append(ret, args[:i]...)
			ret = append(ret, args[i+1:]...)
			return ret
		}
	}
	return args
}

// Find the target command given the args and command tree
// Meant to be run on the highest node. Only searches down.
func (c *Command) Find(args []string) (*Command, []string, error) {
	var innerfind func(*Command, []string) (*Command, []string)

	innerfind = func(c *Command, innerArgs []string) (*Command, []string) {
		if len(innerArgs) == 0 {
			return c, innerArgs
		}
		nextSubCmd := innerArgs[0]
		cmd := c.findNext(nextSubCmd)
		if cmd != nil {
			return innerfind(cmd, argsMinusFirstX(innerArgs, nextSubCmd))
		}
		return c, innerArgs
	}

	commandFound, a := innerfind(c, args)
	if commandFound == c && len(args) > 0 {
		return c, a, fmt.Errorf("unknown command %q for %q", args[0], c.CommandPath())
	}
	return commandFound, a, nil
}

// CommandPath returns the full path to this command.
func (c *Command) CommandPath() string {
	if c.HasParent() {
		return c.Parent().CommandPath() + " " + c.Name()
	}
	return c.Name()
}

// UseLine puts out the full usage for a given command (including parents).
func (c *Command) UseLine() string {
	var useline string
	if c.HasParent() {
		useline = c.parent.CommandPath() + " " + c.Use
	} else {
		useline = c.Use
	}
	return useline
}

// Name returns the command's name: the first word in the use line.
func (c *Command) Name() string {
	name := c.Use
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// HasExample determines if the command has example.
func (c *Command) HasExample() bool {
	return len(c.Example) > 0
}

// Runnable determines if the command is itself runnable.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// HasSubCommands determines if the command has children commands.
func (c *Command) HasSubCommands() bool {
	return len(c.commands) > 0
}

// IsAvailableCommand determines if a command is available as a non-help command
func (c *Command) IsAvailableCommand() bool {
	if c.Runnable() || c.HasSubCommands() {
		return true
	}
	return false
}

// HasAvailableSubCommands determines if a command has available sub commands that
// need to be shown in the usage/help default template under 'available commands'.
func (c *Command) HasAvailableSubCommands() bool {
	// return true on the first found available sub command
	for _, sub := range c.commands {
		if sub.IsAvailableCommand() {
			return true
		}
	}
	// the command has no sub commands
	return false
}

// HasParent determines if the command is a child command.
func (c *Command) HasParent() bool {
	return c.parent != nil
}

// Parent returns a commands parent command.
func (c *Command) Parent() *Command {
	return c.parent
}

var out = os.Stderr

func print(i ...interface{}) {
	fmt.Fprint(out, i...)
}

func println(i ...interface{}) {
	print(fmt.Sprintln(i...))
}

func printf(format string, i ...interface{}) {
	print(fmt.Sprintf(format, i...))
}

func (c *Command) HandlerErr(err error) {
	println("Error:", err.Error())
	printf("Run '%v --help' for usage.\n", c.CommandPath())
}

func (c *Command) UsageTemplate() string {
	return `{{.Short}}

Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

func (c *Command) UsageFunc() (f func(*Command) error) {
	if c.HasParent() {
		return c.Parent().UsageFunc()
	}
	return func(c *Command) error {
		err := tmpl(os.Stderr, c.UsageTemplate(), c)
		if err != nil {
			println(err)
		}
		return err
	}
}

// Usage puts out the usage for the command.
// Used when a user provides invalid input.
func (c *Command) Usage() error {
	return c.UsageFunc()(c)
}

var minNamePadding = 11

// NamePadding returns padding for the name.
func (c *Command) NamePadding() int {
	if c.parent == nil || minNamePadding > c.parent.commandsMaxNameLen {
		return minNamePadding
	}
	return c.parent.commandsMaxNameLen
}

var templateFuncs = template.FuncMap{
	"rpad": rpad,
}

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) error {
	t := template.New("top")
	t.Funcs(templateFuncs)
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}
