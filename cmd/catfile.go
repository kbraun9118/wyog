package cmd

import (
	"fmt"
	"slices"

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
		repoPath, err := repository.FindRequire(".")
		if err != nil {
			return err
		}
		repo, err := repository.New(repoPath)
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
	sha, err := repository.ObjectFind(repo, object, format)
	if err != nil {
		return err
	}
	obj, err := repository.Read(repo, sha)
	if err != nil {
		return err
	}
	fmt.Print(string(obj.Serialize()))

	return nil
}
