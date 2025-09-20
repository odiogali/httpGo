package server

import (
	"fmt"
	"httpGo/internal/request"
	"httpGo/internal/response"
	"io"
	"net"
	"sync/atomic"
)

type (
	Server struct {
		Addr    string
		Closed  atomic.Bool
		Handler Handler
	}
	Handler      func(w *response.Writer, req *request.Request)
	HandlerError struct {
		StatusCode response.StatusCode
		Msg        string
	}
)

func Serve(port uint16, handlerFunc Handler) (*Server, error) {
	server := &Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handlerFunc,
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

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	responseWriter := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Client closed connection.")
		} else {
			responseWriter.WriteStatusLine(response.BadRequest)
			responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		}
		conn.Close()
		return
	}

	s.Handler(responseWriter, req)
	conn.Close()
}
