package cmd

import (
	"strings"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{"up"}
		}
		migrate(config.Load(), args...)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func migrate(cfg config.Config, args ...string) {
	mustRunCommand(
		"goose",
		"-dir",
		cfg.Paths.Migrations,
		"postgres",
		pgConnectString(cfg.DB),
		strings.Join(args, " "),
	)
}
