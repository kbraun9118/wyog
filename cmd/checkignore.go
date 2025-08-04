package cmd

import (
	"fmt"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var checkIgnoreCmd = &cobra.Command{
	Use:   "check-ignore",
	Short: "Check path(s) against ignore rules.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := repository.FindRequire(".")
		if err != nil {
			return err
		}

		repo, err := repository.New(p)
		if err != nil {
			return err
		}

		rules, err := repo.ReadIgnore()
		for _, path := range args {
			if CheckIgnore(rules, path) {
				fmt.Println(path)
			}
		}
		return nil
	},
}
