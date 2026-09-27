package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func ParseLevel(lvl string) Level {
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

var (
	bearerRegex = regexp.MustCompile(`(?i)(bearer\s+)([A-Za-z0-9_\-\.]{8,})`)

	authHeaderRegex = regexp.MustCompile(`(?i)(authorization[:=]\s*)([A-Za-z0-9_\-\.]{8,})`)

	tokenKeyRegex = regexp.MustCompile(`(?i)(token["']?\s*[:=]\s*["']?)([A-Za-z0-9_\-\.]{8,})(["']?)`)
)

func Sanitize(input string) string {
	s := bearerRegex.ReplaceAllString(input, "${1}[REDACTED]")
	s = authHeaderRegex.ReplaceAllString(s, "${1}[REDACTED]")
	s = tokenKeyRegex.ReplaceAllString(s, "${1}[REDACTED]${3}")
	return s
}

type Logger struct {
	mu     sync.Mutex
	level  Level
	writer io.Writer
	closer io.Closer
}

func DefaultLogPath() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "kumo", "kumo.log")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "state", "kumo", "kumo.log")
	}
	return "/tmp/kumo.log"
}

func NewLogger(path string, level Level) (*Logger, error) {
	if path == "" {
		path = DefaultLogPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {

		path = "/tmp/kumo.log"
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
	}

	return &Logger{
		level:  level,
		writer: file,
		closer: file,
	}, nil
}

func NewWriterLogger(w io.Writer, level Level) *Logger {
	return &Logger{
		level:  level,
		writer: w,
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

func (l *Logger) log(lvl Level, format string, args ...any) {
	if lvl < l.level {
		return
	}

	rawMsg := fmt.Sprintf(format, args...)
	safeMsg := Sanitize(rawMsg)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	line := fmt.Sprintf("[%s] [%s] %s\n", timestamp, lvl.String(), safeMsg)

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.writer != nil {
		_, _ = l.writer.Write([]byte(line))
	}
}

func (l *Logger) Debug(format string, args ...any) {
	l.log(LevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...any) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...any) {
	l.log(LevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.log(LevelError, format, args...)
}
