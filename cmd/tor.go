package cmd

import (
	"fmt"
	"os"

	"github.com/mmbros/quote/internal/quote"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// checkTorCmd represents the sources command
var checkTorCmd = &cobra.Command{
	Use:   "tor",
	Short: "Checks if Tor network will be used",
	Long: `Checks if the quotes are retrieved through the Tor network.

To use the Tor network the proxy must be defined through:
  1. --proxy argument parameter
  2. proxy config file parameter
  3. HTTP_PROXY, HTTPS_PROXY and NOPROXY enviroment variables.
`,
	Run: func(cmd *cobra.Command, args []string) {
		proxy := getProxy(cmd)

		configFile := viper.GetViper().ConfigFileUsed()
		fmt.Printf("Using configuration file %q\n", configFile)
		fmt.Printf("Checking Tor connection with proxy %q\n", proxy)
		_, msg, err := quote.TorCheck(proxy)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(2)
		}
		fmt.Println(msg)
		return
	},
	Example: "  HTTPS_PROXY=<tor-proxy> quote tor",
}

func getProxy(cmd *cobra.Command) string {
	// vip := viper.GetViper()
	fgs := cmd.Flags()
	a := &cmdGetArgs{
		PassedProxy: fgs.Changed(nameProxy),
	}
	a.Proxy, _ = fgs.GetString(nameProxy)
	// a.Proxy = vip.GetString(nameProxy)

	allSources := quote.Sources()
	cfg, _ := getFullNotValidatedConfig(a, allSources)
	if cfg != nil {
		return cfg.Proxy
	}
	return ""
}

func init() {
	rootCmd.AddCommand(checkTorCmd)

	flgs := checkTorCmd.Flags()

	flgs.StringP(nameProxy, "p", "", "proxy to test the TOR connection")

	// cobra.MarkFlagRequired(flgs, nameProxy)
}
