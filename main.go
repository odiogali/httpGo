package main

import (
	"fmt"
	"net"

	"github.com/odiogali/httpGo/internal/request"
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

		RequestFromReader()

		fmt.Println("Channel has been closed.")
	}

}
