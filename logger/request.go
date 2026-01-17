package logger

import (
	"log/slog"
	"time"
)

type RequestLogger struct {
	*BaseLogger
}

func NewRequestLogger() *RequestLogger {
	return &RequestLogger{
		BaseLogger: CreateBaseLogger("request.log"),
	}
}

func (rl *RequestLogger) Log(method, path, remoteAddr string, statusCode int, durationMs float64) {
	var logFunc func(msg string, args ...any)

	switch {
	case statusCode >= 500:
		logFunc = rl.logger.Error
	case statusCode >= 400:
		logFunc = rl.logger.Warn
	case statusCode >= 300:
		logFunc = rl.logger.Info
	case statusCode >= 200:
		logFunc = rl.logger.Info
	default:
		logFunc = rl.logger.Info
	}

	timestamp := time.Now()
	logFunc("http_request",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("remote_addr", remoteAddr),
		slog.Int("status_code", statusCode),
		slog.Float64("duration_ms", durationMs),
		slog.Time("timestamp", timestamp),
	)
}
