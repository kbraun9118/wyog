package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout commit path",
	Short: "Checkout a commit inside of a directory.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := repository.FindRequire(".")
		if err != nil {
			return err
		}

		repo, err := repository.New(path)
		if err != nil {
			return err
		}

		commit, path := args[0], args[1]

		sha, err := repository.ObjectFind(&repo, commit, "")
		if err != nil {
			return err
		}
		obj, err := repository.ReadObj(&repo, sha)
		if err != nil {
			return err
		}

		if commitObj, ok := obj.(*repository.Commit); ok {
			tree, ok := commitObj.Kvlm.Headers.Get("tree")
			if !ok {
				return fmt.Errorf("cannot find tree object")
			}
			obj, err = repository.ReadObj(&repo, tree[0])
		}

		treeObj, ok := obj.(*repository.Tree)
		if !ok {
			return fmt.Errorf("")
		}

		if pathStat, err := os.Stat(path); err == nil {
			if !pathStat.IsDir() {
				return fmt.Errorf("Not a directory %s", path)
			}
			if pathDir, err := os.ReadDir(path); err == nil || len(pathDir) != 0 {
				return fmt.Errorf("Not empty %s", path)
			}
		} else {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("error creating directories")
			}
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("cannot create absolute path")
		}

		if err := treeCheckout(&repo, treeObj, absPath); err != nil {
			return err
		}

		return nil
	},
}

func treeCheckout(repo *repository.Repository, tree *repository.Tree, path string) error {
	for _, item := range tree.Items {
		obj, err := repository.ReadObj(repo, item.Sha)
		if err != nil {
			return err
		}
		dest := filepath.Join(path, item.Path)

		switch objType := obj.(type) {
		case *repository.Tree:
			os.Mkdir(dest, 0755)
			err := treeCheckout(repo, objType, dest)
			if err != nil {
				return err
			}
		case *repository.Blob:
			err := os.WriteFile(dest, objType.Serialize(), 0644)
			if err != nil {
				return fmt.Errorf("cannot write file %s", dest)
			}
		default:
			return fmt.Errorf("incorrect repository.type")
		}
	}

	return nil
}
