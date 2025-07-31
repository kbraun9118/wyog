package repository

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func RefResolve(repo *Repository, ref string) (*string, error) {

	if fileStat, err := os.Stat(ref); err != nil || fileStat.IsDir() {
		return nil, nil
	}

	data, err := os.ReadFile(ref)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %s", ref)
	}
	data = data[:len(data)-1]

	if bytes.HasPrefix(data, []byte("ref: ")) {
		path, err := repo.File(string(data[5:]))
		if err != nil {
			return nil, err
		}

		return RefResolve(repo, *path)
	}

	out := string(data)
	return &out, nil
}

func RefList(repo *Repository, path *string) (map[string]any, error) {
	if path == nil {
		var err error
		path, err = repo.Dir("refs")
		if err != nil {
			return nil, err
		}
	}

	ret := make(map[string]any)

	dirs, err := os.ReadDir(*path)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory %s", *path)
	}
	slices.SortFunc(dirs, func(a, b os.DirEntry) int {
		return cmp.Compare(a.Name(), b.Name())
	})

	for _, f := range dirs {
		can := filepath.Join(*path, f.Name())
		canStat, err := os.Stat(can)
		if err != nil {
			return nil, fmt.Errorf("cannot stat file %s", can)
		}
		if canStat.IsDir() {
			out, err := RefList(repo, &can)
			if err != nil {
				return nil, err
			}
			ret[f.Name()] = out
		} else {
			out, err := RefResolve(repo, can)
			if err != nil {
				return nil, err
			}
			ret[f.Name()] = *out
		}
	}

	return ret, nil
}
