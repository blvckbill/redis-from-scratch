package server

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	resp "github.com/blvckbill/redis-from-scratch/internal/protocol"
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

// Replay reads the AOF file and replays the commands into the store
func (a *AOFLogger) Replay(s *Server) error {
	file, err := os.Open(a.file.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	readbuf := make([]byte, 4096)
	var buffer []byte

	// Read the file in chunks and process the commands
	for {
		bytesRead, err := file.Read(readbuf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalf("Error reading file: %v", err)
			}
		}
		buffer = append(buffer, readbuf[:bytesRead]...)

		// Process the buffer for complete RESP commands
		for {
			// Try to parse the buffer as RESP, if successful, execute the command and write the response back to the client
			parsedResp, consumed, ok := resp.Parser(buffer)
			if !ok {
				break
			}
			// Remove the parsed command from the buffer
			buffer = buffer[consumed:]

			//convert to argv
			parsed, ok := ParsedRespToStrings(parsedResp)
			if !ok {
				log.Printf("Error parsing RESP to strings")
			}
			s.commandExecution(nil, parsed)
		}
	}
	return nil
}
