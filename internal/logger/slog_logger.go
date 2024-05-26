package logger

import (
	"io"
	"log/slog"
	"sync"
)

type SLogger struct {
	disabled bool
	logger   *slog.Logger

	mu sync.Mutex
}

func (l *SLogger) Debug(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.disabled {
		return
	}
	l.logger.Debug(msg, args...)
}

func (l *SLogger) Info(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.disabled {
		return
	}
	l.logger.Info(msg, args...)
}

func (l *SLogger) Warn(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.disabled {
		return
	}
	l.logger.Warn(msg, args...)
}

func (l *SLogger) Error(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.disabled {
		return
	}
	l.logger.Error(msg, args...)
}

func (l *SLogger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.disabled = true
}

func (l *SLogger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.disabled = false
}

func NewSLogger(w io.Writer, level string) Logger {
	var lvl slog.Level

	switch level {
	case LevelDebug:
		lvl = slog.LevelDebug
	case LevelInfo:
		lvl = slog.LevelInfo
	case LevelWarn:
		lvl = slog.LevelWarn
	case LevelError:
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	return &SLogger{logger: slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl}))}
}
