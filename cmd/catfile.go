package cmd

import (
	"fmt"
	"slices"

	"github.com/kbraun9118/wyog/gitobject"
	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var catFileCmd = &cobra.Command{
	Use:   "cat-file type object",
	Short: "Provide content of repository objects",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if !slices.Contains([]string{"blob", "commit", "tag", "tree"}, args[0]) {
			return fmt.Errorf("type is not one of [blob, commit, tag, tree]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := repository.Find(".")
		if repoPath == nil {
			return fmt.Errorf("not a git repository (or any of the parent directories)")
		}
		repo, err := repository.New(*repoPath)
		if err != nil {
			return err
		}
		if err := catFile(&repo, args[1], args[0]); err != nil {
			return err
		}
		return nil
	},
}

func catFile(repo *repository.Repository, object string, format string) error {
	obj, err := gitobject.Read(repo, gitobject.Find(repo, object, format))
	if err != nil {
		return err
	}
	fmt.Print(string(obj.Serialize()))

	return nil
}
