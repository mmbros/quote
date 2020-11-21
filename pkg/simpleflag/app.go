//
// Package simpleflag is useful for creating command line Go applications.
//
// Limitations
//
// No arguments are managed, only flags.
// The App must have subcommands.
//
// Configuration
//
// App is the main structure of the cli application.
// The App has a list of Commands.
// Each command has a list of Flags.
// Each flag has a flag.Value and comma separated alternatives names.
//
// Flag of type Bool, Int, String and Strings are defined.
// Bool, Int, String have the Passed field, indicating if the flag was
// setted in the command line.
//
// Example
//
// Example of a configuration of simple "myapp" cli application,
// with a single "get" command.
//
//    type myappArgs struct {
//        config   simpleflag.String
//        workers  simpleflag.Int
//        dryrun   simpleflag.Bool
//        items    simpleflag.Strings
//    }
//
//    args := myappArgs{}
//
//    app := &simpleflag.App{
//        Name:     "myapp",
//        Usage:    "myapp <command>",
//        Commands: []*simpleflag.Command{
//            &simpleflag.Command{
//                Names: "get,g",
//                Usage: "myapp get [options]",
//                Flags: []*simpleflag.Flag{
//                    {Value: &args.config, Names: "c,config"},
//                    {Value: &args.workers, Names: "w,workers"},
//                    {Value: &args.dryrun, Names: "n,dryrun,dry-run"},
//                    {Value: &args.items, Names: "i,items"},
//                },
//            },
//        },
//    }
//
// Usage
//
// First App.Parse function parses the arguments list.
//
// Then the App.CommandName method returns the name of the command invoked.
package simpleflag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// App is the main structure of a cli application.
// The App has a list of Commands.
// Each command has a list of Flags.
type App struct {
	Name          string // name of the program
	Usage         string // printed as it is, without further manipulations
	Commands      []*Command
	Writer        io.Writer // nil means stderr; use Output() accessor
	ErrorHandling flag.ErrorHandling

	// Command invoked by command line arguments.
	// Setted by Parse (if no error is returned).
	invoked *Command
}

// Command represents an application (sub-)command.
type Command struct {
	// Names contains the various names of the command,
	// separated by a comma (",") with no spaces.
	// The first name is the main name of the command.
	// The other names, if present, are aliases.
	Names string

	// Usage string of the Command.
	// It is printed as it is, without further manipulations.
	Usage string

	// Flags of the command.
	Flags []*Flag
}

// A Flag represents the state of a flag.
type Flag struct {
	Value flag.Value // value as set
	Names string     // comma separated aliases of the flag
}

// CommandName returns the Name of the command invoked in the command line.
// It returns an empty string in case no command was selected.
func (app *App) CommandName() string {
	if app.invoked == nil {
		return ""
	}
	return app.invoked.Name()
}

// Name returns the first name of the command.
// It is the main name of the command, returned by App.CommandName().
func (cmd *Command) Name() string {
	return strings.SplitN(cmd.Names, ",", 2)[0]
}

// Output returns the destination for usage and error messages.
// os.Stderr is returned if output was not set or was set to nil.
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

// usageFailf prints to app.Output a formatted error and usage message and
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

// Parse parses flag definitions from the argument list
// which should not include the command name.
// Must be called after all flags in the FlagSet are defined
// and before flags are accessed by the program.
func (app *App) Parse(arguments []string) error {
	// reset the requested command
	app.invoked = nil

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
		app.invoked = cmd
	}
	return err
}

// FlagSet returns a *flag.FlagSet based on command Flags.
// Adds a flag.Flag for each name of each simpleflag.Flag.
func (cmd *Command) FlagSet(out io.Writer) *flag.FlagSet {
	name := cmd.Name()
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(out)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), cmd.Usage)
	}
	// populate FlagSet variables with command options
	for _, opt := range cmd.Flags {
		for _, name := range strings.Split(opt.Names, ",") {
			fs.Var(opt.Value, name, "")
		}
	}
	return fs
}
