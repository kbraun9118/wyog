package cmd

import (
	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use: "rm",
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

		return nil
	},
}
