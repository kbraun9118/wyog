package gitobject

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/kbraun9118/wyog/repository"
)

type GitObject interface {
	Serialize() []byte
}

type GitCommit struct {
}

func (gc GitCommit) Serialize() []byte {
	return nil
}

type GitTree struct {
}

func (gc GitTree) Serialize() []byte {
	return nil
}

type GitTag struct {
}

func (gc GitTag) Serialize() []byte {
	return nil
}

type GitBlob struct {
}

func (gc GitBlob) Serialize() []byte {
	return nil
}

func Read(repo repository.Repository, sha string) (GitObject, error) {
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
	size, err := strconv.Atoi(string(raw[x:y]))
	if err != nil || size != len(raw)-y-1 {
		return nil, fmt.Errorf("Malformed object %s: bad length", sha)
	}

	return nil, nil
}
