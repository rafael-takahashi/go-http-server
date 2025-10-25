package request

import (
	"bytes"
	"errors"
	"go-http-server/internal/headers"
	"go-http-server/internal/tokens"
	"io"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type RequestState int

const (
	Initialized RequestState = iota
	ParsingHeaders
	Done
)

type Request struct {
	RequestLine  RequestLine
	RequestState RequestState
	Headers      headers.Headers
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.RequestState {
	case Initialized:
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes <= 0 {
			return 0, nil
		}

		r.RequestLine = requestLine
		r.RequestState = ParsingHeaders

		return numBytes, nil
	case ParsingHeaders:
		numBytes, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.RequestState = Done
		}

		return numBytes, nil
	case Done:
		return 0, errors.New("request error: trying to read data in a done state")
	default:
		return 0, errors.New("request error: unknown request state")
	}
}

const BufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, BufferSize)
	readToIndex := 0
	req := &Request{
		RequestState: Initialized,
		Headers:      headers.NewHeaders(),
	}

	for req.RequestState != Done {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.RequestState = Done
				break
			}
			return nil, err
		}

		readToIndex += n

		numBytes, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytes:readToIndex])
		readToIndex -= numBytes
	}

	return req, nil
}

func parseRequestLine(b []byte) (RequestLine, int, error) {
	// Line parsing
	i := bytes.Index(b, []byte(tokens.CRLF))
	if i == -1 {
		return RequestLine{}, 0, nil
	}
	line := b[:i]
	//rest := b[i+len(CRLF):]
	// --------------------------

	parts := bytes.Split(line, []byte{tokens.SP})
	if len(parts) != 3 {
		return RequestLine{}, 0, errors.New("request error: invalid number of parts in request-line")
	}

	method := parts[0]
	if err := validateMethod(method); err != nil {
		return RequestLine{}, 0, err
	}
	requestTarget := parts[1]
	httpVersion := parts[2]
	if err := validateHttpVersion(httpVersion); err != nil {
		return RequestLine{}, 0, err
	}

	return RequestLine{
		Method:        string(method),
		RequestTarget: string(requestTarget),
		HttpVersion:   string(httpVersion[len(tokens.HTTPVersionPrefix):]),
	}, len(b), nil
}

func validateMethod(method []byte) error {
	if len(method) == 0 {
		return errors.New("request error: request line empty")
	}

	for i := range method {
		if method[i] < 'A' || method[i] > 'Z' {
			return errors.New("request error: method must be uppercase A–Z")
		}
	}

	return nil
}

func validateHttpVersion(httpVersion []byte) error {
	if !bytes.HasPrefix(httpVersion, []byte(tokens.HTTPVersionPrefix)) {
		return errors.New("request error: invalid http name")
	}

	if !bytes.Equal(httpVersion[len(tokens.HTTPVersionPrefix):], []byte("1.1")) {
		return errors.New("request error: invalid http version")
	}

	return nil
}
