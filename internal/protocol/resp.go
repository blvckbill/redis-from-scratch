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
	Str   string
	Int   int64
	Array []*Resp
}

func Parser(buffer []byte) (*Resp, bool) {
	if len(buffer) == 0 {
		return nil, false
	}
	if buffer[0] == '+' {
		idx := bytes.Index(buffer, []byte{'\r', '\n'})
		if idx == -1 {
			return nil, false
		} else {
			message := buffer[1:idx]
			resp := &Resp{
				Type: SimpleString,
				Str:  string(message),
			}
			return resp, true
		}
	} else if buffer[0] == '-' {
		idx := bytes.Index(buffer, []byte{'\r', '\n'})
		if idx == -1 {
			return nil, false
		} else {
			resp := &Resp{
				Type: Error,
				Str:  string(buffer[1:idx]),
			}
			return resp, true
		}
	} else if buffer[0] == ':' {
		idx := bytes.Index(buffer, []byte{'\r', '\n'})
		if idx == -1 {
			return nil, false
		} else {
			number := buffer[1:idx]
			n, err := strconv.ParseInt(string(number), 10, 64)
			if err != nil {
				return nil, false
			}
			resp := &Resp{
				Type: Integer,
				Int:  n,
			}
			return resp, true
		}
	} else if buffer[0] == '$' {
		idx := bytes.Index(buffer, []byte{'\r', '\n'})
		if idx == -1 {
			return nil, false
		} else {
			idxx := bytes.Index(buffer[idx+2:], []byte{'\r', '\n'})
			message := buffer[idx+2 : idxx]
			if len(message) == int(buffer[1]) {
				resp := &Resp{
					Type: Integer,
					Str:  string(message),
				}
				return resp, true
			} else {
				return nil, false
			}
		}
	}
}
