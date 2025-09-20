package main

import (
	"fmt"
	"httpGo/internal/request"
	"httpGo/internal/response"
	"httpGo/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	port = 42069
)

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		var (
			message []byte
			status  response.StatusCode
		)
		h := response.GetDefaultHeaders(0)

		defaultResponses := make(map[response.StatusCode][]byte, 3)
		defaultResponses[200] = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		defaultResponses[400] = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		defaultResponses[500] = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

		if req.RequestLine.RequestTarget == "/yourproblem" {
			status = 400
			message = defaultResponses[status]

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = 500
			message = defaultResponses[status]
		} else {
			status = 200
			message = defaultResponses[status]
		}
		h.Set("Content-Type", "text/html")
		h.Replace("Content-Length", fmt.Sprintf("%d", len([]byte(message))))

		w.WriteStatusLine(status)
		w.WriteHeaders(h)
		w.WriteBody(message)
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1) // unbuffered channel for signals is created
	// When interrupt and termination signals are sent, send em to the channel
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
