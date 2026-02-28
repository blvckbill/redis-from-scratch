package resp

import (
	"bytes"
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
	Str   *string
	Int   int64
	Array []*Resp
}

/*
respToString is a helper function that converts a RESP object to a string if it is of type SimpleString or BulkString.
It returns the string and a boolean indicating whether the conversion was successful.
*/
func Parser(buffer []byte) (*Resp, int, bool) {
	if len(buffer) == 0 {
		return nil, 0, false
	}
	// Determine the type of RESP message based on the first byte
	switch buffer[0] {

	case '+': // Simple String
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, 0, false
		}
		bytesConsumed := idx + 2
		s := string(buffer[1:idx])
		return &Resp{
			Type: SimpleString,
			Str:  &s,
		}, bytesConsumed, true

	case '-': // Error
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, 0, false
		}
		bytesConsumed := idx + 2
		s := string(buffer[1:idx])
		return &Resp{
			Type: Error,
			Str:  &s,
		}, bytesConsumed, true

	case ':': // Integer
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, 0, false
		}
		// Parse the integer value
		n, err := strconv.ParseInt(string(buffer[1:idx]), 10, 64)
		if err != nil {
			return nil, 0, false
		}
		bytesConsumed := idx + 2
		return &Resp{
			Type: Integer,
			Int:  n,
		}, bytesConsumed, true

	case '$': // Bulk String
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, 0, false
		}

		length, err := strconv.Atoi(string(buffer[1:idx]))
		if err != nil {
			return nil, 0, false
		}
		if length == -1 {
			return &Resp{
				Type: BulkString,
				Str:  nil,
			}, idx + 2, true
		}

		start := idx + 2
		end := start + length

		if len(buffer) < end+2 {
			return nil, 0, false
		}

		if !bytes.Equal(buffer[end:end+2], []byte("\r\n")) {
			return nil, 0, false
		}

		bytesConsumed := idx + 2 + length + 2
		s := string(buffer[start:end])
		return &Resp{
			Type: BulkString,
			Str:  &s,
		}, bytesConsumed, true

	case '*': // Array
		idx := bytes.Index(buffer, []byte("\r\n"))
		if idx == -1 {
			return nil, 0, false
		}

		length, err := strconv.Atoi(string(buffer[1:idx]))
		if err != nil {
			return nil, 0, false
		}

		if length == -1 {
			return &Resp{
				Type:  Array,
				Array: nil,
			}, idx + 2, true
		}

		items := make([]*Resp, 0, length)
		bytesConsumed := idx + 2
		for i := 0; i < length; i++ {
			resp, consumed, ok := Parser(buffer[bytesConsumed:])
			if !ok {
				return nil, 0, false
			}
			items = append(items, resp)
			bytesConsumed += consumed
		}
		return &Resp{
			Type:  Array,
			Array: items,
		}, bytesConsumed, true
	}

	return nil, 0, false
}
