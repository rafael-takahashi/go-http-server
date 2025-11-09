package request

import (
	"bytes"
	"errors"
	"fmt"
	"go-http-server/internal/headers"
	"go-http-server/internal/tokens"
	"io"
	"strconv"
	"strings"
	"unicode"
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

func sanitize(b []byte) string {
	// Replace CR/LF with visible markers to make debugging readable
	s := string(b)
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' {
			return '·'
		}
		return r
	}, s)
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func (r *Request) parse(data []byte) (int, error) {
	fmt.Printf("[DEBUG] parse(): state=%v, data='%s'\n", r.RequestState, sanitize(data))

	switch r.RequestState {
	case Initialized:
		requestLine, numBytes, err := parseRequestLine(data)
		fmt.Printf("[DEBUG] Initialized → parsed request line, numBytes=%d, err=%v\n", numBytes, err)
		if err != nil {
			return 0, err
		}
		if numBytes <= 0 {
			return 0, nil
		}
		r.RequestLine = requestLine
		fmt.Printf("[DEBUG] Request line parsed: %+v\n", requestLine)
		r.RequestState = ParsingHeaders
		return numBytes, nil

	case ParsingHeaders:
		numBytes, done, err := r.Headers.Parse(data)
		fmt.Printf("[DEBUG] ParsingHeaders → numBytes=%d done=%v err=%v\n", numBytes, done, err)
		if err != nil {
			return 0, err
		}
		if done {
			contentLength := r.Headers.Get("content-length")
			fmt.Printf("[DEBUG] Headers done, Content-Length='%s'\n", contentLength)
			r.RequestState = ParsingBody
		}
		return numBytes, nil

	case ParsingBody:
		fmt.Printf("[DEBUG] ParsingBody with %d bytes of data\n", len(data))
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
		fmt.Printf("[DEBUG] Body len=%d / %d\n", len(r.Body), contentLengthVal)
		if len(r.Body) == contentLengthVal {
			r.RequestState = Done
			fmt.Println("[DEBUG] Body complete — transitioning to Done")
		}
		return len(data), nil

	case Done:
		fmt.Println("[DEBUG] Unexpected parse() call after Done")
		return 0, nil
	default:
		fmt.Println("[DEBUG] Unknown request state!")
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

	// TODO: change this to parse until no progress can be made. only them read.
	for req.RequestState != Done {
		fmt.Printf("\n[DEBUG] Loop start — state=%v readToIndex=%d\n", req.RequestState, readToIndex)

		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
			fmt.Printf("[DEBUG] Buffer expanded to %d bytes\n", len(buf))
		}

		n, err := reader.Read(buf[readToIndex:])
		fmt.Printf("[DEBUG] Read %d bytes, err=%v, data='%s'\n", n, err, sanitize(buf[readToIndex:readToIndex+n]))

		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("[DEBUG] Fatal read error: %v\n", err)
			return nil, err
		}

		readToIndex += n
		fmt.Printf("[DEBUG] Total buffered: %d bytes\n", readToIndex)

		numBytes, err := req.parse(buf[:readToIndex])
		fmt.Printf("[DEBUG] parse() consumed %d bytes, state=%v, err=%v\n", numBytes, req.RequestState, err)

		if err != nil {
			fmt.Printf("[DEBUG] parse() returned error: %v\n", err)
			return nil, err
		}

		copy(buf, buf[numBytes:readToIndex])
		readToIndex -= numBytes
		fmt.Printf("[DEBUG] After shifting buffer: readToIndex=%d\n", readToIndex)
	}

	fmt.Println("[DEBUG] Request parsing complete!")

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
