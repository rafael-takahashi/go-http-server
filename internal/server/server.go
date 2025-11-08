package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
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

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		// "Content-Length: 13\r\n" +
		"\r\n" +
		"Hello World!\n"

	conn.Write([]byte(response))
}

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: ln,
	}

	go server.listen()

	return server, nil
}
