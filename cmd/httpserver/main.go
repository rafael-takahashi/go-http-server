package main

import (
	"go-http-server/internal/request"
	"go-http-server/internal/response"
	"go-http-server/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var HtmlResponses = map[response.StatusCode]string{
	response.StatusBadRequest: `<html>
	<head><title>400 Bad Request</title></head>
	<body>
		<h1>Bad Request</h1>
		<p>Malformed Request.</p>
	</body>
</html>`,

	response.StatusInternalServerError: `<html>
	<head>
		<title>500 Internal Server Error</title>
	</head>
	<body>
		<h1>Internal Server Error</h1>
		<p>Something went wrong on our end.</p>
	</body>
</html>`,

	response.StatusOK: `<html>
	<head>
		<title>200 OK</title>
	</head>
	<body>
		<h1>Success!</h1>
		<p>All good.</p>
	</body>
</html>`,
}

const port = 8080

func writeHTMLResponse(w *response.Writer, status response.StatusCode) {
	body := HtmlResponses[status]
	h := response.GetDefaultHeaders(len(body))
	w.WriteStatusLine(status)
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			writeHTMLResponse(w, response.StatusBadRequest)
		case "/myproblem":
			writeHTMLResponse(w, response.StatusInternalServerError)
		default:
			writeHTMLResponse(w, response.StatusOK)
		}
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
