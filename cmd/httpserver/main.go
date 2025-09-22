package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpGo/internal/headers"
	"httpGo/internal/request"
	"httpGo/internal/response"
	"httpGo/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {

			newReqLine := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
			h.Delete("Content-Length")
			h.Set("Transfer-Encoding", "chunked")
			h.Set("Trailer", "X-Content-SHA256")
			h.Set("Trailer", "X-Content-Length")

			resp, err := http.Get(fmt.Sprintf("https://httpbin.org%s", newReqLine))
			if err != nil {
				status = 500
				message = defaultResponses[500]

				h.Set("Content-Type", "text/html")
				h.Replace("Content-Length", fmt.Sprintf("%d", len([]byte(message))))

				w.WriteStatusLine(status)
				w.WriteHeaders(h)
				w.WriteBody(message)
				return
			}
			h.Set("Content-Type", resp.Header.Get("Content-Type"))

			w.WriteStatusLine(response.StatusCode(resp.StatusCode))
			w.WriteHeaders(h)

			fullBody := []byte{}
			for {
				b := make([]byte, 1024)
				n, err := resp.Body.Read(b)
				if err != nil {
					break
				}
				fmt.Printf("Number of bytes read by proxy: %d\n", n)
				fullBody = append(fullBody, b[:n]...)

				w.WriteChunkedBody(b)
			}
			w.WriteChunkedBodyDone()

			trailers := headers.NewHeaders()

			sha256 := sha256.Sum256(fullBody)
			trailers.Set("X-Content-SHA256", hex.EncodeToString(sha256[:]))
			trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
			w.WriteHeaders(trailers)
			return
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/video") {
			status = response.OK
			bytes, err := os.ReadFile("./assets/vim.mp4")
			if err != nil {
				status = 500
				message = defaultResponses[500]

				h.Set("Content-Type", "text/html")
				h.Replace("Content-Length", fmt.Sprintf("%d", len([]byte(message))))

				w.WriteStatusLine(status)
				w.WriteHeaders(h)
				w.WriteBody(message)
				return
			}

			h.Replace("Content-Type", "video/mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(bytes)))

			w.WriteStatusLine(status)
			w.WriteHeaders(h)
			w.WriteBody(bytes)
			return
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
