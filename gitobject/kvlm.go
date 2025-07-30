package gitobject

import (
	"bytes"
	"strings"

	"github.com/kbraun9118/wyog/repository/util"
)

type kvlmData struct {
	Headers util.LinkedMap[string, []string]
	Message []byte
}

func kvlmParse(raw []byte, start int, headers *util.LinkedMap[string, []string]) kvlmData {
	if headers == nil {
		headers = util.NewLinkedMap[string, []string]()
	}

	spc := start + bytes.Index(raw[start:], []byte(" "))
	n1 := start + bytes.Index(raw[start:], []byte("\n"))

	if spc < 0 || n1 < spc {
		if n1 != start {
			panic("n1 and start should equal")
		}

		return kvlmData{
			Headers: *headers,
			Message: raw[start+1:],
		}
	}

	key := raw[start:spc]

	end := start

	for {
		end = end + 1 + bytes.Index(raw[end+1:], []byte("\n"))
		if raw[end+1] != ' ' {
			break
		}
	}

	value := bytes.ReplaceAll(raw[spc+1:end], []byte("\n "), []byte("\n"))

	if v, ok := headers.Get(string(key)); ok {
		headers.Set(string(key), append(v, string(value)))
	} else {
		headers.Set(string(key), []string{string(value)})
	}

	return kvlmParse(raw, end+1, headers)
}

func (kvlm kvlmData) Serialize() []byte {
	ret := make([]byte, 0)

	for k, val := range kvlm.Headers.Iterate() {
		for _, v := range val {
			ret = append(ret, []byte(k)...)
			ret = append(ret, ' ')
			ret = append(ret, []byte(strings.ReplaceAll(v, "\n", "\n "))...)
			ret = append(ret, '\n')
		}
	}

	ret = append(ret, '\n')
	ret = append(ret, kvlm.Message...)

	return ret
}
