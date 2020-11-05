package cmd

import (
	"fmt"
	"os"

	"github.com/mmbros/quote/internal/quote"
	"github.com/spf13/cobra"
)

// checkTorCmd represents the sources command
var checkTorCmd = &cobra.Command{
	Use:   "tor-check",
	Short: "Checks if Tor network will be used",
	Long: `Checks if the quotes are retrieved through the Tor network.

To use the Tor network the proxy must be defined through 
HTTP_PROXY, HTTPS_PROXY and NOPROXY enviroment variables.
`,
	Aliases: []string{"tor"},
	Run: func(cmd *cobra.Command, args []string) {
		proxy, _ := cmd.Flags().GetString(nameProxy)
		_, msg, err := quote.TorCheck(proxy)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(2)
		}
		fmt.Println(msg)
		return
	},
	Example: "  HTTPS_PROXY=<tor-proxy> quote check-tor",
}

func init() {
	rootCmd.AddCommand(checkTorCmd)

	checkTorCmd.Flags().String(nameProxy, "p", "proxy to test the TOR connection")
}
