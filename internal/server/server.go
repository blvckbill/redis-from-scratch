package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	resp "github.com/blvckbill/redis-from-scratch/internal/protocol"
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
	readBuf := make([]byte, 1024)
	var buffer []byte

	for {
		n, err := conn.Read(readBuf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("Connection disconnected or error: %v", err)
				break
			}
		}

		buffer = append(buffer, readBuf[:n]...)
		for {
			parsedResp, consumed, ok := resp.Parser(buffer)
			if !ok {
				break
			}
			buffer = buffer[consumed:]
			
			fmt.Printf("Parsed RESP: %+v\n", parsedResp)
			commandExecution(parsedResp)
		}
	}
}

func commandExecution(r *resp.Resp) *resp.Resp {
	if r.Type != resp.Array || len(r.Array) == 0 {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR unknown command"),
		}
	}

	cmd := strings.ToUpper(*r.Array[0].Str)

	switch cmd {
	case "PING":
		return handlePing(r.Array[1:])
	default:
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR unknown command"),
		}
	}
}

func handlePing(args []*resp.Resp) *resp.Resp {
	if len(args) == 0 {
		return &resp.Resp{
			Type: resp.SimpleString,
			Str:  strPtr("PONG"),
		}
	}

	// PING with message
	if args[0].Type == resp.BulkString && args[0].Str != nil {
		return &resp.Resp{
			Type: resp.BulkString,
			Str:  args[0].Str,
		}
	}

	return &resp.Resp{
		Type: resp.Error,
		Str:  strPtr("ERR invalid PING"),
	}
}

func strPtr(s string) *string {
	return &s
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
