package cmd

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/kbraun9118/wyog/repository"
	"github.com/kbraun9118/wyog/util"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add paths...",
	Short: "Add files contents to index.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, paths []string) error {
		p, err := repository.FindRequire(".")
		if err != nil {
			return err
		}
		repo, err := repository.New(p)
		if err != nil {
			return err
		}

		if err := repo.Rm(false, true, paths...); err != nil {
			return err
		}

		cleanPaths := util.NewLinkedMap[string, string]()
		for _, path := range paths {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("cannot convert %s to absolute path", path)
			}

			relPath, err := filepath.Rel(repo.Worktree, absPath)
			if err != nil {
				return fmt.Errorf("outside the worktree: %s", path)
			}

			cleanPaths.Set(absPath, relPath)
		}

		index, err := repo.ReadIndex()
		if err != nil {
			return err
		}

		for absPath, relPath := range cleanPaths.Iterate() {
			file, err := os.Open(absPath)
			if err != nil {
				return fmt.Errorf("cannot open file: %s", absPath)
			}
			sha, err := ObjectHash(file, "blob", &repo)
			if err != nil {
				return err
			}

			stat, err := os.Stat(absPath)
			if err != nil {
				return fmt.Errorf("cannot stat file: %s", absPath)
			}
			if stat.IsDir() {
				return fmt.Errorf("not a file: %s", relPath)
			}
			sysStat, ok := stat.Sys().(*syscall.Stat_t)
			if !ok {
				return fmt.Errorf("not syscall stat")
			}

			entry := repository.IndexEntry{
				Ctime:       time.Unix(sysStat.Ctim.Sec, sysStat.Ctim.Nsec),
				Mtime:       stat.ModTime(),
				Dev:         int(sysStat.Dev),
				Ino:         int(sysStat.Ino),
				ModeType:    0b1000,
				ModePerms:   0o644,
				Uid:         int(sysStat.Uid),
				Gid:         int(sysStat.Gid),
				Fsize:       int(stat.Size()),
				Sha:         sha,
				AssumeValid: false,
				Stage:       0,
				Name:        relPath,
			}

			index.Entries = append(index.Entries, entry)
		}

		slices.SortFunc(index.Entries, func(a, b repository.IndexEntry) int {
			return cmp.Compare(a.Name, b.Name)
		})

		if err := repo.WriteIndex(index); err != nil {
			return err
		}

		return nil
	},
}
