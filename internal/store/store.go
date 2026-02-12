package store

import (
	"sync"
	"time"
)

type Value struct {
	data      string
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

func (s *Store) Set(key, value string, ttlSeconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expires int64
	if ttlSeconds > 0 {
		expires = time.Now().Unix() + ttlSeconds
	}

	s.data[key] = Value{
		data:      value,
		expiresAt: expires,
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return "", false
	}

	if val.expiresAt > 0 && time.Now().Unix() > val.expiresAt {
		return "", false
	}

	return val.data, true
}
