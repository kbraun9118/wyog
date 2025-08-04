package cmd

import (
	"fmt"
	"slices"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

func init() {
	revParseCmd.Flags().StringVarP(&objType, "type", "t", "", "Specify the expected type")
}

var (
	objType     string
	revParseCmd = &cobra.Command{
		Use:   "rev-parse name",
		Short: "Parse revision (or other objects) identifiers",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(objType) == 0 {
				return nil
			}
			if !slices.Contains([]string{"blob", "commit", "tag", "tree"}, objType) {
				return fmt.Errorf("type must be one of [blob, commit, tag, tree]")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			path, err := repository.FindRequire(".")
			if err != nil {
				return err
			}

			repo, err := repository.New(path)
			if err != nil {
				return err
			}

			obj, err := repository.ObjectFind(&repo, name, objType)

			fmt.Println(obj)

			return nil
		},
	}
)
