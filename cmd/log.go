package cmd

import (
	"fmt"
	"strings"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log [commit]",
	Short: "Display history of a given commit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := repository.FindRequire(".")
		if err != nil {
			return err
		}

		repo, err := repository.New(r)
		if err != nil {
			return err
		}

		fmt.Printf("digraph wyoglog{\n")
		fmt.Printf("  node[shape=rect]\n")

		sha := "HEAD"
		if len(args) > 0 {
			sha = args[0]
		}

		obj, err := repository.ObjectFind(&repo, sha, "commit")
		if err != nil {
			return err
		}
		err = logGraphviz(&repo, obj, make(map[string]bool))
		if err != nil {
			return err
		}

		fmt.Printf("}\n")

		return nil
	},
}

func logGraphviz(repo *repository.Repository, sha string, seen map[string]bool) error {
	if seen[sha] {
		return nil
	}
	seen[sha] = true

	obj, err := repository.Read(repo, sha)
	if err != nil {
		return err
	}

	commit, ok := obj.(*repository.Commit)
	if !ok {
		return fmt.Errorf("%s is not a commit", sha)
	}

	message := strings.TrimSpace(string(commit.Kvlm.Message))
	message = strings.ReplaceAll(message, "\\", "\\\\")
	message = strings.ReplaceAll(message, "\"", "\\\"")

	splits := strings.SplitN(message, "\n", 1)
	message = splits[0]

	fmt.Printf("  c_%s [label=\"%s: %s\"]\n", sha, sha[:7], message)

	parent, ok := commit.Kvlm.Headers.Get("parent")
	if !ok {
		return nil
	}

	for _, p := range parent {
		fmt.Printf("  c_%s -> c_%s;\n", sha, p)
		logGraphviz(repo, p, seen)
	}

	return nil
}
