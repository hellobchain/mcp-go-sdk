// Package app provides the main Server application for the MCP framework.
// This file implements structured startup information output.
package app

import (
	"encoding/json"
	"log/slog"

	"github.com/hellobchain/mcp-go-sdk/config"
)

// StartupInfo contains structured information about server startup configuration.
type StartupInfo struct {
	// Server identification
	ServerName    string `json:"server_name"`
	ServerVersion string `json:"server_version"`

	// Transport configuration
	Transport string `json:"transport"`
	Addr      string `json:"addr,omitempty"`

	// Feature flags
	AuthEnabled    bool `json:"auth_enabled"`
	TracingEnabled bool `json:"tracing_enabled"`
	MetricsEnabled bool `json:"metrics_enabled"`
	PprofEnabled   bool `json:"pprof_enabled"`
	CORSEnabled    bool `json:"cors_enabled"`

	// Timeouts
	RequestTimeout  string `json:"request_timeout"`
	ToolTimeout     string `json:"tool_timeout"`
	ShutdownTimeout string `json:"shutdown_timeout"`

	// Rate limiting
	RateLimitEnabled bool `json:"rate_limit_enabled"`
	RequestsPerSec   int  `json:"requests_per_second,omitempty"`

	// Logging
	LogLevel string `json:"log_level"`
}

// LogStartupInfo outputs structured startup information in JSON format.
// This makes it easy to parse and process startup configuration in logging systems.
func LogStartupInfo(cfg *config.Config, logger *slog.Logger) {
	info := StartupInfo{
		ServerName:    cfg.ServerName,
		ServerVersion: cfg.ServerVersion,
		Transport:     string(cfg.Transport),
		Addr:          cfg.Addr,
		AuthEnabled:   cfg.AuthEnabled,
		TracingEnabled: cfg.TracingEnabled,
		MetricsEnabled: cfg.MetricsEnabled,
		PprofEnabled:  cfg.PprofEnabled,
		CORSEnabled:   cfg.CORSEnabled,
		RequestTimeout:  cfg.RequestTimeout.String(),
		ToolTimeout:     cfg.ToolTimeout.String(),
		ShutdownTimeout: cfg.ShutdownTimeout.String(),
		RateLimitEnabled: cfg.RateLimitEnabled,
		RequestsPerSec:   cfg.RequestsPerSecond,
		LogLevel:      string(cfg.LogLevel),
	}

	// Convert to JSON for structured output
	jsonBytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		logger.Error("failed to marshal startup info", "error", err)
		return
	}

	logger.Info("server startup configuration", "config", string(jsonBytes))
}
