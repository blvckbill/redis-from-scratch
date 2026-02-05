package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"github.com/blvckbill/redis-from-scratch/internal/protocol"
)

func Start() {
	addr := ":6369"
	// Create a TCP listening socket bound to the address.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// if there is an error, fail fast like redis does
		log.Fatalf("Fatal: could not start listener: %v", err)
	}

	defer ln.Close()

	log.Printf("GoRedis is listening on %s", addr)

	// create a loop to wait for an Accept on the listener
	for {
		conn, err := ln.Accept()
		if err != nil {
			// Hnadle Accept error
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				fmt.Printf("Timeout occurred, just waiting for the next caller...")
				continue
			}
			log.Fatalf("Accept error: %v", err)
		}
		// once there is a connection, hand it off to another process using go concurrency so bloacking is avoided
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Connection established successfully")
	// create a buffer to read data from the connection and write it back to the client
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("Connection disconnected or error: %v", err)
				break
			}	
		}
		readBuffer := buf[:n]
		parsed_resp, consumed, bool := resp.Parser(buf[readBuffer])
		if parsed_resp != nil && consumed != 0 && bool != false {
				
			}
		}	
	}
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
