package cmd

import (
	"fmt"

	"github.com/mmbros/quote/internal/quote"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// chars tha can be used for separate sorurce name and workers
	sepsSourceWorkers = "/:#"

	defaultWorkers = 1

	nameIsin    = "isins"
	nameSource  = "sources"
	nameWorkers = "workers"
	nameProxy   = "proxy"
	// nameTor      = "tor"
	nameDryRun   = "dry-run"
	nameDatabase = "database"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use: "get",
	//	Use:   "get <isin1> [<isin2> ...]",
	Short: "Get the quotes of the specified isins",
	Long: `Get the quotes of the specified isins from the sources.
If source options are not specified, all the available sources are used.
See 'quote sources' for a list of the available sources.
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// build cmdGetArgs
		fgs := cmd.Flags()
		a := &cmdGetArgs{
			PassedDatabase: fgs.Changed(nameDatabase),
			PassedWorkers:  fgs.Changed(nameWorkers),
			PassedProxy:    fgs.Changed(nameProxy),
		}
		a.DryRun, _ = fgs.GetBool(nameDryRun)
		a.Database, _ = fgs.GetString(nameDatabase)
		a.Workers, _ = fgs.GetInt(nameWorkers)
		a.Proxy, _ = fgs.GetString(nameProxy)
		a.Isins, _ = fgs.GetStringSlice(nameIsin)
		a.Sources, _ = fgs.GetStringSlice(nameSource)

		cfg, err := getConfig(a, quote.Sources())
		sis := cfg.getSourceIsinsList()

		if a.DryRun {
			fmt.Printf("ARGS: %v\n", a)

			configFile := viper.GetViper().ConfigFileUsed()
			fmt.Printf("Using configuration file %q\n", configFile)
			fmt.Printf("Database: %q\n", cfg.Database)

			fmt.Println(jsonString(sis))

			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}

			return nil
		}

		// do retrieves the quotes
		if err == nil {
			err = quote.Get(sis, cfg.Database)
		}
		return err
	},

	Example: `    quote get -i isin1,isin2 -s sourceA/4,sourceB, -s sourceC --workers 2
  retrieves 2 isins from 3 sources: A with 4 workers, B and C with 2 workers each.`,
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	flgs := getCmd.Flags()

	flgs.BoolP(nameDryRun, "n", false, "perform a trial run with no request/updates made")
	// flgs.Bool(nameTor, false, "must use Tor network")
	flgs.StringSliceP(nameSource, "s", nil, "list of sources to get the quotes from")
	flgs.IntP(nameWorkers, "w", defaultWorkers, "number of workers")
	flgs.StringSliceP(nameIsin, "i", nil, "list of isins to get the quotes")
	// getCmd.Flags().StringP("output", "o", "", "output file")

	// commented because MarkFlagRequired doesn't check config file
	// cobra.MarkFlagRequired(flgs, nameIsin)

	// XXX commented
	// viper.BindPFlags(flgs)
}
