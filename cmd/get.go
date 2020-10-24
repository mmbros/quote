package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mmbros/quote/internal/quote"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// chars tha can be used for separate sorurce name and workers
	sepsSourceWorkers = "/:#"

	nameIsin    = "isins"
	nameSource  = "sources"
	nameWorkers = "workers"
	nameTor     = "tor"
	nameDruRun  = "dry-run"
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
		vip := viper.GetViper()

		// get the parameters of the get function
		isins := vip.GetStringSlice(nameIsin)
		defaultWorkers := vip.GetInt(nameWorkers)
		torIsMandatory := vip.GetBool(nameTor)
		sources, workers, err := parseSources(
			vip.GetStringSlice(nameSource),
			defaultWorkers,
		)
		if err != nil {
			return err
		}
		database := vip.GetString("database")
		proxy := vip.GetString("proxy")

		// handle --dry-run flag
		if vip.GetBool(nameDruRun) {
			fmt.Printf(`quote get:
  isins:          %v
  sources:        %v
  workers:        %v
  defaultWorkers: %d
  tor:            %v
  proxy:          %s
  database:       %s
`, isins, sources, workers, defaultWorkers, torIsMandatory, proxy, database)

			return nil
		}

		// handle --tor flag
		if torIsMandatory {
			ok, _, err := quote.TorCheck()
			if !ok && err == nil {
				err = fmt.Errorf("Tor network not available")
			}
			if err != nil {
				return err
			}
		}

		// retrieves the quotes
		return quote.Get(isins, sources, workers, database)
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

	flgs.BoolP(nameDruRun, "n", false, "perform a trial run with no request/updates made")
	flgs.Bool(nameTor, false, "must use Tor network")
	flgs.StringSliceP(nameSource, "s", nil, "list of sources to get the the quotes")
	flgs.IntP(nameWorkers, "w", 1, "number of workers")
	flgs.StringSliceP(nameIsin, "i", nil, "list of isins to get the the quotes")
	// getCmd.Flags().StringP("output", "o", "", "output file")

	// commented because MarkFlagRequired doesn't check config file
	// cobra.MarkFlagRequired(flgs, nameIsin)

	viper.BindPFlags(flgs)
}

// splitSourceWorkers gets the sourceWorkers string
// and returns the two component source and workers.
// The components must be separated by one of the sepsSourceWorkers chars.
// If no separator char is found,
// retuns the input string as source and defWorkers as workers.
func splitSourceWorkers(sourceWorkers string, defWorkers int) (source string, workers int, err error) {
	idx := strings.IndexAny(sourceWorkers, sepsSourceWorkers)
	if idx < 0 {
		source = sourceWorkers
		workers = defWorkers
	} else if idx == 0 || idx == len(sourceWorkers)-1 {
		goto labelReturnError
	} else {
		source = sourceWorkers[:idx]
		sw := sourceWorkers[idx+1:]
		workers, err = strconv.Atoi(sw)
		if err != nil || workers <= 0 {
			goto labelReturnError
		}
	}

	source = strings.TrimSpace(source)
	// fmt.Printf("%s -> (%s, %d)\n", sourceWorkers, source, workers)
	return

labelReturnError:
	err = fmt.Errorf("Invalid source: %q", sourceWorkers)
	// fmt.Printf("%s -> ERR %v\n", sourceWorkers, err)
	return
}

// parseSources parse all the input string and return the corresponding sources and workers array.
// In case no source is given, returns all the available sources as sources,
// and an array with one element equal to defWorkers as workers:
// the same number of workers will be used for all the sources
func parseSources(sourceWorkers []string, defWorkers int) (sources []string, workers []int, err error) {
	L := len(sourceWorkers)
	if L == 0 {
		sources = quote.Sources()
		workers = []int{defWorkers}
		return
	}
	sources = make([]string, L)
	workers = make([]int, L)

	for j, sw := range sourceWorkers {
		sources[j], workers[j], err = splitSourceWorkers(sw, defWorkers)
		if err != nil {
			return nil, nil, err
		}
	}

	return
}
