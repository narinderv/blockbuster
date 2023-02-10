package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

// Logging Levels
const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

// Return the string representation of the error level
func (l Level) ErrorLevel() string {
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

// The logger structure
type Logger struct {
	out      io.Writer
	minLevel Level
	mutx     sync.Mutex
}

// Create a new Logger

func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// Helper funtion to write out the actual log message
func (logger *Logger) print(level Level, msg string, properties map[string]string) (int, error) {

	// Wirte the log message only if it is above the minimum level configured for logging
	if level < logger.minLevel {
		return 0, nil
	}

	// Declare an anonymous struct holding the data for the log entry.
	logMsg := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.ErrorLevel(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    msg,
		Properties: properties,
	}

	// Add the stack trace if this is an error
	if level >= LevelError {
		logMsg.Trace = string(debug.Stack())
	}

	// Convert the structure into a byte array for actual logging
	var logEntry []byte

	/// Create the log entry, in case of error, log the error message itself
	logEntry, err := json.Marshal(logMsg)
	if err != nil {
		logEntry = []byte(level.ErrorLevel() + " unableto marshal log entry: " + err.Error())
	}

	// Lock the mutex
	logger.mutx.Lock()
	defer logger.mutx.Unlock()

	// Log the entry
	return logger.out.Write(append(logEntry, '\n'))

}

// Implement the Write() function for satisfying the io.Writer interface
func (logger *Logger) Write(msg []byte) (n int, err error) {
	return logger.print(LevelError, string(msg), nil)
}

// Create helper functions for writing out the log messages
func (logger *Logger) PrintInfo(msg string, properties map[string]string) {
	logger.print(LevelInfo, msg, properties)
}

func (logger *Logger) PrintError(err error, properties map[string]string) {
	logger.print(LevelError, err.Error(), properties)
}

func (logger *Logger) PrintFatal(err error, properties map[string]string) {
	logger.print(LevelFatal, err.Error(), properties)

	// In this case we terminate the application
	os.Exit(1)
}
