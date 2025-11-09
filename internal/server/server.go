package server

import (
	"bytes"
	"fmt"
	"go-http-server/internal/request"
	"go-http-server/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) Write(w io.Writer) error {
	s := fmt.Sprintf("%d %s", h.StatusCode, h.Message)
	_, err := w.Write([]byte(s))
	return err
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
	handler  Handler
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for !s.isClosed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			log.Println("Accept error:", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})

	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	body := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	h := response.GetDefaultHeaders(len(body))
	response.WriteHeaders(conn, h)
	conn.Write(body)
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: ln,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}
