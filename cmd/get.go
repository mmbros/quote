package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mmbros/quote/internal/quote"
	"github.com/spf13/cobra"
)

const (
	// chars tha can be used for separate sorurce name and workers
	sepsSourceWorkers = "/:#"
	optDefaultWorkers = 1
)

var (
	optTor bool
	// optDryRun        bool
	optWorkers       int
	optSourceWorkers []string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <isin1> [<isin2> ...]",
	Short: "Get the quotes of the specified isins",
	Long: `Get the quotes of the specified isins from the sources.
If source options are not specified, all the available sources are used.
See 'quote sources' for a list of the available sources.
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		sources, workers, err := parseSources(optSourceWorkers, optWorkers)
		if err != nil {
			return err
		}

		// handle --tor flag
		if optTor {
			ok, _, err := quote.TorCheck()
			if !ok && err == nil {
				err = fmt.Errorf("Tor network not available")
			}
			if err != nil {
				return nil
			}
		}

		return quote.Get(args, sources, workers)
	},

	Example: `    quote get isin1 isin2 -s sourceA/4,sourceB, -s sourceC --workers 2
  retrieves isins from 3 sources A with 4 workers, B and C with 2 workers each.`,

	Args: cobra.MinimumNArgs(1),
}

func init() {
	//getCmd.SetUsageTemplate("SetUsageTemplate")

	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// getCmd.Flags().BoolVarP(&optDryRun, "dry-run", "n", false, "perform a trial run with no request/updates made")
	getCmd.Flags().BoolVar(&optTor, "tor", false, "must use Tor network")

	getCmd.Flags().StringSliceVarP(&optSourceWorkers, "source", "s", nil, "list of sources to get the the quotes")

	getCmd.Flags().IntVarP(&optWorkers, "workers", "w", 1, "number of workers")
	// getCmd.Flags().StringP("output", "o", "", "output file")
}

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

func parseSources(sourceWorkers []string, defWorkers int) (sources []string, workers []int, err error) {
	L := len(sourceWorkers)
	if L == 0 {
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
