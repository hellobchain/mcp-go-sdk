package app

import (
	"log/slog"

	"github.com/hellobchain/mcp-go-sdk/config"
)

// StartupWarnings logs configuration warnings at startup.
func StartupWarnings(cfg *config.Config, logger *slog.Logger) {
	if cfg.Transport == config.TransportHTTP {
		if !cfg.AuthEnabled {
			logger.Warn("HTTP mode running WITHOUT authentication — not recommended for production")
		}
		if len(cfg.AuthTokens) == 0 && cfg.AuthEnabled {
			logger.Warn("auth enabled but no tokens configured — all requests will be rejected")
		}
	}
}
