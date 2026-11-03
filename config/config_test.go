package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid stdio config",
			cfg: Config{
				Transport:       TransportStdio,
				LogLevel:        "info",
				RequestTimeout:  30 * time.Second,
				ToolTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				IdleConnTimeout: 90 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid http config",
			cfg: Config{
				Transport:       TransportHTTP,
				Addr:            ":8080",
				LogLevel:        "debug",
				RequestTimeout:  30 * time.Second,
				ToolTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				IdleConnTimeout: 90 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid transport",
			cfg: Config{
				Transport:       "grpc",
				LogLevel:        "info",
				RequestTimeout:  30 * time.Second,
				ToolTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				IdleConnTimeout: 90 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "http without addr",
			cfg: Config{
				Transport:       TransportHTTP,
				Addr:            "",
				LogLevel:        "info",
				RequestTimeout:  30 * time.Second,
				ToolTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				IdleConnTimeout: 90 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: Config{
				Transport:       TransportStdio,
				LogLevel:        "verbose",
				RequestTimeout:  30 * time.Second,
				ToolTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				IdleConnTimeout: 90 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseAuthTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "single token",
			input: "abc123:alice",
			want:  map[string]string{"abc123": "alice"},
		},
		{
			name:  "multiple tokens",
			input: "abc123:alice,def456:bob",
			want:  map[string]string{"abc123": "alice", "def456": "bob"},
		},
		{
			name:  "empty string",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "with spaces",
			input: " abc123:alice , def456:bob ",
			want:  map[string]string{"abc123": "alice", "def456": "bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAuthTokens(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseAuthTokens() returned %d tokens, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseAuthTokens()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestMapTokenValidator(t *testing.T) {
	v := &MapTokenValidator{
		Tokens: map[string]string{"valid-token": "alice"},
	}

	userID, roles, ok := v.ValidateToken("valid-token")
	if !ok {
		t.Fatal("expected valid token to pass")
	}
	if userID != "alice" {
		t.Errorf("userID = %q, want %q", userID, "alice")
	}
	if len(roles) == 0 {
		t.Error("expected non-empty roles")
	}

	_, _, ok = v.ValidateToken("invalid-token")
	if ok {
		t.Error("expected invalid token to fail")
	}
}

func TestEnvOrDefault(t *testing.T) {
	os.Setenv("TEST_ENV_VAR_MCP", "from_env")
	defer os.Unsetenv("TEST_ENV_VAR_MCP")

	if got := envOrDefault("TEST_ENV_VAR_MCP", "fallback"); got != "from_env" {
		t.Errorf("envOrDefault() = %q, want %q", got, "from_env")
	}
	if got := envOrDefault("NONEXISTENT_VAR_MCP", "fallback"); got != "fallback" {
		t.Errorf("envOrDefault() = %q, want %q", got, "fallback")
	}
}

func TestEnvBoolOrDefault(t *testing.T) {
	os.Setenv("TEST_BOOL_VAR", "true")
	defer os.Unsetenv("TEST_BOOL_VAR")

	if got := envBoolOrDefault("TEST_BOOL_VAR", false); !got {
		t.Error("envBoolOrDefault() = false, want true")
	}
	if got := envBoolOrDefault("NONEXISTENT_BOOL", true); !got {
		t.Error("envBoolOrDefault() = false, want true (fallback)")
	}
}

func TestEnvInt64OrDefault(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "2048")
	defer os.Unsetenv("TEST_INT_VAR")

	if got := envInt64OrDefault("TEST_INT_VAR", 1024); got != 2048 {
		t.Errorf("envInt64OrDefault() = %d, want 2048", got)
	}
	if got := envInt64OrDefault("NONEXISTENT_INT", 1024); got != 1024 {
		t.Errorf("envInt64OrDefault() = %d, want 1024 (fallback)", got)
	}
}

func TestConfigTokenValidator(t *testing.T) {
	cfg := &Config{
		AuthTokens: map[string]string{"tok1": "user1"},
	}
	v := cfg.TokenValidator()
	userID, _, ok := v.ValidateToken("tok1")
	if !ok || userID != "user1" {
		t.Errorf("TokenValidator: ok=%v, userID=%q", ok, userID)
	}
}

func TestLoadFrom_Defaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := LoadFrom(fs, []string{})
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg.Transport != TransportStdio {
		t.Errorf("Transport = %q, want %q", cfg.Transport, TransportStdio)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.MaxBodySize != 10*1024*1024 {
		t.Errorf("MaxBodySize = %d, want %d", cfg.MaxBodySize, 10*1024*1024)
	}
	if cfg.ServerName != "mcp-go-server" {
		t.Errorf("ServerName = %q, want %q", cfg.ServerName, "mcp-go-server")
	}
	if cfg.AuthEnabled {
		t.Error("AuthEnabled should be false by default")
	}
	if cfg.TracingEnabled {
		t.Error("TracingEnabled should be false by default")
	}
	if cfg.MetricsEnabled {
		t.Error("MetricsEnabled should be false by default")
	}
}

func TestLoadFrom_Flags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := LoadFrom(fs, []string{
		"-transport", "http",
		"-addr", ":9090",
		"-auth",
		"-auth-tokens", "tok1:alice,tok2:bob",
		"-log-level", "debug",
		"-max-body-size", "2048",
		"-server-name", "my-server",
		"-server-version", "1.0.0",
		"-tracing",
		"-tracing-endpoint", "otel:4318",
		"-tracing-insecure=false",
		"-metrics",
	})
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg.Transport != TransportHTTP {
		t.Errorf("Transport = %q, want %q", cfg.Transport, TransportHTTP)
	}
	if cfg.Addr != ":9090" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":9090")
	}
	if !cfg.AuthEnabled {
		t.Error("AuthEnabled should be true")
	}
	if cfg.AuthTokens["tok1"] != "alice" {
		t.Errorf("AuthTokens[tok1] = %q, want %q", cfg.AuthTokens["tok1"], "alice")
	}
	if cfg.AuthTokens["tok2"] != "bob" {
		t.Errorf("AuthTokens[tok2] = %q, want %q", cfg.AuthTokens["tok2"], "bob")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.MaxBodySize != 2048 {
		t.Errorf("MaxBodySize = %d, want 2048", cfg.MaxBodySize)
	}
	if cfg.ServerName != "my-server" {
		t.Errorf("ServerName = %q, want %q", cfg.ServerName, "my-server")
	}
	if cfg.ServerVersion != "1.0.0" {
		t.Errorf("ServerVersion = %q, want %q", cfg.ServerVersion, "1.0.0")
	}
	if !cfg.TracingEnabled {
		t.Error("TracingEnabled should be true")
	}
	if cfg.TracingEndpoint != "otel:4318" {
		t.Errorf("TracingEndpoint = %q, want %q", cfg.TracingEndpoint, "otel:4318")
	}
	if cfg.TracingInsecure {
		t.Error("TracingInsecure should be false")
	}
	if !cfg.MetricsEnabled {
		t.Error("MetricsEnabled should be true")
	}
}

func TestLoadFrom_EnvVars(t *testing.T) {
	os.Setenv("MCP_TRANSPORT", "http")
	os.Setenv("MCP_ADDR", ":7070")
	os.Setenv("MCP_LOG_LEVEL", "warn")
	defer func() {
		os.Unsetenv("MCP_TRANSPORT")
		os.Unsetenv("MCP_ADDR")
		os.Unsetenv("MCP_LOG_LEVEL")
	}()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := LoadFrom(fs, []string{})
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg.Transport != TransportHTTP {
		t.Errorf("Transport = %q, want %q", cfg.Transport, TransportHTTP)
	}
	if cfg.Addr != ":7070" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":7070")
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "warn")
	}
}

