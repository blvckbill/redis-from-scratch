package server

import (
	"strconv"
	"strings"

	resp "github.com/blvckbill/redis-from-scratch/internal/protocol"
)

/*
handlePing takes the arguments for the PING command and returns a RESP response.
If no arguments are provided, it returns "PONG".
If one argument is provided, it returns that argument as a bulk string.
If more than one argument is provided, it returns an error.
*/
func handlePing(args []string) *resp.Resp {
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

	s := args[0]
	return &resp.Resp{
		Type: resp.BulkString,
		Str:  &s,
	}
}

/*
handleEcho takes the arguments for the ECHO command and returns a RESP response.
If no arguments are provided, it returns an error.
If one or more arguments are provided, it concatenates them with spaces and returns the result as a bulk string.
*/
func handleEcho(args []string) *resp.Resp {
	if len(args) < 1 {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("ERR nothing to ECHO"),
		}
	}

	s := strings.Join(args, " ")
	return &resp.Resp{
		Type: resp.BulkString,
		Str:  &s,
	}
}

func handleSet(args []string) *resp.Resp {
	if len(args) < 2 {
		return &resp.Resp{
			Type: resp.Error,
			Str:  strPtr("Error wrong number of arguments for 'SET'"),
		}
	}

	key := args[0]
	val := args[1]
	var ttl int64
	// SET key value EX seconds
	if len(args) == 4 && strings.ToUpper(args[2]) == "EX" {
		ttl, _ := strconv.ParseInt(args[3], 10, 64)
	}

	db.Set(key, val, ttl)

	return &resp.Resp{
		Type: resp.SimpleString,
		Str:  strPtr("OK"),
	}
}
