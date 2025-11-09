package response

import (
	"fmt"
	"go-http-server/internal/headers"
	"io"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var response string
	switch statusCode {
	case StatusOK:
		response = "HTTP/1.1 200 OK\r\n"
	case StatusBadRequest:
		response = "HTTP/1.1 400 Bad Request\r\n"
	case StatusInternalServerError:
		response = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		response = ""
	}

	_, err := w.writer.Write([]byte(response))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	b := []byte{}
	for k, v := range headers {
		b = fmt.Appendf(b, "%s: %s\r\n", k, v)
	}
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.writer.Write(p)
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = fmt.Sprintf("%d", contentLength)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"

	return h
}
