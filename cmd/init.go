package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kbraun9118/wyog/repository"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new, empty repository.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%s is not a valid path", path)
		}
		return repoCreate(absPath)
	},
}

func repoCreate(path string) error {
	repo, err := repository.NewWithForce(path)
	if err != nil {
		return err
	}

	if repoStat, err := os.Stat(repo.Worktree); err == nil {
		if !repoStat.IsDir() {
			return fmt.Errorf("%s is not a directory!\n", path)
		}
		if gitDirFiles, err := os.ReadDir(repo.Gitdir); err == nil && len(gitDirFiles) > 0 {
			return fmt.Errorf("%s is not empty!\n", path)
		}
	} else {
		if err := os.MkdirAll(repo.Worktree, 0755); err != nil {
			return fmt.Errorf("Cannot create worktree\n")
		}
	}

	for _, dir := range [][]string{{"branches"}, {"objects"}, {"refs", "tags"}, {"refs", "heads"}} {
		if _, err := repo.DirMk(dir...); err != nil {
			return err
		}
	}

	descFile, err := repo.File("description")
	if err != nil {
		return err
	}
	if err := os.WriteFile(
		*descFile,
		[]byte("Unnamed repository; edit this file 'description' to name the repository.\n"),
		0755,
	); err != nil {
		return fmt.Errorf("Could not create description file")
	}

	headFile, err := repo.File("HEAD")
	if err != nil {
		return err
	}
	if err := os.WriteFile(
		*headFile,
		[]byte("ref: refs/heads/main\n"),
		0755,
	); err != nil {
		return fmt.Errorf("Could not create HEAD file")
	}

	configFile, err := repo.File("config")
	if err != nil {
		return fmt.Errorf("could not create config file")
	}
	configIni, err := defaultConfig()
	if err != nil {
		return fmt.Errorf("could not create config")
	}
	if err := configIni.SaveTo(*configFile); err != nil {
		return fmt.Errorf("Could not write config file")
	}

	return nil
}

func defaultConfig() (*ini.File, error) {

	configIni := ini.Empty()
	coreSection, err := configIni.NewSection("core")
	if err != nil {
		return nil, err
	}

	coreSection.NewKey("repositoryformatversion", "0")
	coreSection.NewKey("filemode", "false")
	coreSection.NewKey("bare", "false")

	return configIni, nil
}
