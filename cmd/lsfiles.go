package cmd

import (
	"fmt"
	"os/user"
	"strconv"
	"time"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
)

func init() {
	lsFilesCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show everyting")
}

var (
	verbose    bool
	lsFilesCmd = &cobra.Command{
		Short: "List all staged files",
		Use:   "ls-files",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := repository.FindRequire(".")
			if err != nil {
				return nil
			}
			repo, err := repository.New(path)
			if err != nil {
				return err
			}
			index, err := repo.ReadIndex()
			if err != nil {
				return err
			}

			fmt.Printf("Index file format v%d, containing %d entries.\n", index.Version, len(index.Entries))

			for _, e := range index.Entries {
				fmt.Println(e.Name)

				if verbose {
					var entryType string
					switch e.ModeType {
					case 0b1000:
						entryType = "regular file"
					case 0b1010:
						entryType = "symlink"
					case 0b1110:
						entryType = "git link"
					default:
						return fmt.Errorf("invalid entry type")
					}

					fmt.Printf("  %s with perms: %04o\n", entryType, e.ModePerms)
					fmt.Printf("  on blob: %s\n", e.Sha)
					fmt.Printf("  created: %s, modified: %s\n", e.Ctime.Format(time.RFC3339Nano), e.Mtime.Format(time.RFC3339Nano))
					fmt.Printf("  device: %d, inode: %d\n", e.Dev, e.Ino)
					u, err := user.LookupId(strconv.Itoa(e.Uid))
					if err != nil {
						return fmt.Errorf("cannot lookup user")
					}
					group, err := user.LookupGroupId(strconv.Itoa(e.Gid))
					if err != nil {
						return fmt.Errorf("cannot lookup group")
					}
					fmt.Printf("  user: %s (%d)  group: %s (%d)\n", u.Username, e.Uid, group.Name, e.Gid)
					fmt.Printf("  flags: stage=%d assume_valid=%t\n", e.Stage, e.AssumeValid)
				}
			}

			return nil
		},
	}
)
