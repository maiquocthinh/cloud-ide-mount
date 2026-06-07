package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents a log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelLabels = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

// Logger is a simple structured logger that writes to both stdout (user-facing)
// and a log file (structured with timestamps and fields).
type Logger struct {
	mu      sync.Mutex
	level   Level
	logFile *os.File
}

var std *Logger

// Options configures the logger.
type Options struct {
	Level Level
	Path  string // path to log file
}

// Init initializes the default package-level logger.
// It opens the log file at the given path (creating parent directories as needed)
// and sets the minimum log level. Call Close() to clean up.
func Init(opts Options) error {
	dir := filepath.Dir(opts.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating log dir: %w", err)
	}
	f, err := os.OpenFile(opts.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	std = &Logger{
		level:   opts.Level,
		logFile: f,
	}
	return nil
}

// Close closes the log file if open. Safe to call multiple times.
func Close() {
	if std != nil {
		std.mu.Lock()
		defer std.mu.Unlock()
		if std.logFile != nil {
			std.logFile.Close()
			std.logFile = nil
		}
	}
}

// SetLevel changes the minimum log level.
func SetLevel(level Level) {
	if std == nil {
		return
	}
	std.mu.Lock()
	std.level = level
	std.mu.Unlock()
}

// log writes a log entry at the given level.
// msg is the human-readable message (used for both stdout and file output).
// keysAndValues are structured key-value pairs (only written to the log file).
func (l *Logger) log(level Level, msg string, keysAndValues ...interface{}) {
	if l == nil {
		return
	}

	l.mu.Lock()
	if level < l.level {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	label := levelLabels[level]

	// Build structured fields string for file output
	fields := ""
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		fields += fmt.Sprintf(" %s=%v", key, keysAndValues[i+1])
	}

	// Write to file with timestamp and level
	l.mu.Lock()
	if l.logFile != nil {
		fmt.Fprintf(l.logFile, "%s %s %s%s\n", timestamp, label, msg, fields)
	}
	l.mu.Unlock()

	// Write to stdout
	var prefix string
	switch level {
	case LevelWarn:
		prefix = "  Warning: "
	case LevelError:
		prefix = "  Error: "
	default:
		prefix = "  "
	}

	// stdout: no extra lock needed (os.Stdout is goroutine-safe)
	// but we lock to interleave with file writes cleanly
	l.mu.Lock()
	fmt.Fprintf(os.Stdout, "%s%s\n", prefix, msg)
	l.mu.Unlock()
}

// Debug logs at DEBUG level (only written to file, not stdout by default).
func Debug(msg string, keysAndValues ...interface{}) {
	std.log(LevelDebug, msg, keysAndValues...)
}

// Info logs at INFO level (written to both stdout and file).
func Info(msg string, keysAndValues ...interface{}) {
	std.log(LevelInfo, msg, keysAndValues...)
}

// Warn logs at WARN level (written to both stdout and file, prefixed with "Warning:").
func Warn(msg string, keysAndValues ...interface{}) {
	std.log(LevelWarn, msg, keysAndValues...)
}

// Error logs at ERROR level (written to both stdout and file, prefixed with "Error:").
func Error(msg string, keysAndValues ...interface{}) {
	std.log(LevelError, msg, keysAndValues...)
}
