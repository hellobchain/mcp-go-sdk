// Package config provides configuration management for the MCP server.
// It supports loading configuration from YAML files, CLI flags, and environment variables
// with the following priority: environment variables > CLI flags > config file > defaults.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// TokenValidator defines the interface for authentication token validation.
// Implementations verify tokens and return user identity and roles.
type TokenValidator interface {
	// ValidateToken checks if a token is valid and returns user info.
	// Returns (userID, roles, true) if valid, ("", nil, false) if invalid.
	ValidateToken(token string) (userID string, roles []string, ok bool)
}

// MapTokenValidator validates tokens against a static map.
// This is the default validator used when auth is enabled.
type MapTokenValidator struct {
	// Tokens maps token strings to user IDs
	Tokens map[string]string
}

// ValidateToken checks if the token exists in the map.
// All valid tokens are assigned the "user" role.
func (v *MapTokenValidator) ValidateToken(token string) (string, []string, bool) {
	userID, ok := v.Tokens[token]
	if !ok {
		return "", nil, false
	}
	return userID, []string{"user"}, true
}

// Load loads configuration from CLI flags and environment variables.
// This is the standard entry point for configuration loading.
func Load() (*Config, error) {
	return LoadFrom(flag.CommandLine, os.Args[1:])
}

// LoadFrom loads configuration with a custom flag set and arguments.
// This is primarily used for testing to avoid conflicts with test binary flags.
// 
// Loading sequence:
// 1. Start with default configuration
// 2. Scan args for -config flag to find config file path
// 3. Load config file if specified (overrides defaults)
// 4. Parse CLI flags (overrides config file)
// 5. Apply environment variables (overrides everything)
// 6. Validate final configuration
func LoadFrom(fs *flag.FlagSet, args []string) (*Config, error) {
	cfg := defaultConfig()

	// Scan for -config flag before full parsing
	// This allows config file to be loaded first
	configPath := envOrDefault("MCP_CONFIG_FILE", "")
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-config" || args[i] == "--config" {
			configPath = args[i+1]
			break
		}
	}

	// Load config file if specified
	if configPath != "" {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Convert custom types to strings for flag parsing
	// The flag package doesn't support custom types directly
	transportStr := string(cfg.Transport)
	logLevelStr := string(cfg.LogLevel)

	// Register all CLI flags with current config values as defaults
	// This allows flags to override config file values
	fs.StringVar(&transportStr, "transport", transportStr, "Transport mode: stdio or http")
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP listen address")
	fs.BoolVar(&cfg.AuthEnabled, "auth", cfg.AuthEnabled, "Enable authentication")
	fs.StringVar(&logLevelStr, "log-level", logLevelStr, "Log level: debug, info, warn, error")
	fs.Int64Var(&cfg.MaxBodySize, "max-body-size", cfg.MaxBodySize, "Max request body size in bytes")
	fs.StringVar(&cfg.ServerName, "server-name", cfg.ServerName, "MCP server name")
	fs.StringVar(&cfg.ServerVersion, "server-version", cfg.ServerVersion, "MCP server version")
	authTokensStr := fs.String("auth-tokens", "", "Comma-separated token:userID pairs")
	fs.BoolVar(&cfg.TracingEnabled, "tracing", cfg.TracingEnabled, "Enable OpenTelemetry tracing")
	fs.StringVar(&cfg.TracingEndpoint, "tracing-endpoint", cfg.TracingEndpoint, "OTLP HTTP endpoint")
	fs.BoolVar(&cfg.TracingInsecure, "tracing-insecure", cfg.TracingInsecure, "Use insecure connection for OTLP")
	fs.BoolVar(&cfg.MetricsEnabled, "metrics", cfg.MetricsEnabled, "Enable Prometheus metrics endpoint")
	fs.String("config", configPath, "Path to YAML config file")

	// Parse CLI flags (overrides config file values)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Convert string flags back to custom types
	cfg.Transport = Transport(transportStr)
	cfg.LogLevel = LogLevel(logLevelStr)

	// Parse auth tokens from CLI flag if provided
	if *authTokensStr != "" {
		cfg.AuthTokens = parseAuthTokens(*authTokensStr)
	}

	// Apply environment variables (highest priority, overrides everything)
	applyEnvVars(cfg)

	// Validate the final configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// defaultConfig returns a Config with all default values.
func defaultConfig() *Config {
	return &Config{
		// Core
		Transport:       DefaultTransport,
		Addr:            DefaultAddr,
		StdioEnabled:    false,
		AuthEnabled:     false,
		AuthTokens:      make(map[string]string),
		LogLevel:        DefaultLogLevel,
		MaxBodySize:     DefaultMaxBodySize,
		ServerName:      DefaultServerName,
		ServerVersion:   DefaultServerVersion,

		// Timeouts
		RequestTimeout:  DefaultRequestTimeout,
		ToolTimeout:     DefaultToolTimeout,
		ShutdownTimeout: DefaultShutdownTimeout,

		// Rate limiting
		RateLimitEnabled:       false,
		RequestsPerSecond:      DefaultRequestsPerSecond,
		BurstSize:              DefaultBurstSize,
		PerUserLimit:           DefaultPerUserLimit,
		PerToolLimit:           DefaultPerToolLimit,

		// CORS
		CORSEnabled:     false,
		CORSOrigins:     []string{"*"},
		CORSMethods:     []string{"GET", "POST", "DELETE"},
		CORSHeaders:     []string{"Content-Type", "Authorization", "X-Request-Id"},
		CORSCredentials: false,

		// Observability
		TracingEnabled:  false,
		TracingEndpoint: DefaultTracingEndpoint,
		TracingInsecure: true,
		MetricsEnabled:  false,
		AuditLogEnabled: false,
		PprofEnabled:    false,

		// Error tracking
		ErrorTrackingEnabled: false,
		ErrorTrackingDSN:     "",

		// Connection pool
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,

		// Advanced
		ConfigHotReload:  false,
		ConfigWatchPath:  "",
		PluginDir:        "",
		SecureConfigPath: "",
	}
}

// loadFromFile reads and parses a YAML configuration file.
// The file values are unmarshaled into the provided Config struct.
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

// applyEnvVars overrides config values with environment variables.
// Environment variables have the highest priority and override all other sources.
// Variable names follow the pattern: MCP_<CONFIG_KEY>
func applyEnvVars(cfg *Config) {
	if v := os.Getenv("MCP_TRANSPORT"); v != "" {
		cfg.Transport = Transport(v)
	}
	if v := os.Getenv("MCP_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("MCP_AUTH_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.AuthEnabled = b
		}
	}
	if v := os.Getenv("MCP_LOG_LEVEL"); v != "" {
		cfg.LogLevel = LogLevel(v)
	}
	if v := os.Getenv("MCP_MAX_BODY_SIZE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.MaxBodySize = n
		}
	}
	if v := os.Getenv("MCP_SERVER_NAME"); v != "" {
		cfg.ServerName = v
	}
	if v := os.Getenv("MCP_SERVER_VERSION"); v != "" {
		cfg.ServerVersion = v
	}
	if v := os.Getenv("MCP_AUTH_TOKENS"); v != "" {
		cfg.AuthTokens = parseAuthTokens(v)
	}
	if v := os.Getenv("MCP_TRACING_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.TracingEnabled = b
		}
	}
	if v := os.Getenv("MCP_TRACING_ENDPOINT"); v != "" {
		cfg.TracingEndpoint = v
	}
	if v := os.Getenv("MCP_TRACING_INSECURE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.TracingInsecure = b
		}
	}
	if v := os.Getenv("MCP_METRICS_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.MetricsEnabled = b
		}
	}
}

