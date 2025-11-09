package headers

import (
	"bytes"
	"errors"
	"go-http-server/internal/tokens"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if len(data) <= 0 {
		return 0, false, errors.New("headers error: missing end of headers")
	}

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

	if !isValidFieldName(fieldName) {
		return 0, false, errors.New("headers error: invalid character in field name")
	}

	h.Set(string(fieldName), string(fieldValue))

	return crlfIndex + len(tokens.CRLF), false, nil
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key string, val string) {
	lowerKey := strings.ToLower(key)
	if existing, ok := h[lowerKey]; ok && existing != "" {
		h[lowerKey] = existing + ", " + val
	} else {
		h[lowerKey] = val
	}
}

func (h Headers) Delete(key string) {
	delete(h, strings.ToLower(key))
}

func (h Headers) Replace(key string, val string) {
	h.Delete(key)
	h[strings.ToLower(key)] = val
}

func isValidFieldName(fieldName []byte) bool {
	for _, b := range fieldName {
		switch {
		case 'A' <= b && b <= 'Z':
		case 'a' <= b && b <= 'z':
		case '0' <= b && b <= '9':
		case b == '!' || b == '#' || b == '$' || b == '%' || b == '&' ||
			b == '\'' || b == '*' || b == '+' || b == '-' || b == '.' ||
			b == '^' || b == '_' || b == '`' || b == '|' || b == '~':
		default:
			return false
		}
	}
	return true
}
