package repository

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type GitObject interface {
	Serialize() []byte
	Fmt() []byte
}

type Commit struct {
	Kvlm KvlmData
}

func NewCommit(data []byte) *Commit {
	return &Commit{
		Kvlm: kvlmParse(data, 0, nil),
	}
}

func (gc *Commit) Serialize() []byte {
	return gc.Kvlm.Serialize()
}

func (gc *Commit) Fmt() []byte {
	return []byte("commit")
}

type Tree struct {
	Items []TreeLeaf
}

func NewTree(data []byte) *Tree {
	return &Tree{
		Items: TreeParse(data),
	}
}

func (gc *Tree) Serialize() []byte {
	items := slices.Clone(gc.Items)
	slices.SortFunc(items, treeLeafSort)

	ret := make([]byte, 0)
	for _, i := range items {
		ret = append(ret, i.Mode...)
		ret = append(ret, ' ')
		ret = append(ret, []byte(i.Path)...)
		ret = append(ret, '\x00')
		sha, err := hex.DecodeString(i.Sha)
		if err != nil {
			panic(fmt.Errorf("error decoding to hex"))
		}
		ret = append(ret, sha...)
	}

	return ret
}

func (gc *Tree) Fmt() []byte {
	return []byte("tree")
}

type Tag struct {
	*Commit
}

func NewTag(data []byte) *Tag {
	return &Tag{
		Commit: NewCommit(data),
	}
}

func (gc *Tag) Fmt() []byte {
	return []byte("tag")
}

type Blob struct {
	data []byte
}

func NewBlob(data []byte) *Blob {
	return &Blob{
		data: data,
	}
}

func (gc *Blob) Deserialize(data []byte) {
	gc.data = data
}

func (gc *Blob) Serialize() []byte {
	return gc.data
}

func (gc *Blob) Fmt() []byte {
	return []byte("blob")
}

func Read(repo *Repository, sha string) (GitObject, error) {
	path, err := repo.File("objects", sha[0:2], sha[2:])
	if err != nil {
		return nil, fmt.Errorf("cannot open object: %s\n", sha)
	}

	file, err := os.ReadFile(*path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read object %s\n", sha)
	}

	zReader, err := zlib.NewReader(bytes.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("Cannot read object: %s\n", sha)
	}
	defer zReader.Close()

	raw, err := io.ReadAll(zReader)
	if err != nil {
		return nil, fmt.Errorf("Cannot read object: %s\n", sha)
	}

	x := bytes.Index(raw, []byte{' '})
	format := raw[:x]

	y := bytes.Index(raw, []byte{'\x00'})
	size, err := strconv.Atoi(string(raw[x+1 : y]))
	if err != nil || size != len(raw)-y-1 {
		return nil, fmt.Errorf("malformed object %s: bad length", sha)
	}

	data := raw[y+1:]
	switch string(format) {
	case "commit":
		return NewCommit(data), nil
	case "tree":
		return NewTree(data), nil
	case "tag":
		return NewTag(data), nil
	case "blob":
		return NewBlob(data), nil
	default:
		return nil, fmt.Errorf("unknown type %s for object %s", string(format), sha)
	}
}

func Write(obj GitObject, repo *Repository) (string, error) {
	data := obj.Serialize()

	result := append(obj.Fmt(), byte(' '))
	result = append(result, []byte(strconv.Itoa(len(data)))...)
	result = append(result, byte('\x00'))
	result = append(result, data...)

	h := sha1.New()
	_, err := h.Write(result)
	if err != nil {
		return "", fmt.Errorf("cannot write sha")
	}
	sha := hex.EncodeToString(h.Sum(nil))

	if repo != nil {
		path, err := repo.FileMk("objects", sha[0:2], sha[2:])
		if err != nil {
			return "", fmt.Errorf("cannot create sha file")
		}

		if _, err = os.Stat(*path); err != nil {
			var compressed bytes.Buffer
			writer := zlib.NewWriter(&compressed)
			if _, err := writer.Write(result); err != nil {
				return "", fmt.Errorf("cannot compress data")
			}

			if err := writer.Close(); err != nil {
				return "", fmt.Errorf("cannot close writer")
			}

			if err := os.WriteFile(*path, compressed.Bytes(), 0755); err != nil {
				return "", fmt.Errorf("cannot write compressed file")
			}
		}
	}

	return sha, nil
}

func find(repo *Repository, name string, format string, follow bool) (string, error) {
	shas, err := repo.Resolve(name)
	if err != nil {
		return "", err
	}

	if len(shas) == 0 {
		return "", fmt.Errorf("no such reference %s", name)
	}

	if len(shas) > 1 {
		return "", fmt.Errorf("ambiguous refernce %s: candidates are:\n - %s", name, strings.Join(shas, "\n - "))
	}

	sha := shas[0]

	if len(format) == 0 {
		return sha, nil
	}

	for {
		obj, err := Read(repo, sha)
		if err != nil {
			return "", err
		}

		if string(obj.Fmt()) == format {
			return sha, nil
		}

		if !follow {
			return "", nil
		}

		switch obj := obj.(type) {
		case *Tag:
			tagObj, ok := obj.Kvlm.Headers.Get("object")
			if !ok || len(tagObj) == 0 {
				return "", fmt.Errorf("no object found")
			}
			sha = tagObj[0]
			continue
		case *Commit:
			if format == "tree" {
				treeObj, ok := obj.Kvlm.Headers.Get("tree")
				if !ok || len(treeObj) == 0 {
					return "", fmt.Errorf("no object found")
				}
				sha = treeObj[0]
				continue
			}
		}

		return "", nil
	}
}

func ObjectFind(repo *Repository, name string, format string) (string, error) {
	return find(repo, name, format, true)
}

func FindNoFollow(repo *Repository, name string, format string) (string, error) {
	return find(repo, name, format, false)
}
