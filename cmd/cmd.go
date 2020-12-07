package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mmbros/quote/internal/quote"
	"github.com/mmbros/quote/pkg/simpleflag"
)

const (
	defaultConfigType = "yaml"
	defaultMode       = "1"
)

type appArgs struct {
	config     simpleflag.String
	configType simpleflag.String
	database   simpleflag.String
	dryrun     simpleflag.Bool
	isins      simpleflag.Strings
	proxy      simpleflag.String
	sources    simpleflag.Strings
	workers    simpleflag.Int
	mode       simpleflag.String
}

const (
	usageApp = `Usage:
    quote <command> [options]

Available Commands:
    get      Get the quotes of the specified isins
    sources  Show available sources
    tor      Checks if Tor network will be used
`
	usageGet = `Usage:
    quote get [options]

Options:
    -c, --config      path     config file (default is $HOME/.quote.yaml)
        --config-type string   used if config file does not have the extension in the name;
                               accepted values are: YAML, TOML and JSON 
    -i, --isins       strings  list of isins to get the quotes
    -n, --dry-run              perform a trial run with no request/updates made
    -p, --proxy       url      default proxy
    -s, --sources     strings  list of sources to get the quotes from
    -w, --workers     int      number of workers (default 1)
    -d, --database    dns      sqlite3 database used to save the quotes
    -m, --mode        char     result mode: "1" first success or last error (default)
                                            "U" all errors until first success 
                                            "A" all 
`

	usageTor = `Usage:
     quote tor [options]

Checks if Tor network will be used to get the quote.

To use the Tor network the proxy must be defined through:
	1. proxy argument parameter
	2. proxy config file parameter
	3. HTTP_PROXY, HTTPS_PROXY and NOPROXY enviroment variables.

Options:
    -c, --config      path    config file (default is $HOME/.quote.yaml)
	    --config-type string  used if config file does not have the extension in the name;
	                          accepted values are: YAML, TOML and JSON 
    -p, --proxy       url     proxy to test the Tor network
`

	usageSources = `Usage:
	quote sources

Prints list of available sources.
`
)

func initCommandGet(args *appArgs) *simpleflag.Command {

	flags := []*simpleflag.Flag{
		{Value: &args.config, Names: "c,config"},
		{Value: &args.configType, Names: "config-type"},
		{Value: &args.database, Names: "d,database"},
		{Value: &args.dryrun, Names: "n,dryrun,dry-run"},
		{Value: &args.isins, Names: "i,isins"},
		{Value: &args.proxy, Names: "p,proxy"},
		{Value: &args.sources, Names: "s,sources"},
		{Value: &args.workers, Names: "w,workers"},
		{Value: &args.mode, Names: "m,mode"},
	}

	cmd := &simpleflag.Command{
		Names: "get,g",
		Usage: usageGet,
		Flags: flags,
	}
	return cmd
}

func initCommandTor(args *appArgs) *simpleflag.Command {

	flags := []*simpleflag.Flag{
		{Value: &args.config, Names: "c,config"},
		{Value: &args.configType, Names: "config-ype"},
		{Value: &args.proxy, Names: "p,proxy"},
	}

	cmd := &simpleflag.Command{
		Names: "tor,t",
		Usage: usageTor,
		Flags: flags,
	}
	return cmd
}

func initCommandSources(args *appArgs) *simpleflag.Command {

	cmd := &simpleflag.Command{
		Names: "sources,s",
		Usage: usageSources,
	}
	return cmd
}

func initApp(args *appArgs) *simpleflag.App {

	app := &simpleflag.App{
		ErrorHandling: flag.ExitOnError,
		Name:          "quote",
		Usage:         usageApp,
		Commands: []*simpleflag.Command{
			initCommandGet(args),
			initCommandTor(args),
			initCommandSources(args),
		},
	}
	return app
}

func execTor(args *appArgs, cfg *Config) error {
	if args.config.Passed {
		fmt.Printf("Using configuration file %q\n", args.config.Value)
	}
	proxy := cfg.Proxy
	// proxy = "x://\\"
	fmt.Printf("Checking Tor connection with proxy %q\n", proxy)
	_, msg, err := quote.TorCheck(proxy)
	if err == nil {
		// ok checking Tor network:
		// prints the result: it can be ok or ko
		fmt.Println(msg)
	}
	return err
}

func execGet(args *appArgs, cfg *Config) error {

	sis := cfg.SourceIsinsList()

	if args.dryrun.Value {
		// fmt.Printf("ARGS: %v\n", args)
		if args.config.Passed {
			fmt.Printf("Using configuration file %q\n", args.config.Value)
		}

		if cfg.Database != "" {
			fmt.Printf("Database: %q\n", cfg.Database)
		}
		// if cfg.Mode != "" {
		// 	fmt.Printf("Mode: %q\n", cfg.Mode)
		// }
		fmt.Printf("Mode: %q (%d)\n", cfg.Mode, cfg.mode)
		fmt.Println("Tasks:", jsonString(sis))

		return nil
	}

	// do retrieves the quotes
	return quote.Get(sis, cfg.Database, cfg.mode)
}

func execSources(args *appArgs, cfg *Config) error {
	sources := quote.Sources()
	fmt.Printf("Available sources: \"%s\"\n", strings.Join(sources, "\", \""))
	return nil
}

// Execute is the main function
func Execute() {

	arguments := os.Args[1:]
	// arguments := strings.Split("get -c user.quote.yaml -s morningstarit:33", " ")

	var (
		app  *simpleflag.App
		cfg  *Config // TODO rename in appConfig
		args *appArgs
	)

	args = &appArgs{}

	app = initApp(args)

	// NOTE: if app.Parse ha success, as a side effect
	// the args struct is initialized
	err := app.Parse(arguments)

	// get configuration
	if err == nil {
		cfg, err = GetConfig(args, quote.Sources())
	}

	if err == nil {
		switch app.CommandName() {
		case "get":
			err = execGet(args, cfg)
		case "tor":
			err = execTor(args, cfg)
		case "sources":
			err = execSources(args, cfg)
		}
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
