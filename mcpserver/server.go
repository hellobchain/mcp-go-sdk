// Package mcpserver provides MCP server wrapper with hooks for observability.
// It wraps the mark3labs/mcp-go SDK and adds logging and metrics collection
// for tool calls through the hooks system.
package mcpserver

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/hellobchain/mcp-go-sdk/metrics"
)

type ToolHandler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

type ToolDefinition struct {
	Name        string
	Description string
	Handler     ToolHandler
	Options     []mcp.ToolOption
}

type MCPServer struct {
	mcpServer *server.MCPServer
	hooks     *server.Hooks
	logger    *slog.Logger
}

func NewMCPServer(name, version string, logger *slog.Logger) *MCPServer {
	hooks := &server.Hooks{}

	mcpSrv := server.NewMCPServer(name, version,
		server.WithRecovery(),
		server.WithHooks(hooks),
	)

	s := &MCPServer{
		mcpServer: mcpSrv,
		hooks:     hooks,
		logger:    logger,
	}

	s.registerToolCallHooks()

	return s
}

func (s *MCPServer) registerToolCallHooks() {
	startTimes := &contextKeyMap{}

	s.hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		startTimes.Set(ctx, time.Now())
		s.logger.Info("tool call start",
			"tool", message.Params.Name,
			"request_id", id,
		)
	})

	s.hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result any) {
		startTime := startTimes.Get(ctx)
		elapsed := time.Since(startTime)

		isError := false
		if r, ok := result.(*mcp.CallToolResult); ok {
			isError = r.IsError
		}

		s.logger.Info("tool call complete",
			"tool", message.Params.Name,
			"request_id", id,
			"is_error", isError,
			"duration_ms", elapsed.Milliseconds(),
		)

		metrics.RecordToolCall(message.Params.Name, isError, elapsed)
	})

	s.hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		if method == mcp.MethodToolsCall {
			if req, ok := message.(*mcp.CallToolRequest); ok {
				s.logger.Error("tool call error",
					"tool", req.Params.Name,
					"request_id", id,
					"err", err,
				)
			}
		}
	})
}

func (s *MCPServer) AddTool(tool ToolDefinition) {
	mcpTool := mcp.NewTool(tool.Name, tool.Options...)
	s.mcpServer.AddTool(mcpTool, server.ToolHandlerFunc(tool.Handler))
	s.logger.Info("tool registered", "name", tool.Name)
}

func (s *MCPServer) Server() *server.MCPServer {
	return s.mcpServer
}

type contextKeyMap struct {
	m sync.Map
}

func (c *contextKeyMap) Set(ctx context.Context, v any) {
	c.m.Store(ctx, v)
}

func (c *contextKeyMap) Get(ctx context.Context) time.Time {
	if v, ok := c.m.Load(ctx); ok {
		return v.(time.Time)
	}
	return time.Now()
}
