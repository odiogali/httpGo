package main

import (
	"fmt"
	"net"

	"httpGo/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Connection has been accepted!")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			// WARNING: This is incorrect behaviour
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
	}
}
