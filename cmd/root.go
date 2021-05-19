package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"stageai.tech/sunshine/sunshine/config"

	"github.com/spf13/cobra"
)

const fmtConn = "host='%s' port='%d' dbname='%s' sslmode=disable binary_parameters=yes"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sunshine",
	Short: "Backend of the Latvian Energy Efficiency Multi-stakeholder platform",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func pgConnectString(cfg config.DB) string {
	var v = fmt.Sprintf(fmtConn, cfg.Host, cfg.Port, cfg.Name)
	if cfg.Username != "" {
		v = fmt.Sprintf("%s user='%s'", v, cfg.Username)
	}
	if cfg.Password != "" {
		v = fmt.Sprintf("%s password='%s'", v, cfg.Password)
	}
	return v
}

func mustRunCommand(args ...string) {
	var cmd = exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Running %q failed", args[0])
	}
}
