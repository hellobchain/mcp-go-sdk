package config

import "time"

const (
	DefaultAddr            = ":8080"
	DefaultMaxBodySize     = int64(10 * 1024 * 1024)
	DefaultServerName      = "mcp-go-server"
	DefaultServerVersion   = "0.1.0"
	DefaultTracingEndpoint = "localhost:4318"
	DefaultLogLevel        = LogLevelInfo
	DefaultTransport       = TransportStdio

	// Timeout defaults
	DefaultRequestTimeout = 30 * time.Second
	DefaultToolTimeout    = 60 * time.Second
	DefaultShutdownTimeout = 30 * time.Second

	// Rate limit defaults
	DefaultRequestsPerSecond = 100
	DefaultBurstSize         = 200
	DefaultPerUserLimit      = 50
	DefaultPerToolLimit      = 30

	// Connection pool defaults
	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 10
	DefaultIdleConnTimeout     = 90 * time.Second
)

type Transport string

const (
	TransportStdio Transport = "stdio"
	TransportHTTP  Transport = "http"
	TransportMulti Transport = "multi" // Run both stdio and HTTP
)

func (t Transport) Valid() bool {
	return t == TransportStdio || t == TransportHTTP || t == TransportMulti
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

func (l LogLevel) Valid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return true
	}
	return false
}

// Config holds all configuration for the MCP server.
type Config struct {
	// Core settings
	Transport     Transport         `yaml:"transport" json:"transport"`
	Addr          string            `yaml:"addr" json:"addr"`
	StdioEnabled  bool              `yaml:"stdio_enabled" json:"stdio_enabled"` // For multi mode
	AuthEnabled   bool              `yaml:"auth_enabled" json:"auth_enabled"`
	AuthTokens    map[string]string `yaml:"auth_tokens" json:"auth_tokens"`
	LogLevel      LogLevel          `yaml:"log_level" json:"log_level"`
	MaxBodySize   int64             `yaml:"max_body_size" json:"max_body_size"`
	ServerName    string            `yaml:"server_name" json:"server_name"`
	ServerVersion string            `yaml:"server_version" json:"server_version"`

	// Timeout settings
	RequestTimeout  time.Duration `yaml:"request_timeout" json:"request_timeout"`
	ToolTimeout     time.Duration `yaml:"tool_timeout" json:"tool_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`

	// Rate limiting
	RateLimitEnabled       bool `yaml:"rate_limit_enabled" json:"rate_limit_enabled"`
	RequestsPerSecond      int  `yaml:"requests_per_second" json:"requests_per_second"`
	BurstSize              int  `yaml:"burst_size" json:"burst_size"`
	PerUserLimit           int  `yaml:"per_user_limit" json:"per_user_limit"`
	PerToolLimit           int  `yaml:"per_tool_limit" json:"per_tool_limit"`

	// CORS
	CORSEnabled      bool     `yaml:"cors_enabled" json:"cors_enabled"`
	CORSOrigins      []string `yaml:"cors_origins" json:"cors_origins"`
	CORSMethods      []string `yaml:"cors_methods" json:"cors_methods"`
	CORSHeaders      []string `yaml:"cors_headers" json:"cors_headers"`
	CORSCredentials  bool     `yaml:"cors_credentials" json:"cors_credentials"`

	// Observability
	TracingEnabled    bool   `yaml:"tracing_enabled" json:"tracing_enabled"`
	TracingEndpoint   string `yaml:"tracing_endpoint" json:"tracing_endpoint"`
	TracingInsecure   bool   `yaml:"tracing_insecure" json:"tracing_insecure"`
	MetricsEnabled    bool   `yaml:"metrics_enabled" json:"metrics_enabled"`
	AuditLogEnabled   bool   `yaml:"audit_log_enabled" json:"audit_log_enabled"`
	PprofEnabled      bool   `yaml:"pprof_enabled" json:"pprof_enabled"`

	// Error tracking
	ErrorTrackingEnabled bool   `yaml:"error_tracking_enabled" json:"error_tracking_enabled"`
	ErrorTrackingDSN     string `yaml:"error_tracking_dsn" json:"error_tracking_dsn"`

	// Connection pool
	MaxIdleConns        int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" json:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout" json:"idle_conn_timeout"`

	// Advanced features
	ConfigHotReload   bool   `yaml:"config_hot_reload" json:"config_hot_reload"`
	ConfigWatchPath   string `yaml:"config_watch_path" json:"config_watch_path"`
	PluginDir         string `yaml:"plugin_dir" json:"plugin_dir"`
	SecureConfigPath  string `yaml:"secure_config_path" json:"secure_config_path"`
}
