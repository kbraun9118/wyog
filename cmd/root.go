package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wyog",
	Short: "An implementation of a subset of git commands",
	Long: `This is a personal project intended to learn more
about the internals of git by creating a toy 
implementation of git`,
	Version: "0.0.1",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		addCmd,
		catFileCmd,
		checkIgnoreCmd,
		checkoutCmd,
		commitCmd,
		hashObjectCmd,
		initCmd,
		logCmd,
		lsFilesCmd,
		lsTreeCmd,
		revParseCmd,
		rmCmd,
		showRefCmd,
		statusCmd,
		tagCmd,
	)
}
