package main

import (
	"fmt"
	"log"
	"net"

	"github.com/iamsuudi/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "error", err)
		} else if req != nil {
			fmt.Println("Requesst line:")
			fmt.Printf("- Method: %s\n", req.RequestLine.Method)
			fmt.Printf("- Request Target: %s\n", req.RequestLine.RequestTarget)
			fmt.Printf("- Http Version: %s\n", req.RequestLine.HttpVersion)

			fmt.Println("Headers:")
			for key, values := range req.Headers {
				fmt.Printf("- %s: %s\n", key, values)
			}
		}
	}
}
