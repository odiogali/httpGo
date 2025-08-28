package server

import (
	"fmt"
	"httpGo/internal/request"
	"io"
	"net"
	"sync/atomic"
)

type Server struct {
	Addr   string
	Closed atomic.Bool
}

func Serve(port int) (*Server, error) {
	server := &Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	server.Closed.Store(false)

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	return nil
}

func (s *Server) listen() {
	listener, _ := net.Listen("tcp", s.Addr)
	defer listener.Close()

	for { // repeatedly accept connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection: ", err)
			return
		}

		if s.Closed.Load() {
			err := conn.Close()
			if err != nil {
				fmt.Println("error closing connection: ", err)
			}
			return // if server is closed, close connection and exit function
		}

		go func() {
			defer conn.Close()
			for !s.Closed.Load() {
				req, err := request.RequestFromReader(conn)
				if err != nil {
					if err == io.EOF {
						fmt.Println("Client closed connection.")
						return
					}
					panic(err)
				}

				fmt.Println("Request line:")
				fmt.Printf("- Method: %s\n", req.RequestLine.Method)
				fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
				fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
				fmt.Println("Headers:")
				for key, val := range req.Headers {
					fmt.Printf("- %s: %s\n", key, val)
				}
				fmt.Println("Body:")
				fmt.Println(string(req.Body))

				s.handle(conn)
			}
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	initial := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 12\r\n\r\nHello World!"

	n, err := conn.Write([]byte(initial))
	if err != nil || n != len(initial) {
		fmt.Println("error writing to connection: ", err)
		return
	}
}
