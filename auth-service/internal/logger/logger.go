// Package logger настраивает структурированный slog (JSON).
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

// New создаёт JSON-логгер с уровнем debug|info|error.
func New(level string) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	return slog.New(h)
}

// WithRequestID добавляет request_id в контекст для логов.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID из контекста.
func RequestID(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}
