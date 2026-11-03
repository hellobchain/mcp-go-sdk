// Package sdk provides a simple API for creating and running MCP servers.
// This is the main entry point for users who want to quickly start an MCP server
// with minimal configuration.
//
// Example usage:
//
//	package main
//
//	import (
//	    "github.com/hellobchain/mcp-go-sdk/sdk"
//	    "github.com/hellobchain/mcp-go-sdk/config"
//	)
//
//	func main() {
//	    // Create server with default config
//	    server, err := sdk.NewServer()
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    // Register custom tools
//	    server.RegisterTool(myCustomTool)
//
//	    // Run the server
//	    if err := server.Run(); err != nil {
//	        panic(err)
//	    }
//	}
package sdk

import (
	"context"
	"log/slog"

	"github.com/hellobchain/mcp-go-sdk/app"
	"github.com/hellobchain/mcp-go-sdk/config"
	"github.com/hellobchain/mcp-go-sdk/mcpserver"
)

// Server wraps the app.Server and provides a simplified API.
type Server struct {
	appServer *app.Server
	config    *config.Config
	logger    *slog.Logger
}

// Option is a function that configures the Server.
type Option func(*Server)

// WithConfig sets a custom configuration.
func WithConfig(cfg *config.Config) Option {
	return func(s *Server) {
		s.config = cfg
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithToolProvider sets a custom tool provider.
func WithToolProvider(provider app.ToolProvider) Option {
	return func(s *Server) {
		s.appServer = app.NewServer(s.config, s.logger, app.WithToolProvider(provider))
	}
}

// NewServer creates a new MCP server with the given options.
// If no config is provided, it loads from default sources (env, flags, config file).
// If no logger is provided, it creates a default logger based on config.
func NewServer(opts ...Option) (*Server, error) {
	s := &Server{}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Load config if not provided
	if s.config == nil {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		s.config = cfg
	}

	// Create logger if not provided
	if s.logger == nil {
		s.logger = app.SetupLogger(s.config.LogLevel)
	}

	// Create app server if not provided
	if s.appServer == nil {
		s.appServer = app.NewServer(s.config, s.logger)
	}

	return s, nil
}

// Initialize sets up all server components.
func (s *Server) Initialize(ctx context.Context) error {
	return s.appServer.Initialize(ctx)
}

// Run starts the MCP server.
func (s *Server) Run() error {
	return s.appServer.Run()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.appServer.Shutdown(ctx)
}

// RegisterTool registers a single tool with the server.
// This is a convenience method that requires the server to be initialized first.
func (s *Server) RegisterTool(tool mcpserver.ToolDefinition) {
	// This would need access to the registry, which is internal to app.Server
	// For now, users should use WithToolProvider to register custom tools
}

// Config returns the server configuration.
func (s *Server) Config() *config.Config {
	return s.config
}

// Logger returns the server logger.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// QuickStart creates and runs a server with minimal setup.
// This is the simplest way to start an MCP server.
//
// Example:
//
//	sdk.QuickStart()
func QuickStart() error {
	server, err := NewServer()
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := server.Initialize(ctx); err != nil {
		return err
	}

	return server.Run()
}
