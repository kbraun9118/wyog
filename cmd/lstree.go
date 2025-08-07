package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

func init() {
	lsTreeCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recurse into sub-trees")
}

var (
	recursive bool
	lsTreeCmd = &cobra.Command{
		Use:   "ls-tree tree-ish",
		Short: "Pretty-print a tree object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := repository.FindRequire(".")
			if err != nil {
				return err
			}
			repo, err := repository.New(path)
			if err != nil {
				return err
			}

			if err := ls_tree(repo, args[0], ""); err != nil {
				return err
			}

			return nil
		},
	}
)

func ls_tree(repo repository.Repository, ref string, prefix string) error {
	sha, err := repository.ObjectFind(&repo, ref, "tree")
	if err != nil {
		return err
	}
	obj, err := repository.ReadObj(&repo, sha)
	if err != nil {
		return err
	}

	treeObj, ok := obj.(*repository.Tree)
	if !ok {
		return fmt.Errorf("incorrect object type")
	}

	for _, item := range treeObj.Items {
		var itemType []byte
		if len(item.Mode) == 5 {
			itemType = item.Mode[0:1]
		} else {
			itemType = item.Mode[0:2]
		}

		var typeName string
		switch string(itemType) {
		case "4", "04":
			typeName = "tree"
		case "10":
			typeName = "blob"
		case "12":
			typeName = "blob"
		case "16":
			typeName = "commit"
		default:
			return fmt.Errorf("invalid tree leaf mode %s", string(item.Mode))
		}

		if !(recursive && typeName == "tree") {
			fmt.Printf("%06s %s %s\t%s\n", item.Mode, typeName, item.Sha, filepath.Join(prefix, item.Path))
		} else {
			ls_tree(repo, item.Sha, filepath.Join(prefix, item.Path))
		}
	}

	return nil
}
