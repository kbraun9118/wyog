package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"
)

type Repository struct {
	Worktree string
	Gitdir   string
	Conf     *ini.File
}

func New(path string) (Repository, error) {
	return new(path, false)
}

func NewWithForce(path string) (Repository, error) {
	return new(path, true)
}

func new(path string, force bool) (Repository, error) {
	repo := Repository{
		Worktree: path,
		Gitdir:   filepath.Join(path, ".git"),
	}

	if force {
		return repo, nil
	}

	gitDirInfo, err := os.Stat(repo.Gitdir)
	if err != nil {
		return repo, fmt.Errorf("Not a Git repository %s\n", repo.Gitdir)
	}

	if !gitDirInfo.IsDir() {
		return repo, fmt.Errorf("Not a Git repository %s\n", repo.Gitdir)
	}

	cf := repo.Path("config")

	if _, err := os.Stat(cf); err != nil {
		return repo, fmt.Errorf("Configuration file missing\n")
	}

	repo.Conf, err = ini.Load(cf)
	if err != nil {
		return repo, fmt.Errorf("Invalid configuration file\n")
	}

	ver, err := repo.Conf.Section("core").GetKey("repositoryformatversion")
	if err != nil {
		return repo, fmt.Errorf("No core section in config\n")
	}
	if ver.String() != "0" {
		return repo, fmt.Errorf("Unsupported repositoryformatversion: %s\n", ver.String())
	}

	return repo, nil
}

func (r *Repository) Path(paths ...string) string {
	return filepath.Join(r.Gitdir, filepath.Join(paths...))
}

func (r *Repository) dir(mkdir bool, path ...string) (*string, error) {
	repoPath := r.Path(path...)

	if pathStat, err := os.Stat(repoPath); err == nil {
		if pathStat.IsDir() {
			return &repoPath, nil
		}

		return nil, fmt.Errorf("Not a directory %s\n", repoPath)
	}

	if mkdir {
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return nil, fmt.Errorf("Error creating directories\n")
		}

		return &repoPath, nil
	}

	return nil, nil
}

func (r *Repository) Dir(path ...string) (*string, error) {
	return r.dir(false, path...)
}

func (r *Repository) DirMk(path ...string) (*string, error) {
	return r.dir(true, path...)
}

func (r *Repository) file(mkdir bool, path ...string) (*string, error) {
	p, err := r.dir(mkdir, path[:len(path)-1]...)
	if err != nil {
		return nil, err
	}

	if p != nil {
		repoPath := r.Path(path...)
		return &repoPath, nil
	}

	return nil, nil
}

func (r *Repository) File(path ...string) (*string, error) {
	return r.file(false, path...)
}

func (r *Repository) FileMk(path ...string) (*string, error) {
	return r.file(true, path...)
}

var hashRe *regexp.Regexp = regexp.MustCompile("^[0-9A-Fa-f]{4,40}$")

func (r *Repository) Resolve(name string) ([]string, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, fmt.Errorf("name must be supplied")
	}

	if name == "HEAD" {
		headPath, err := r.File("HEAD")
		if err != nil {
			return nil, err
		}
		head, err := RefResolve(r, *headPath)
		if err != nil {
			return nil, err
		}
		return []string{*head}, nil
	}

	candidates := make([]string, 0)

	if hashRe.Match([]byte(name)) {
		name := strings.ToLower(name)
		prefix := name[0:2]
		path, err := r.Dir("objects", prefix)
		if err != nil {
			return nil, err
		}
		if path != nil {
			rem := name[2:]
			dirs, err := os.ReadDir(*path)
			if err != nil {
				return nil, err
			}

			for _, f := range dirs {
				if strings.HasPrefix(f.Name(), rem) {
					candidates = append(candidates, prefix+f.Name())
				}
			}
		}
	}

	tagPath, err := r.File("refs/tags/" + name)
	if err != nil {
		return nil, err
	}
	asTag, err := RefResolve(r, *tagPath)
	if err != nil {
		return nil, err
	}
	if asTag != nil {
		candidates = append(candidates, *asTag)
	}

	branchPath, err := r.File("refs/heads/" + name)
	if err != nil {
		return nil, err
	}
	asBranch, err := RefResolve(r, *branchPath)
	if err != nil {
		return nil, err
	}
	if asBranch != nil {
		candidates = append(candidates, *asBranch)
	}

	remoteBranchPath, err := r.File("refs/remotes/" + name)
	if err != nil {
		return nil, err
	}
	asRemoteBranch, err := RefResolve(r, *remoteBranchPath)
	if err != nil {
		return nil, err
	}
	if asRemoteBranch != nil {
		candidates = append(candidates, *asRemoteBranch)
	}

	return candidates, nil
}

func Find(path string) *string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil
	}

	if absPathGitStat, err := os.Stat(filepath.Join(absPath, ".git")); err == nil && absPathGitStat.IsDir() {
		return &absPath
	}

	parent := filepath.Dir(absPath)

	if parent == absPath {
		return nil
	}

	return Find(parent)
}

func FindRequire(path string) (string, error) {
	p := Find(path)
	if p == nil {
		return "", fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	return *p, nil
}
