package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/gorilla/securecookie"
	"github.com/spf13/cobra"
)

var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate session keys",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("auth = \"%s\"\nencr = \"%s\"\n",
			base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)),
			base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)),
		)
	},
}

func init() {
	rootCmd.AddCommand(genkeyCmd)
}
