package transport

import (
	"context"
	"os"
	"testing"
)

func TestCollectEnv(t *testing.T) {
	os.Setenv("MCP_TEST_VAR", "test_value")
	os.Setenv("MCP_GO_ANOTHER", "another_value")
	os.Setenv("UNRELATED_VAR", "unrelated")
	defer func() {
		os.Unsetenv("MCP_TEST_VAR")
		os.Unsetenv("MCP_GO_ANOTHER")
		os.Unsetenv("UNRELATED_VAR")
	}()

	env := collectEnv()

	if env["MCP_TEST_VAR"] != "test_value" {
		t.Errorf("MCP_TEST_VAR = %q, want %q", env["MCP_TEST_VAR"], "test_value")
	}
	if env["MCP_GO_ANOTHER"] != "another_value" {
		t.Errorf("MCP_GO_ANOTHER = %q, want %q", env["MCP_GO_ANOTHER"], "another_value")
	}
	if _, ok := env["UNRELATED_VAR"]; ok {
		t.Error("UNRELATED_VAR should not be collected")
	}
}

func TestIsRelevantEnvKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"MCP_TRANSPORT", true},
		{"MCP_ADDR", true},
		{"MCP_GO_CUSTOM", true},
		{"PATH", false},
		{"HOME", false},
		{"MCPX", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isRelevantEnvKey(tt.key); got != tt.want {
				t.Errorf("isRelevantEnvKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestEnvFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	env := EnvFromContext(ctx)
	if env != nil {
		t.Error("expected nil env from empty context")
	}
}

func TestEnvFromContext_WithValue(t *testing.T) {
	env := map[string]string{"MCP_TEST": "value"}
	ctx := context.WithValue(context.Background(), stdioEnvKey, env)

	got := EnvFromContext(ctx)
	if got["MCP_TEST"] != "value" {
		t.Errorf("EnvFromContext() = %v, want map with MCP_TEST=value", got)
	}
}

func TestStdioConfigFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	cfg := StdioConfigFromContext(ctx)
	if cfg != nil {
		t.Error("expected nil config from empty context")
	}
}

func TestStdioConfigFromContext_WithValue(t *testing.T) {
	cfg := &EnvConfig{
		Env:    map[string]string{"MCP_FOO": "bar"},
		Server: "stdio",
	}
	ctx := context.WithValue(context.Background(), stdioConfigKey, cfg)

	got := StdioConfigFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil config")
	}
	if got.Server != "stdio" {
		t.Errorf("Server = %q, want %q", got.Server, "stdio")
	}
	if got.Env["MCP_FOO"] != "bar" {
		t.Errorf("Env[MCP_FOO] = %q, want %q", got.Env["MCP_FOO"], "bar")
	}
}

func TestShutdownStdio(t *testing.T) {
	err := ShutdownStdio(context.Background())
	if err != nil {
		t.Errorf("ShutdownStdio() error = %v", err)
	}
}
