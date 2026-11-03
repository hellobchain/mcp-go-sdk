package app

import (
	"log/slog"
	"os"

	"github.com/hellobchain/mcp-go-sdk/config"
)

// SetupLogger creates a structured JSON logger with the specified level.
func SetupLogger(level config.LogLevel) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case config.LogLevelDebug:
		logLevel = slog.LevelDebug
	case config.LogLevelWarn:
		logLevel = slog.LevelWarn
	case config.LogLevelError:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}
