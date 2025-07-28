package gitobject

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/kbraun9118/wyog/repository"
)

type GitObject interface {
	Serialize() []byte
	Fmt() []byte
}

type Commit struct {
}

func NewCommit(data []byte) *Commit {
	return &Commit{}
}

func (gc *Commit) Serialize() []byte {
	return nil
}

func (gc *Commit) Fmt() []byte {
	return []byte("commit")
}

type Tree struct {
}

func NewTree(data []byte) *Tree {
	return &Tree{}
}

func (gc *Tree) Serialize() []byte {
	return nil
}

func (gc *Tree) Fmt() []byte {
	return []byte("tree")
}

type Tag struct {
}

func NewTag(data []byte) *Tag {
	return &Tag{}
}

func (gc *Tag) Serialize() []byte {
	return nil
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

func Read(repo *repository.Repository, sha string) (GitObject, error) {
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

func Write(obj GitObject, repo *repository.Repository) (string, error) {
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

func Find(repo *repository.Repository, name string, format string) string {
	return name
}
