package repository

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

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

func (r *Repository) ReadIndex() (*Index, error) {
	indexFile, err := r.File("index")
	if err != nil {
		return nil, fmt.Errorf("could not find index path")
	}

	if _, err := os.Stat(*indexFile); err != nil {
		return nil, nil
	}

	raw, err := os.ReadFile(*indexFile)
	if err != nil {
		return nil, fmt.Errorf("error reading index file")
	}

	rawHeader := raw[:12]
	var header IndexHeader
	if err := binary.Read(bytes.NewBuffer(rawHeader), binary.BigEndian, &header); err != nil {
		return nil, fmt.Errorf("cannot read index file header")
	}
	if !bytes.Equal(header.Signature[:], []byte("DIRC")) {
		return nil, fmt.Errorf("incorrect header signature")
	}
	if header.Version != 2 {
		return nil, fmt.Errorf("wyog only supports index file version 2")
	}

	content := raw[12:]
	idx := 0
	entries := make([]IndexEntry, 0)
	for range header.Count {
		var entry IndexBinaryEntry
		if err := binary.Read(bytes.NewBuffer(content[idx:idx+62]), binary.BigEndian, &entry); err != nil {
			return nil, fmt.Errorf("cannot read index entry")
		}

		ctime := time.Unix(int64(entry.CtimeSec), int64(entry.CtimeNSec))
		mtime := time.Unix(int64(entry.MtimeSec), int64(entry.MtimeNSec))

		if entry.Unused != 0 {
			return nil, fmt.Errorf("incorrect index entry format")
		}

		modeType := uint16(entry.Mode >> 12)
		if !slices.Contains([]uint16{0b1000, 0b1010, 0b1110}, modeType) {
			return nil, fmt.Errorf("incorrect mode type")
		}
		modePerms := entry.Mode & 0b0000000111111111

		sha := fmt.Sprintf("%020s", hex.EncodeToString(entry.Sha[:]))

		assumeValid := (entry.Flags & 0b1000000000000000) != 0
		extended := (entry.Flags & 0b0100000000000000) != 0
		if extended {
			return nil, fmt.Errorf("version 2 does not support extended")
		}
		stage := entry.Flags & 0b0011000000000000

		nameLength := entry.Flags & 0b0000111111111111

		idx += 62

		var rawName []byte
		if nameLength < 0xFFF {
			if content[idx+int(nameLength)] != 0x00 {
				return nil, fmt.Errorf("index entry name is incorrect format")
			}

			rawName = content[idx : idx+int(nameLength)]
			idx += int(nameLength) + 1
		} else {
			nullIdx := bytes.Index(content[idx+0xFFF:], []byte{'\x00'})
			rawName = content[idx : idx+nullIdx]
			idx += nullIdx + 1
		}

		name := string(rawName)
		idx = int(8 * math.Ceil(float64(idx)/8))

		entries = append(entries, IndexEntry{
			Ctime:       ctime,
			Mtime:       mtime,
			Dev:         int(entry.Dev),
			Ino:         int(entry.Ino),
			ModeType:    int(modeType),
			ModePerms:   int(modePerms),
			Uid:         int(entry.Uid),
			Gid:         int(entry.Gid),
			Fsize:       int(entry.Fsize),
			Sha:         sha,
			AssumeValid: assumeValid,
			Stage:       int(stage),
			Name:        name,
		})
	}

	index := NewIndexV2(entries)
	return &index, nil
}

func (r *Repository) ReadGitignore() (Ignores, error) {
	ret := NewIgnores()

	repoFile := filepath.Join(r.Gitdir, "info", "exclude")
	if _, err := os.Stat(repoFile); err == nil {
		data, err := os.ReadFile(repoFile)
		if err != nil {
			return Ignores{}, fmt.Errorf("error reading .git/info/exclude")
		}
		lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		ignores := gitignoreParse(lines)
		ret.Absolute = append(ret.Absolute, ignores...)
	}

	configHome, err := filepath.Abs("~/.config")
	if err != nil {
		return Ignores{}, fmt.Errorf("cannot create path %s", configHome)
	}
	if xdgHome, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		config, err := filepath.Abs(xdgHome)
		if err != nil {
			return Ignores{}, fmt.Errorf("cannot create path %s", config)
		}
		configHome = config
	}
	globalFile := filepath.Join(configHome, "git", "ignore")

	if _, err := os.Stat(globalFile); err == nil {
		data, err := os.ReadFile(globalFile)
		if err != nil {
			return Ignores{}, fmt.Errorf("error reading global config")
		}
		lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		ignores := gitignoreParse(lines)
		ret.Absolute = append(ret.Absolute, ignores...)
	}

	index, err := r.ReadIndex()
	if err != nil {
		return Ignores{}, err
	}

	for _, entry := range index.Entries {
		if entry.Name == ".gitignore" || strings.HasSuffix(entry.Name, "/.gitignore") {
			dirName := filepath.Dir(entry.Name)
			contents, err := ReadObj(r, entry.Sha)
			if err != nil {
				return Ignores{}, err
			}
			blobContents, ok := contents.(*Blob)
			if !ok {
				return Ignores{}, fmt.Errorf("%s is not a blob type", entry.Sha)
			}
			lines := strings.Split(strings.ReplaceAll(string(blobContents.data), "\r\n", "\n"), "\n")
			ret.Scoped[dirName] = gitignoreParse(lines)
		}
	}

	return ret, nil
}

func (r *Repository) ActiveBranch() (string, error) {
	headPath, err := r.File("HEAD")
	if err != nil {
		return "", err
	}
	head, err := os.ReadFile(*headPath)
	if err != nil {
		return "", fmt.Errorf("cannot open HEAD")
	}

	if bytes.HasPrefix(head, []byte("ref: refs/heads/")) {
		return string(head[16:]), nil
	}

	return "", nil
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
