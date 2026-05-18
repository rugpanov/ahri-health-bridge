package gateways

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Logger struct {
	file *os.File
}

func NewLogger(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{file: f}, nil
}

func (l *Logger) Log(source string, body []byte) {
	entry := map[string]any{
		"source":    source,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"payload":   string(body),
	}
	line, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("[%s] ERROR marshalling log entry: %v\n", source, err)
		return
	}
	fmt.Printf("[%s] %s\n", source, string(line))
	if _, err := l.file.Write(append(line, '\n')); err != nil {
		fmt.Printf("[%s] ERROR writing to log file: %v\n", source, err)
	}
}
