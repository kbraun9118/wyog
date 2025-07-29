package gitobject

import "github.com/kbraun9118/wyog/repository/util"

func kflmParse(raw []byte, start int, dct *util.LinkedMap[string, string]) {
	if dct == nil {
		dct = util.NewLinkedMap[string, string]()
	}
}
