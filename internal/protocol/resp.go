package resp

import (
	"bytes"
	"go/types"
	"strconv"
)

type Resptype int

const (
	SimpleString Resptype = iota
	Error
	Integer
	BulkString
	Array
)

type Resp struct {
	Type  Resptype
	Str   string
	Int   int64
	Array []*Resp
}

func Parser(buffer []byte) (*Resp, int, bool) {
	if len(buffer) == 0 {
		return nil, false
	}

	switch buffer[0] {

	case '+': // Simple String
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, false
		}
		n, err := strconv.ParseInt(string(buffer[1:idx]), 10, 64)
		return &Resp{
			Type: SimpleString,
			Str:  string(buffer[1:idx]),
		}, n, true

	case '-': // Error
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, false
		}
		n, err := strconv.ParseInt(string(buffer[1:idx]), 10, 64)
		return &Resp{
			Type: Error,
			Str:  string(buffer[1:idx]),
		}, n, true

	case ':': // Integer
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, false
		}
		n, err := strconv.ParseInt(string(buffer[1:idx]), 10, 64)
		if err != nil {
			return nil, false
		}
		return &Resp{
			Type: Integer,
			Int:  n,
		}, n, true

	case '$': // Bulk String
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, false
		}

		length, err := strconv.Atoi(string(buffer[1:idx]))
		if err != nil {
			return nil, false
		}

		start := idx + 2
		end := start + length

		if len(buffer) < end+2 {
			return nil, false
		}

		if !bytes.Equal(buffer[end:end+2], []byte("\r\n")) {
			return nil, false
		}
		n, err := strconv.ParseInt(string(buffer[start:end]), 10, 64)
		return &Resp{
			Type: BulkString,
			Str:  string(buffer[start:end]),
		}, n, true

	case '*': // Array
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, false
		}

		length, err := strconv.Atoi(string(buffer[1:idx]))
		if err != nil {
			return nil, false
		}
		buffer := buffer[idx+2:]
		items = []*Resp{}
		for i = 0; i < length; i++ {
			resp, consumed := Parser(buffer)
			items = append(items, resp)
			
			buffer := buffer[consumed:]
		}
		return &Resp{
			Type: Array,
			Array: items
		}
	}

	return nil, false
}
