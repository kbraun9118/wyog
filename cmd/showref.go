package cmd

import (
	"fmt"

	"github.com/kbraun9118/wyog/ref"
	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var showRefCmd = &cobra.Command{
	Use:   "show-ref",
	Short: "List references",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := repository.FindRequire("")
		if err != nil {
			return err
		}

		repo, err := repository.New(path)
		if err != nil {
			return err
		}

		refs, err := ref.List(&repo, nil)
		if err != nil {
			return err
		}

		if err := showRef(&repo, refs, true, "refs"); err != nil {
			return err
		}

		return nil
	},
}

func showRef(repo *repository.Repository, refs map[string]any, withHash bool, prefix string) error {
	if len(refs) > 0 {
		prefix = prefix + "/"
	}

	for k, v := range refs {
		switch v := v.(type) {
		case string:
			if withHash {
				fmt.Printf("%s %s%s\n", v, prefix, k)
			} else {
				fmt.Printf("%s%s\n", prefix, k)
			}
		case map[string]any:
			err := showRef(repo, v, withHash, prefix+k)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid type")
		}
	}

	return nil
}
