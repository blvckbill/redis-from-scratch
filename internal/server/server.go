package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
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
			parsed := commandExecution(parsedResp)
			bytes_parsed := respEncoder(parsed)

			_, err := conn.Write(bytes_parsed)

			if err != nil {
				log.Printf("Error writing to connection: %v", err)
				return
			}
			fmt.Printf("Sent response: %s", string(bytes_parsed))
		}
	}
}

// listen for conn - call accept - read from conn into buffer - call parser on buffer after completion - returns structured resp - if command, handle using command execution - returns resp object in structured format - encode back into bytes - write to conn
func commandExecution(r *resp.Resp) *resp.Resp {
	if len(r.Array) == 0 {
		return nil
	}
	if r.Type != resp.Array {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR unknown command"),
		}
	}

	cmd := strings.ToUpper(*r.Array[0].Str)

	switch cmd {
	case "PING":
		return handlePing(r.Array[1:])
	case "ECHO":
		return handleEcho(r.Array[1:])
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

	if len(args) > 1 {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR wrong number of arguments for 'ping' command"),
		}
	}

	// PING with message
	str, ok := respToString(args[0])
	if !ok {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR invalid argument for 'PING'"),
		}
	}

	return &resp.Resp{
		Type: resp.BulkString,
		Str:  &str,
	}
}

func handleEcho(args []*resp.Resp) *resp.Resp {
	if len(args) < 1 {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR nothing to ECHO"),
		}
	}
	for _, el := args {
		
	}
	str, ok := respToString(echo_args)
	if !ok {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR invalid argument for 'ECHO'"),
		}
	}

	return &resp.Resp{
		Type: resp.BulkString,
		Str:  &str,
	}
}

func strPtr(s string) *string {
	return &s
}

func respToString(r *resp.Resp) (string, bool) {
	switch r.Type {
	case resp.BulkString, resp.SimpleString:
		if r.Str == nil {
			return "", false
		}
		return *r.Str, true

	case resp.Integer:
		return fmt.Sprintf("%d", r.Int), true

	default:
		return "", false
	}
}

func respEncoder(r *resp.Resp) []byte {
	switch r.Type {

	case resp.SimpleString:
		return []byte("+" + *r.Str + "\r\n")

	case resp.Error:
		return []byte("-" + *r.Str + "\r\n")

	case resp.Integer:
		return []byte(":" + strconv.FormatInt(r.Int, 10) + "\r\n")

	case resp.BulkString: //$4/r/nPING/r/n
		if r.Str == nil {
			return []byte("$-1\r\n")
		}
		s := *r.Str
		return []byte("$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n")

	case resp.Array: //*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n
		if r.Array == nil {
			return []byte("*-1\r\n")
		}
		var out []byte

		out = append(out, []byte("*"+strconv.Itoa(len(r.Array))+"\r\n")...)

		for _, el := range r.Array {
			out = append(out, respEncoder(el)...)
		}
		return out
	}
	return []byte("-ERR unknown RESP type\r\n")
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
