package logger

import (
	"log/slog"
	"time"
)

type DatabaseLogger struct {
	*BaseLogger
}

func NewDatabaseLogger() *DatabaseLogger {
	return &DatabaseLogger{
		BaseLogger: CreateBaseLogger("database.log"),
	}
}

func (dl *DatabaseLogger) Log(operation, table string, details string, durationMs float64, err error) {
	timestamp := time.Now()
	if err != nil {
		dl.logger.Error("database_error",
			slog.String("operation", operation),
			slog.String("table", table),
			slog.String("details", details),
			slog.Float64("duration_ms", durationMs),
			slog.Time("timestamp", timestamp),
			slog.String("error", err.Error()),
		)
	} else {
		dl.logger.Info("database_operation",
			slog.String("operation", operation),
			slog.String("table", table),
			slog.String("details", details),
			slog.Float64("duration_ms", durationMs),
			slog.Time("timestamp", timestamp),
		)
	}
}
