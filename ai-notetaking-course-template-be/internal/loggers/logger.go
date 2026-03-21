package loggers

import (
	"log"
	"time"
)

type Logger struct{}

func NewLooger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(msg string, fields map[string]interface{}) {
	log.Printf("[INFO] %s | %s | %+v\n", time.Now().Format(time.RFC3339), msg, fields)
}

func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	log.Printf("[WARN] %s | %s | %+v\n", time.Now().Format(time.RFC3339), msg, fields)
}
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	log.Printf("[ERROR] %s | %s | %+v\n", time.Now().Format(time.RFC3339), msg, fields)
}

func (l *Logger) TrackTime(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
