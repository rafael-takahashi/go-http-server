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

func WriteStatusLine(
	w io.Writer,
	statusCode StatusCode,
) error {
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
	_, err := w.Write([]byte(response))
	return err
}

func WriteHeaders(
	w io.Writer,
	headers headers.Headers,
) error {
	for k, v := range headers {
		_, err := w.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = fmt.Sprintf("%d", contentLength)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"

	return h
}
