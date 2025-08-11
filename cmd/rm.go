package cmd

import (
	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm paths...",
	Short: "Remove files from the working tree and the index.",
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

		index, err := repo.ReadIndex()
		if err != nil {
			return err
		}

		if err := repo.WriteIndex(index); err != nil {
			return err
		}

		if err := repo.Rm(true, false, args...); err != nil {
			return err
		}

		return nil
	},
}