func TestLoadFrom_ValidationError(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_, err := LoadFrom(fs, []string{"-transport", "invalid"})
	if err == nil {
		t.Error("expected validation error for invalid transport")
	}
}

func TestLoadFrom_ConfigFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `
transport: http
addr: ":9999"
auth_enabled: true
auth_tokens:
  "file-token": "file-user"
log_level: debug
max_body_size: 5242880
server_name: file-server
server_version: "2.0.0"
tracing_enabled: true
tracing_endpoint: "file-otel:4318"
tracing_insecure: false
metrics_enabled: true
`
	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	tmpFile.Close()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := LoadFrom(fs, []string{"-config", tmpFile.Name()})
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Transport != TransportHTTP {
		t.Errorf("Transport = %q, want %q", cfg.Transport, TransportHTTP)
	}
	if cfg.Addr != ":9999" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":9999")
	}
	if !cfg.AuthEnabled {
		t.Error("AuthEnabled should be true from file")
	}
	if cfg.AuthTokens["file-token"] != "file-user" {
		t.Errorf("AuthTokens[file-token] = %q, want %q", cfg.AuthTokens["file-token"], "file-user")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.MaxBodySize != 5242880 {
		t.Errorf("MaxBodySize = %d, want 5242880", cfg.MaxBodySize)
	}
	if cfg.ServerName != "file-server" {
		t.Errorf("ServerName = %q, want %q", cfg.ServerName, "file-server")
	}
	if cfg.ServerVersion != "2.0.0" {
		t.Errorf("ServerVersion = %q, want %q", cfg.ServerVersion, "2.0.0")
	}
	if !cfg.TracingEnabled {
		t.Error("TracingEnabled should be true from file")
	}
	if cfg.TracingEndpoint != "file-otel:4318" {
		t.Errorf("TracingEndpoint = %q, want %q", cfg.TracingEndpoint, "file-otel:4318")
	}
	if cfg.TracingInsecure {
		t.Error("TracingInsecure should be false from file")
	}
	if !cfg.MetricsEnabled {
		t.Error("MetricsEnabled should be true from file")
	}
}

