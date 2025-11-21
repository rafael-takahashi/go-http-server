# HTTP/1.1 Server in Go

This project implements a server capable of parsing HTTP/1.1 requests according to the relevant RFCs and allows users to define custom handlers for flexible, extensible behavior. This implementation is inspired by the walkthrough provided in the [Learn the HTTP Protocol in Go](https://www.boot.dev/courses/learn-http-protocol-golang) course from Boot.Dev.

---

## Features

- **Full HTTP/1.1 request parsing** based on RFC specifications
- **Concurrent connection handling** using goroutines
- **Minimal dependencies**, built on Go’s standard library

---

## How It Works

### Server Initialization

The server begins by listening on a TCP port using Go’s `net.Listen`:

    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

The listener and handler are stored in a `Server` struct:

    server := &Server{
        listener: ln,
        handler:  handler,
    }

The server starts processing connections asynchronously:

    go server.listen()

Each connection is handled in its own goroutine, enabling high concurrency and responsiveness.

---

## Request Representation

Incoming HTTP/1.1 messages are parsed into the following structure:

    type Request struct {
        RequestLine  RequestLine
        RequestState RequestState
        Headers      headers.Headers
        Body         []byte
    }

This in-memory representation provides access to the request line, headers, and full body content so handlers can process requests however they need.

---

## Handlers

Handlers define server behavior.  
By passing in your own handler when starting the server, you can:

- Create REST-style endpoints
- Implement custom routing
- Serve files or binary formats

---
