package request

import (
	"bytes"
	"errors"
	"go-http-server/internal/headers"
	"go-http-server/internal/tokens"
	"io"
	"strconv"
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
	ParsingBody
	Done
)

type Request struct {
	RequestLine  RequestLine
	RequestState RequestState
	Headers      headers.Headers
	Body         []byte
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
			contentLength := r.Headers.Get("content-length")
			if contentLength == "" || contentLength == "0" {
				r.RequestState = Done
			} else {
				r.RequestState = ParsingBody
			}
		}
		return numBytes, nil

	case ParsingBody:
		contentLength := r.Headers.Get("content-length")
		if contentLength == "" || contentLength == "0" {
			r.RequestState = Done
			return len(data), nil
		}
		r.Body = append(r.Body, data...)
		contentLengthVal, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, err
		}
		if len(r.Body) == contentLengthVal {
			r.RequestState = Done
		}
		return len(data), nil

	case Done:
		return 0, nil
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

		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		readToIndex += n

		for {
			numBytes, err := req.parse(buf[:readToIndex])

			if err != nil {
				return nil, err
			}

			if numBytes <= 0 {
				break
			}

			copy(buf, buf[numBytes:readToIndex])
			readToIndex -= numBytes

			if readToIndex <= 0 {
				break
			}
		}
	}

	return req, nil
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	crlfIndex := bytes.Index(data, []byte(tokens.CRLF))
	if crlfIndex == -1 {
		return RequestLine{}, 0, nil
	}
	line := data[:crlfIndex]

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
	}, len(line) + len(tokens.CRLF), nil
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
