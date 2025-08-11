package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kbraun9118/wyog/repository"
	"github.com/kbraun9118/wyog/util"
	"github.com/spf13/cobra"
)

func init() {
	commitCmd.Flags().StringVarP(&message, "message", "m", "", "Message to associate with this commit")
	commitCmd.MarkFlagRequired("message")
}

var (
	message   string
	commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Record changes to the repository.",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := repository.ReadConfig()
			if err != nil {
				return err
			}

			path, err := repository.FindRequire(".")
			if err != nil {
				return err
			}
			repo, err := repository.New(path)
			if err != nil {
				return err
			}

			user := config.User()

			head, err := repository.ObjectFind(&repo, "HEAD", "")
			if err != nil {
				return err
			}

			index, err := repo.ReadIndex()
			if err != nil {
				return err
			}
			tree, err := repo.TreeFromIndex(index)
			if err != nil {
				return err
			}

			commit, err := CreateCommit(&repo, time.Now(), tree, head, user, message)
			if err != nil {
				return err
			}
			
			activeBranch, err := repo.ActiveBranch()
			if err != nil {
				return err
			}
			var file *string
			if len(activeBranch) != 0 {
			
				file, err = repo.File("refs", "heads", activeBranch)
				if err != nil {
					return err
				}
			} else {
				file, err = repo.File("HEAD")
				if err != nil {
					return err
				}
			}
			os.WriteFile(*file, []byte(commit+"\n"), 0644)

			return nil
		},
	}
)

func CreateCommit(
	repo *repository.Repository,
	timestamp time.Time,
	tree, parent, author, message string,
) (string, error) {
	commit := repository.Commit{
		Kvlm: repository.KvlmData{
			LinkedMap: util.NewLinkedMap[string, []string](),
		},
	}
	commit.Kvlm.Set("tree", []string{tree})

	if len(parent) != 0 {
		commit.Kvlm.Set("parent", []string{parent})
	}

	message = strings.ReplaceAll(message, "\n", "")
	author = fmt.Sprintf("%s %d %s", author, timestamp.Unix(), timestamp.Format("-7000"))

	commit.Kvlm.Set("author", []string{author})
	commit.Kvlm.Set("committer", []string{author})
	commit.Kvlm.Message = []byte(message)

	sha, err := repository.Write(&commit, repo)
	if err != nil {
		return "", err
	}

	return sha, nil
}
