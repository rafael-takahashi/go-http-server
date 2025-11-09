package main

import (
	"go-http-server/internal/request"
	"go-http-server/internal/response"
	"go-http-server/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 8080

func main() {
	handler := func(w io.Writer, req *request.Request) *server.HandlerError {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your problem is not my problem\n",
			}
		}

		if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "Woopsie, my bad\n",
			}
		}

		w.Write([]byte("All good\n"))
		return nil
	}

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
