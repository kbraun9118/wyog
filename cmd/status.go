package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show the working tree status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := repository.FindRequire(".")
		if err != nil {
			return err
		}
		repo, err := repository.New(path)
		if err != nil {
			return err
		}

		index, err := repo.ReadIndex()
		if err != nil {
			return err
		}

		if err := StatusBranch(&repo); err != nil {
			return err
		}

		if err := StatusHeadIndex(&repo, index); err != nil {
			return err
		}

		fmt.Println()

		if err := StatusIndexWorktree(&repo, index); err != nil {
			return err
		}

		return nil
	},
}

func StatusBranch(repo *repository.Repository) error {
	branch, err := repo.ActiveBranch()
	if err != nil {
		return nil
	}
	if branch != "" {
		fmt.Printf("On branch %s\n", branch)
	} else {
		headObj, err := repository.ObjectFind(repo, "HEAD", "")
		if err != nil {
			return err
		}
		fmt.Printf("HEAD detatched at %s\n", headObj)
	}

	return nil
}

func StatusHeadIndex(repo *repository.Repository, index *repository.Index) error {
	fmt.Printf("Changes to be committed:\n")

	head, err := repository.TreeToDict(repo, "HEAD", "")
	if err != nil {
		return err
	}
	for _, entry := range index.Entries {
		if sha, ok := head[entry.Name]; ok {
			if sha != entry.Sha {
				fmt.Printf("  modified:  %s\n", entry.Name)
			}
			delete(head, entry.Name)
		} else {
			fmt.Printf("  new file:  %s\n", entry.Name)
		}
	}

	for entry := range head {
		fmt.Printf("  deleted:   %s\n", entry)
	}
	return nil
}

func StatusIndexWorktree(repo *repository.Repository, index *repository.Index) error {
	fmt.Printf("Changes not staged for commit:\n")

	ignore, err := repo.ReadGitignore()
	if err != nil {
		return err
	}
	gitdirPrefix := fmt.Sprintf("%s%c", repo.Gitdir, filepath.Separator)

	allFiles := make([]string, 0)

	filepath.WalkDir(repo.Worktree, func(path string, d fs.DirEntry, err error) error {
		if path == repo.Gitdir || strings.HasPrefix(path, gitdirPrefix) {
			return nil
		}
		if !d.IsDir() {
			relPath, err := filepath.Rel(repo.Worktree, path)
			if err != nil {
				return err
			}
			allFiles = append(allFiles, relPath)
		}
		return nil
	})

	for _, entry := range index.Entries {
		fullPath := filepath.Join(repo.Worktree, entry.Name)

		if pathStat, err := os.Stat(fullPath); err == nil {
			if !pathStat.ModTime().Equal(entry.Mtime) {
				fd, err := os.Open(fullPath)
				if err != nil {
					return fmt.Errorf("cannot open file")
				}
				newSha, err := objectHash(fd, "blob", nil)
				if err != nil {
					return err
				}
				if entry.Sha != newSha {
					fmt.Printf("  modified:  %s\n", entry.Name)
				}
			}
		} else {
			fmt.Printf("  deleted:   %s\n", entry.Name)
		}

		allFiles = slices.DeleteFunc(allFiles, func(name string) bool {
			return name == entry.Name
		})
	}

	fmt.Println()
	fmt.Println("Untracked files:")

	for _, f := range allFiles {
		ignored, err := ignore.CheckIgnore(f)
		if err != nil {
			return err
		}
		if !ignored {
			fmt.Printf("  %s\n", f)
		}
	}

	return nil
}
