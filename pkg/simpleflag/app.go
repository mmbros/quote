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
	Name          string
	Usage         string
	Commands      []*Command
	Writer        io.Writer
	ErrorHandling flag.ErrorHandling

	// Command that will be executed.
	// Setted by Parse (if no error is returned).
	execCmdCommand *Command
}

func (app *App) CommandName() string {
	if app.execCmdCommand == nil {
		return ""
	}
	return app.execCmdCommand.Name()
}

// Command is ...
type Command struct {
	// Names contains the various names of the command,
	// separated by a comma (",") with no spaces.
	// The first name is the main name of the command.
	// Eventually, the other name are aliases.
	Names   string
	Usage   string
	Options []*Option

	Run func() error
}

// Option is ...
type Option struct {
	Value flag.Value
	Names string
}

// Name returns the first name of the command
func (cmd *Command) Name() string {
	return strings.SplitN(cmd.Names, ",", 2)[0]
}

// Output is ...
func (app *App) Output() io.Writer {
	if app.Writer == nil {
		return os.Stderr
	}
	return app.Writer
}

func (app *App) findCommandByName(name string) *Command {
	for _, cmd := range app.Commands {
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
func (app *App) usageFailf(format string, v ...interface{}) error {
	out := app.Output()
	err := fmt.Errorf(format, v...)
	fmt.Fprintln(out, err)
	fmt.Fprintln(out, app.Usage)

	switch app.ErrorHandling {
	case flag.PanicOnError:
		panic(err)
	case flag.ExitOnError:
		os.Exit(2)
	}

	return err
}

// FlagSet returns a FlagSet based on command options
func (cmd *Command) FlagSet(out io.Writer) *flag.FlagSet {
	name := cmd.Name()
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(out)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), cmd.Usage)
	}
	// populate FlagSet variables with command options
	for _, opt := range cmd.Options {
		for _, name := range strings.Split(opt.Names, ",") {
			fs.Var(opt.Value, name, "")
		}
	}
	return fs
}

// Parse parses flag definitions from the argument list,
// which should not include the command name.
// Must be called after all flags in the FlagSet are defined
// and before flags are accessed by the program.
func (app *App) Parse(arguments []string) error {
	// reset the requested command
	app.execCmdCommand = nil

	if arguments == nil || len(arguments) == 0 {
		return app.usageFailf("no arguments")
	}
	out := app.Output()

	initFlagSet := func(name, usage string) *flag.FlagSet {
		fs := flag.NewFlagSet(name, app.ErrorHandling)
		fs.SetOutput(out)
		fs.Usage = func() {
			fmt.Fprintln(out, usage)
		}
		return fs
	}

	cmdName := arguments[0]
	if strings.HasPrefix(cmdName, "-") {
		// TODO: make app like command interface and use (cmd *Command) FlagSet
		fs := initFlagSet("", app.Usage)
		return fs.Parse(arguments)
	}

	cmd := app.findCommandByName(cmdName)
	if cmd == nil {
		return app.usageFailf("unknown command %q", cmdName)
	}

	fs := cmd.FlagSet(out)

	err := fs.Parse(arguments[1:])

	if err == nil {
		// save the requested command
		app.execCmdCommand = cmd
	}
	return err
}

func (app *App) Run() error {

	if app.execCmdCommand == nil {
		return fmt.Errorf("nothing to do: no command has successfully be parsed")
	}
	if app.execCmdCommand.Run == nil {
		return fmt.Errorf("nothing to do: %q command has Run function undefined", app.execCmdCommand.Name())
	}
	return app.execCmdCommand.Run()
}

func (app *App) GetCommand(name string) *Command {
	for _, cmd := range app.Commands {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
