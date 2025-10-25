package headers

import (
	"bytes"
	"errors"
	"go-http-server/internal/tokens"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIndex := -1
	colonIndex := -1

	for i := range data {
		if colonIndex == -1 && data[i] == tokens.COLON {
			colonIndex = i
		}
		if i+1 < len(data) && data[i] == tokens.CR && data[i+1] == tokens.LF {
			crlfIndex = i
			break
		}
	}

	if crlfIndex == 0 {
		return len(tokens.CRLF), true, nil
	}

	if crlfIndex == -1 {
		return 0, false, nil
	}

	if colonIndex <= 0 || data[colonIndex-1] == tokens.SP {
		return 0, false, errors.New("headers error: 400 (bad request)")
	}

	fieldName := bytes.TrimSpace(data[:colonIndex])
	fieldValue := bytes.TrimSpace(data[colonIndex+1 : crlfIndex])

	h[string(fieldName)] = string(fieldValue)

	return crlfIndex + len(tokens.CRLF), false, nil
}
