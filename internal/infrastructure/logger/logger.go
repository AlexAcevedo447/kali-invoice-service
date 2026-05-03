package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Level represents log severity level.
type Level string

const (
	InfoLevel  Level = "INFO"
	WarnLevel  Level = "WARN"
	ErrorLevel Level = "ERROR"
	FatalLevel Level = "FATAL"
)

// Fields are arbitrary key-value pairs for structured logging.
type Fields map[string]interface{}

// Entry represents a structured log entry.
type Entry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp string                 `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// log writes a structured JSON log entry to stdout.
func log(level Level, msg string, fields Fields) {
	entry := Entry{
		Level:     string(level),
		Message:   msg,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Fields:    fields,
	}

	data, _ := json.Marshal(entry)
	fmt.Fprintf(os.Stdout, "%s\n", string(data))

	// Exit on Fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Info logs an informational message with optional fields.
func Info(msg string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = fields[0]
	}
	log(InfoLevel, msg, f)
}

// Warn logs a warning message with optional fields.
func Warn(msg string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = fields[0]
	}
	log(WarnLevel, msg, f)
}

// Error logs an error message with optional fields.
func Error(msg string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = fields[0]
	}
	log(ErrorLevel, msg, f)
}

// Fatal logs a fatal message with optional fields and exits.
func Fatal(msg string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = fields[0]
	}
	log(FatalLevel, msg, f)
}
