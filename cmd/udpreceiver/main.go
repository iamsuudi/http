package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal("error: ", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("error: ", err)
	}
	defer conn.Close()

	lines := getLinesFromReader(conn)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesFromReader(conn *net.UDPConn) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		str := ""
		buffer := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}
			data := buffer[:n]
			for len(data) > 0 {
				if i := bytes.IndexByte(data, '\n'); i != -1 {
					str += string(data[:i])
					data = data[i+1:]
					out <- str
					str = ""
				} else {
					break
				}
			}
			str += string(data)
			// Emit remaining data without newline (UDP has no EOF, so each datagram is complete)
			if len(str) > 0 {
				out <- str
				str = ""
			}
		}
	}()

	return out
}