func TestLoadFrom_PriorityOrder(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `
transport: stdio
addr: ":1111"
log_level: debug
server_name: file-server
`
	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	tmpFile.Close()

	os.Setenv("MCP_LOG_LEVEL", "warn")
	os.Setenv("MCP_SERVER_NAME", "env-server")
	defer func() {
		os.Unsetenv("MCP_LOG_LEVEL")
		os.Unsetenv("MCP_SERVER_NAME")
	}()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := LoadFrom(fs, []string{
		"-config", tmpFile.Name(),
		"-addr", ":2222",
		"-log-level", "error",
	})
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Transport != TransportStdio {
		t.Errorf("Transport = %q, want %q (from file)", cfg.Transport, TransportStdio)
	}
	if cfg.Addr != ":2222" {
		t.Errorf("Addr = %q, want %q (flag overrides file)", cfg.Addr, ":2222")
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want %q (env overrides flag)", cfg.LogLevel, "warn")
	}
	if cfg.ServerName != "env-server" {
		t.Errorf("ServerName = %q, want %q (env overrides file)", cfg.ServerName, "env-server")
	}
}

func TestLoadFrom_ConfigFileNotFound(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	_, err := LoadFrom(fs, []string{"-config", "/nonexistent/config.yaml"})
	if err == nil {
		t.Error("expected error for nonexistent config file")
	}
}

func TestLoadFrom_ConfigFileEnvVar(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	yamlContent := `
transport: http
addr: ":3333"
log_level: info
`
	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	tmpFile.Close()

	os.Setenv("MCP_CONFIG_FILE", tmpFile.Name())
	defer os.Unsetenv("MCP_CONFIG_FILE")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := LoadFrom(fs, []string{})
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	if cfg.Addr != ":3333" {
		t.Errorf("Addr = %q, want %q (from env-specified config file)", cfg.Addr, ":3333")
	}
}
