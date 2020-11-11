package simpleflag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// App is ...
type App struct {
	Name     string
	Usage    string
	Commands []*Command

	Writer        io.Writer
	ErrorHandling flag.ErrorHandling
}

// Command is ...
type Command struct {
	Names   string
	Usage   string
	Options []*Option
}

// Option is ...
type Option struct {
	Value flag.Value
	Names string
}

// Output is ...
func (a *App) Output() io.Writer {
	if a.Writer == nil {
		return os.Stderr
	}
	return a.Writer
}

func (a *App) findCommandByName(name string) *Command {
	for _, cmd := range a.Commands {
		for _, n := range strings.Split(cmd.Names, ",") {
			if n == name {
				return cmd
			}
		}
	}
	return nil
}

// failf prints to app.Output a formatted error and usage message and
// returns the error.
func (a *App) usageFailf(format string, v ...interface{}) error {
	out := a.Output()
	err := fmt.Errorf(format, v...)
	fmt.Fprintln(out, err)
	fmt.Fprintln(out, a.Usage)

	switch a.ErrorHandling {
	case flag.PanicOnError:
		panic(err)
	case flag.ExitOnError:
		os.Exit(2)
	}

	return err
}

// Parse parses flag definitions from the argument list,
// which should not include the command name.
// Must be called after all flags in the FlagSet are defined
// and before flags are accessed by the program.
func (a *App) Parse(arguments []string) error {
	if arguments == nil || len(arguments) == 0 {
		return a.usageFailf("no arguments")
	}
	out := a.Output()

	initFlagSet := func(name, usage string) *flag.FlagSet {
		fs := flag.NewFlagSet(name, a.ErrorHandling)
		fs.SetOutput(out)
		fs.Usage = func() {
			fmt.Fprintln(out, usage)
		}
		return fs
	}

	cmdName := arguments[0]
	if strings.HasPrefix(cmdName, "-") {
		fs := initFlagSet("", a.Usage)
		return fs.Parse(arguments)
	}

	cmd := a.findCommandByName(cmdName)
	if cmd == nil {
		return a.usageFailf("unknown command %q", cmdName)
	}

	fs := initFlagSet(cmdName, cmd.Usage)

	// populate FlagSet variables with command options
	for _, opt := range cmd.Options {
		for _, name := range strings.Split(opt.Names, ",") {
			fs.Var(opt.Value, name, "")
		}
	}

	return fs.Parse(arguments[1:])
}
