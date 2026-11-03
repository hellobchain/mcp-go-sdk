package transport

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

type stdioContextKey int

const (
	stdioEnvKey stdioContextKey = iota
	stdioConfigKey
)

type EnvConfig struct {
	Env    map[string]string
	Server string
}

func ServeStdio(mcpServer *server.MCPServer, logger *slog.Logger) error {
	logger.Info("starting stdio transport")

	contextFunc := func(ctx context.Context) context.Context {
		env := collectEnv()
		cfg := &EnvConfig{
			Env:    env,
			Server: "stdio",
		}
		ctx = context.WithValue(ctx, stdioEnvKey, env)
		ctx = context.WithValue(ctx, stdioConfigKey, cfg)
		logger.Debug("stdio context injected", "env_vars", len(env))
		return ctx
	}

	return server.ServeStdio(mcpServer, server.WithStdioContextFunc(contextFunc))
}

func ShutdownStdio(_ context.Context) error {
	return nil
}

func EnvFromContext(ctx context.Context) map[string]string {
	if v, ok := ctx.Value(stdioEnvKey).(map[string]string); ok {
		return v
	}
	return nil
}

func StdioConfigFromContext(ctx context.Context) *EnvConfig {
	if v, ok := ctx.Value(stdioConfigKey).(*EnvConfig); ok {
		return v
	}
	return nil
}

func collectEnv() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if isRelevantEnvKey(key) {
				env[key] = parts[1]
			}
		}
	}
	return env
}

func isRelevantEnvKey(key string) bool {
	prefixes := []string{"MCP_", "MCP_GO_"}
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
