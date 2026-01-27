package server

import (
	"fmt"
	"net"
)

func Start() {
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		if err net.Error {
			if err.Timeout() {
				
			}
		}

	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			if net.Error.Timeout() {
				fmt.Println("Timeout error")
			}
		}
		go handleConnection(conn)
	}
	
}

func handleConnection(conn net.Conn) {
	fmt.Println("Connection established successfully")
	fmt.Println(conn)
}

