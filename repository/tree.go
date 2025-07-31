package repository

import (
	"bytes"
	"cmp"
	"encoding/hex"
	"fmt"
)

type TreeLeaf struct {
	Mode []byte
	Path string
	Sha  string
}

func treeParseOne(raw []byte, start int) (int, TreeLeaf) {
	x := start + bytes.Index(raw[start:], []byte{' '})
	if !(x-start == 5 || x-start == 6) {
		panic("invalid mode definition")
	}

	mode := raw[start:x]
	if len(mode) == 5 {
		mode = append([]byte("0"), mode...)
	}

	y := x + bytes.Index(raw[x:], []byte{'\x00'})
	path := raw[x+1 : y]

	raw_sha := hex.EncodeToString(raw[y+1 : y+21])
	sha := fmt.Sprintf("%020s", raw_sha)
	return y + 21, TreeLeaf{
		Mode: mode,
		Path: string(path),
		Sha:  sha,
	}
}

func TreeParse(raw []byte) []TreeLeaf {
	pos := 0
	max := len(raw)
	ret := make([]TreeLeaf, 0)
	for pos < max {
		var data TreeLeaf
		pos, data = treeParseOne(raw, pos)
		ret = append(ret, data)
	}

	return ret
}

func treeLeafSort(a, b TreeLeaf) int {
	aPath := a.Path
	if bytes.HasPrefix(a.Mode, []byte("10")) {
		aPath += "/"
	}

	bPath := b.Path
	if bytes.HasPrefix(b.Mode, []byte("10")) {
		bPath += "/"
	}

	return cmp.Compare(aPath, bPath)
}
