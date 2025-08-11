package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

func init() {
	hashObjectCmd.Flags().StringVarP(&flagType, "type", "t", "blob", "Specify the type (one of [blob, commit, tag, tree])")
	hashObjectCmd.Flags().BoolVarP(&write, "write", "w", false, "Write the object into the database")
}

var (
	flagType      string
	write         bool
	hashObjectCmd = &cobra.Command{
		Use:   "hash-object path",
		Short: "Compute object ID and optionally creates a blob from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var repo *repository.Repository
			if write {
				repoPath := repository.Find(".")
				r, err := repository.New(*repoPath)
				if err != nil {
					return err
				}
				repo = &r
			}

			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("%s not found", args[0])
			}
			defer file.Close()

			sha, err := ObjectHash(file, flagType, repo)
			if err != nil {
				return err
			}

			fmt.Println(sha)
			return nil
		},
	}
)

func ObjectHash(fd *os.File, format string, repo *repository.Repository) (string, error) {
	data, err := io.ReadAll(fd)
	if err != nil {
		return "", fmt.Errorf("cannot read file")
	}

	var obj repository.GitObject
	switch format {
	case "commit":
		obj = repository.NewCommit(data)
	case "tree":
		obj = repository.NewTree(data)
	case "tag":
		obj = repository.NewTag(data)
	case "blob":
		obj = repository.NewBlob(data)
	default:
		return "", fmt.Errorf("invalid type")
	}

	return repository.Write(obj, repo)
}
