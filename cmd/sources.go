package cmd

import (
	"github.com/spf13/cobra"
)

// sourcesCmd represents the sources command
var sourcesCmd = &cobra.Command{
	Use:     "sources",
	Short:   "Show available sources",
	Long:    `Show available sources.`,
	Aliases: []string{"list"},
	Run: func(cmd *cobra.Command, args []string) {
		//		fmt.Println(sources.Names())
	},
}

func init() {
	rootCmd.AddCommand(sourcesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sourcesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sourcesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
