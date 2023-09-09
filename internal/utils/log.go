package utils

import (
	"log/slog"
	"os"
)

type LogSeverity uint8

const (
	LogDebug LogSeverity = iota + 1
	LogInfo
	LogWarning
	LogError
)

var logMap = map[LogSeverity]slog.Level{
	LogDebug:   slog.LevelDebug,
	LogInfo:    slog.LevelInfo,
	LogWarning: slog.LevelWarn,
	LogError:   slog.LevelError,
}

// NewLogger returns new slog ref
func NewLogger(severity LogSeverity) *slog.Logger {

	logger := slog.New(slog.NewTextHandler(os.Stdin, &slog.HandlerOptions{
		Level:     GetlogLevel(severity),
		AddSource: severity == LogDebug,
	}))

	return logger
}

// GetlogLevel gets 'slog' level based on severity specified by user
func GetlogLevel(s LogSeverity) slog.Level {

	val, ok := logMap[s]
	if !ok {
		// default logger is Info
		return logMap[2]
	}

	return val
}
