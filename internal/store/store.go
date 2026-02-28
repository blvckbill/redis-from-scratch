package store

import (
	"fmt"
	"math/rand"
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
	s := &Store{
		data: make(map[string]Value),
	}

	go s.cleanupExpiredKeys()

	return s
}

func (s *Store) Set(key string, value string, ttlSeconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expires int64
	if ttlSeconds > 0 {
		expires = time.Now().UnixMilli() + (ttlSeconds * 1000)
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
			return 0, fmt.Errorf("value is not an integer")
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

func (s *Store) isExpired(val Value) bool {
	return val.expiresAt > 0 && time.Now().UnixMilli() > val.expiresAt
}

func (s *Store) TTL(key string) int64 {
	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return -2
	}

	if s.isExpired(val) {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return -2
	}

	if val.expiresAt == 0 {
		return -1
	}

	ttl := (val.expiresAt - time.Now().UnixMilli()) / 1000
	if ttl <= 0 {
		ttl = 0
	}
	return ttl
}

func (s *Store) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	for range ticker.C {
		s.mu.Lock()

		// Collect keys with expiration
		expKeys := make([]string, 0, len(s.data))
		for k, v := range s.data {
			if v.expiresAt > 0 {
				expKeys = append(expKeys, k)
			}
		}

		// If no expiring keys, skip this tick
		if len(expKeys) == 0 {
			s.mu.Unlock()
			continue
		}

		// Pick a random subset
		subsetSize := int(0.2 * float64(len(expKeys)))
		if len(expKeys) < subsetSize {
			subsetSize = len(expKeys)
		}

		for i := 0; i < subsetSize; i++ {
			idx := randGen.Intn(len(expKeys))
			key := expKeys[idx]
			val := s.data[key]

			if s.isExpired(val) {
				delete(s.data, key)
			}

			// Remove from expKeys to avoid re-checking
			expKeys = append(expKeys[:idx], expKeys[idx+1:]...)
		}

		s.mu.Unlock()
	}
}
