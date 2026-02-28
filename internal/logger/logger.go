package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	eventFile io.WriteCloser
	mutex     sync.Mutex
}

func Start(logPath string) (*Logger, error) {
	// Lumberjack handles rotation automatically
	logFile := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10,   // MB - rotate after 10MB
		MaxBackups: 5,    // Keep 5 old files
		MaxAge:     7,    // Days - delete files older than 7 days
		Compress:   true, // Compress old files with gzip
	}

	return &Logger{
		eventFile: logFile,
	}, nil
}

func (l *Logger) Info(eventType string, data map[string]interface{}) {
	l.log("INFO", eventType, data)
}

func (l *Logger) Error(eventType string, data map[string]interface{}) {
	l.log("ERROR", eventType, data)
}

func (l *Logger) log(level, eventType string, data map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	event := map[string]interface{}{
		"time":  time.Now().Format(time.RFC3339),
		"level": level,
		"event": eventType,
		"data":  data,
	}

	// Write to JSONL file (lumberjack handles rotation)
	json.NewEncoder(l.eventFile).Encode(event)

	// Print to console for development visibility
	fmt.Printf("[%s] %s: %v\n", level, eventType, data)
}

func (l *Logger) Close() error {
	return l.eventFile.Close()
}
