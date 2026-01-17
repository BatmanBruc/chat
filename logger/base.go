package logger

import (
	"chat-api/utils"
	"io"
	"log/slog"
	"os"
	"time"
)

type BaseLogger struct {
	logger *slog.Logger
}

func CreateBaseLogger(filename string) *BaseLogger {
	writers := []io.Writer{os.Stdout}
	logToFile := utils.GetEnv("LOG_TO_FILE", "on")

	if filename != "" && logToFile == "on" {
		if err := os.MkdirAll("logs", 0755); err != nil {
			slog.Default().Error("failed to create logs directory", "error", err)
		} else {
			logFile, err := os.OpenFile("logs/"+filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				slog.Default().Error("failed to open log file", "error", err, "filename", filename)
			} else {
				writers = append(writers, logFile)
			}
		}
	}

	multiWriter := io.MultiWriter(writers...)

	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &BaseLogger{
		logger: slog.New(handler),
	}
}

func (dl *BaseLogger) LogError(operation string, err error) {
	timestamp := time.Now()
	dl.logger.Error(operation, err.Error(), slog.Time("timestamp", timestamp))
}

func (dl *BaseLogger) LogWarn(operation string, warn string) {
	timestamp := time.Now()
	dl.logger.Error(operation, warn, slog.Time("timestamp", timestamp))
}

func (dl *BaseLogger) LogInfo(operation string, info string) {
	timestamp := time.Now()
	dl.logger.Error(operation, info, slog.Time("timestamp", timestamp))
}
