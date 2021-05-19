package cmd

import (
	"fmt"
	"log"
	"os"

	"stageai.tech/sunshine/sunshine/openapi"

	"github.com/spf13/cobra"
)

var (
	openapiSource string
	openapiOutput string
)

var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Concatenate OpenAPI files",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)

		f, err := os.OpenFile(openapiOutput, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(fmt.Errorf("open %s: %v", openapiOutput, err))
		}
		defer f.Close()

		if err := openapi.Concatenate(f, openapiSource); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	openapiCmd.Flags().StringVarP(&openapiSource, "source", "s", "./openapi", "directory with files to read")
	openapiCmd.Flags().StringVarP(&openapiOutput, "output", "o", "openapi.json", "output file")
	rootCmd.AddCommand(openapiCmd)
}
