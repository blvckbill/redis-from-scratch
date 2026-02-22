package store

import (
	"strconv"
	"sync"
	"time"
)

type Encoding int

const (
	StringEncoding Encoding = iota
	IntEncoding
)

type Value struct {
	encoding  Encoding
	strVal    string
	intVal    int64
	expiresAt int64
}

type Store struct {
	mu   sync.RWMutex
	data map[string]Value
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]Value),
	}
}

func (s *Store) Set(key string, value string, ttlSeconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expires int64
	if ttlSeconds > 0 {
		expires = time.Now().Unix() + ttlSeconds
	}

	s.data[key] = Value{
		encoding:  StringEncoding,
		strVal:    value,
		expiresAt: expires,
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return "", false
	}

	if val.expiresAt > 0 && time.Now().Unix() > val.expiresAt {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()

		return "", false
	}

	if val.encoding == IntEncoding {
		return strconv.FormatInt(val.intVal, 10), true
	}

	return val.strVal, true
}

func (s *Store) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.data[key]
	if !ok {
		s.data[key] = Value{
			encoding: IntEncoding,
			intVal:   1,
		}
		return 1, nil
	}

	if val.expiresAt > 0 && time.Now().Unix() > val.expiresAt {
		delete(s.data, key)
		s.data[key] = Value{
			encoding: IntEncoding,
			intVal:   1,
		}
		return 1, nil
	}

	switch val.encoding {
	case IntEncoding:
		val.intVal++
		s.data[key] = val
		return val.intVal, nil

	case StringEncoding:
		parsed, err := strconv.ParseInt(val.strVal, 10, 64)
		if err != nil {
			return 0, nil
		}

		parsed++
		val.encoding = IntEncoding
		val.intVal = parsed
		val.strVal = ""
		s.data[key] = val

		return parsed, nil
	}
	return 0, nil
}

func (s *Store) Del(keys []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0

	for _, key := range keys {
		val, ok := s.data[key]
		if !ok {
			continue
		}

		if val.expiresAt > 0 && time.Now().Unix() > val.expiresAt {
			delete(s.data, key)
			continue
		}

		delete(s.data, key)
		count++
	}

	return count
}
