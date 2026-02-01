package server

import (
	"fmt"
	"log"
	"net"
)

func Start() {
	addr := ":6369"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Fatal: could not start listener: %v", err)
	}

	defer ln.Close()

	log.Printf("GoRedis is listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				fmt.Printf("Timeout occurred, just waiting for the next caller...")
				continue
			}
			log.Fatalf("Accept error: %v", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Connection established successfully")
	fmt.Println(conn)
}

// func RetryWithBackoff(ln net.Listener) (net.Conn, error) {
// 	backoff := 1 * time.Second

// 	for i := 0; i < 3; i++ {
// 		conn, err := ln.Accept()
// 		if err == nil {
// 			return conn, nil
// 		}
// 		fmt.Printf("Attemp %d failed, retrying in %v...", i+1, backoff)

// 		time.Sleep(backoff)
// 		backoff *= 2
// 	}
// 	return nil, fmt.Errorf("Failed after 3 attempts")
// }
