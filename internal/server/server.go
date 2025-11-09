package server

import (
	"fmt"
	"go-http-server/internal/request"
	"go-http-server/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Handler func(w *response.Writer, req *request.Request)

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

	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	s.handler(w, req)
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
