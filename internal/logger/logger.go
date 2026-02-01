package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger is a wrapper around slog.Logger
type Logger struct {
	*slog.Logger
}

// New creates a new Logger that writes to stdout and an optional file
func New(filePath string, debug bool) *Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// AddSource: true, // Optional: adds source file/line to logs
	}

	var w io.Writer = os.Stdout

	if filePath != "" {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err == nil {
			f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				w = io.MultiWriter(os.Stdout, f)
			}
		}
	}

	handler := slog.NewJSONHandler(w, opts)
	return &Logger{
		Logger: slog.New(handler),
	}
}