func (c *Config) Validate() error {
	// Transport validation
	if !c.Transport.Valid() {
		return fmt.Errorf("invalid transport: %s (must be %q, %q, or %q)", 
			c.Transport, TransportStdio, TransportHTTP, TransportMulti)
	}
	
	// HTTP/Multi mode requires address
	if (c.Transport == TransportHTTP || c.Transport == TransportMulti) && c.Addr == "" {
		return fmt.Errorf("listen address is required for HTTP/Multi transport")
	}
	
	// Log level validation
	if !c.LogLevel.Valid() {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	
	// Timeout validation
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request_timeout must be positive")
	}
	if c.ToolTimeout <= 0 {
		return fmt.Errorf("tool_timeout must be positive")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown_timeout must be positive")
	}
	
	// Rate limit validation
	if c.RateLimitEnabled {
		if c.RequestsPerSecond <= 0 {
			return fmt.Errorf("requests_per_second must be positive when rate limiting is enabled")
		}
		if c.BurstSize <= 0 {
			return fmt.Errorf("burst_size must be positive when rate limiting is enabled")
		}
	}
	
	// CORS validation
	if c.CORSEnabled {
		if len(c.CORSOrigins) == 0 {
			return fmt.Errorf("cors_origins cannot be empty when CORS is enabled")
		}
	}
	
	// Tracing validation
	if c.TracingEnabled && c.TracingEndpoint == "" {
		return fmt.Errorf("tracing_endpoint is required when tracing is enabled")
	}
	
	// Error tracking validation
	if c.ErrorTrackingEnabled && c.ErrorTrackingDSN == "" {
		return fmt.Errorf("error_tracking_dsn is required when error tracking is enabled")
	}
	
	// Connection pool validation
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns cannot be negative")
	}
	if c.MaxIdleConnsPerHost < 0 {
		return fmt.Errorf("max_idle_conns_per_host cannot be negative")
	}
	if c.IdleConnTimeout <= 0 {
		return fmt.Errorf("idle_conn_timeout must be positive")
	}
	
	// Config hot reload validation
	if c.ConfigHotReload && c.ConfigWatchPath == "" {
		return fmt.Errorf("config_watch_path is required when config_hot_reload is enabled")
	}
	
	// Auth validation
	if c.AuthEnabled && len(c.AuthTokens) == 0 {
		return fmt.Errorf("auth_tokens cannot be empty when auth is enabled")
	}
	
	return nil
}

func (c *Config) TokenValidator() TokenValidator {
	return &MapTokenValidator{Tokens: c.AuthTokens}
}

func parseAuthTokens(s string) map[string]string {
	tokens := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			tokens[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return tokens
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}

func envInt64OrDefault(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return fallback
}
