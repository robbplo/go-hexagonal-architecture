package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type SlogLogger struct {
	logger *slog.Logger
}

func New(level string) *SlogLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	})
	return &SlogLogger{
		logger: slog.New(handler),
	}
}

func (l *SlogLogger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	l.logger.DebugContext(ctx, msg, keysAndValues...)
}

func (l *SlogLogger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	l.logger.InfoContext(ctx, msg, keysAndValues...)
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	l.logger.WarnContext(ctx, msg, keysAndValues...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	l.logger.ErrorContext(ctx, msg, keysAndValues...)
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
