package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	out   io.Writer
	level Level
	mu    sync.Mutex
}

type Session struct {
	logger *Logger
	prefix string
	mu     sync.Mutex
}

func New(level Level) *Logger {
	return &Logger{
		out:   os.Stdout,
		level: level,
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

func (l *Logger) NewSession(prefix string) *Session {
	return &Session{
		logger: l,
		prefix: prefix,
	}
}

func (s *Session) log(level Level, format string, args ...interface{}) {
	if level < s.logger.level {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[level]

	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(s.logger.out, "%s [%s] %s%s\n", timestamp, levelStr, s.prefix, msg)
}

func (s *Session) Debug(format string, args ...interface{}) {
	s.log(DEBUG, format, args...)
}

func (s *Session) Info(format string, args ...interface{}) {
	s.log(INFO, format, args...)
}

func (s *Session) Warn(format string, args ...interface{}) {
	s.log(WARN, format, args...)
}

func (s *Session) Error(format string, args ...interface{}) {
	s.log(ERROR, format, args...)
}

func (s *Session) SetPrefix(prefix string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prefix = prefix
}
