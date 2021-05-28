package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Define a Level type to represent the severity of a log entry
type Level int8

// Initialize constants which represent a specific severity level.
// iota is used to easily assign successive integer values to a set of integer constants.
// It starts at zero, and increments by 1 for every constant declaration, resetting to zero
// again when the word const appears in the code again.
const (
	LevelInfo Level = iota  // Has a value of 0
	LevelError						 // Has a value of 1
	LevelFatal						 // Has a value of 2
	LevelOff							// Has a value of 3
)

// Return a human-friendly string for the severity level
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Return a new Logger instance which writes log entries at or above a minimum severity 
// level to a specific output destination
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out: out,
		minLevel: minLevel,
	}
}

// Define a custom logger type.
// It holds the output destination that the log entries will be written to, the minimum severity level that the 
// log entries will be written for, plus a mutex for coordinating the writes.
type Logger struct {
	out io.Writer
	minLevel Level
	mu sync.Mutex
}

// print() is an internal method for writing the log entry
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// If the severity level of the log entry is below the minimum severity level for the logger, 
	// return and do nothing
	if level < l.minLevel {
		return 0, nil
	}

	// Declare an annonymous struct for holding the log entry data
	aux := struct {
		Level string `json:"level"`
		Time string `json:"time"`
		Message string `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace string `json:"trace,omitempty"`
	} {
		Level: level.String(),
		Time: time.Now().UTC().Format(time.RFC3339),
		Message: message,
		Properties: properties,
	}

	// Log stack trace only on error level
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// Define a line variable for holding the log entry text
	var line []byte

	// Marshal the annonymous struct into JSON and store it in the line variable.
	// If there is a problem creating the JSON, log a single error message instead.
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log messages:" + err.Error())
	}

	// Locks the mutex so that no two writes to the output destination can happen concurrently.
	// Else, it is possible for the text of two or more log entries to be intermingled in the output.
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append(line, '\n'))
}

// Declare some helper methods for writing log entries at different levels.
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // For entries at the FATAL level, we also terminate the application
}

// Implement a Write() method on the logger such that it satisfies the io.Writer interface.
// This writes a log entry at the ERROR level with no additional properties.
func (l *Logger) Write(message []byte) (int, error) {
	return l.print(LevelError, string(message), nil)
}