package cmd

import (
	"fmt"

	"stageai.tech/sunshine/sunshine"

	"github.com/spf13/cobra"
)

// versionCmd represents the migrate command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(sunshine.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
