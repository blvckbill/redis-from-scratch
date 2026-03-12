package server

import (
	"log"
	"os"
	"sync"
	"time"
)

type AOFLogger struct {
	file *os.File
	mu   sync.RWMutex
}

func NewAOFLogger(path string) (*AOFLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	a := &AOFLogger{
		file: file,
	}
	go a.BackgroundFsync()
	return a, nil
}

func (a *AOFLogger) Append(cmd []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, err := a.file.Write(cmd)

	return err
}

func (a *AOFLogger) BackgroundFsync() error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		err := a.file.Sync()
		a.mu.Unlock()

		if err != nil {
			log.Printf("AOF fsync error: %v", err)
		}
	}

	return nil
}
