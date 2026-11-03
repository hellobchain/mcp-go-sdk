// Package app provides the main Server application for the MCP framework.
// It orchestrates initialization, runtime, and shutdown of all components.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/hellobchain/mcp-go-sdk/config"
	"github.com/hellobchain/mcp-go-sdk/mcpserver"
	"github.com/hellobchain/mcp-go-sdk/middleware"
	"github.com/hellobchain/mcp-go-sdk/tools"
	"github.com/hellobchain/mcp-go-sdk/tracing"
	"github.com/hellobchain/mcp-go-sdk/transport"
)

var serverStartTime = time.Now()

// Runner defines the interface for server lifecycle management.
// Implementations must support initialization, running, and graceful shutdown.
type Runner interface {
	// Initialize sets up all server components (tracing, tools, MCP server).
	Initialize(ctx context.Context) error
	// Run starts the server in the configured transport mode (stdio or HTTP).
	Run() error
	// Shutdown gracefully stops the server and cleans up resources.
	Shutdown(ctx context.Context) error
}

// ToolProvider defines the interface for tool registration.
// Implementations can provide custom sets of tools to register with the MCP server.
type ToolProvider interface {
	// RegisterTools adds tools to the provided registry.
	RegisterTools(registry *mcpserver.Registry)
}

// DefaultToolProvider provides the default set of example tools.
// It registers echo, time, and divide tools for demonstration purposes.
type DefaultToolProvider struct{}

// RegisterTools registers the default set of tools.
// This includes:
//   - echo: Returns the input message (demonstrates basic tool)
//   - time: Returns current time with auth context (demonstrates context usage)
//   - divide: Performs division with error handling (demonstrates IsError pattern)
func (p *DefaultToolProvider) RegisterTools(registry *mcpserver.Registry) {
	registry.Register(tools.NewEchoTool())
	registry.Register(tools.NewTimeTool())
	registry.Register(tools.NewDivideTool())
}

// ServerOption is a function that configures a Server.
// Used in the builder pattern to customize server behavior.
type ServerOption func(*Server)

// WithToolProvider sets a custom tool provider.
// Use this to register custom tools instead of the default set.
// Example:
//   server := NewServer(cfg, logger, WithToolProvider(&MyCustomTools{}))
func WithToolProvider(provider ToolProvider) ServerOption {
	return func(s *Server) {
		s.toolProvider = provider
	}
}

// WithTracingProvider sets a custom tracing provider.
// Use this to inject a pre-configured tracer provider (e.g., for testing).
func WithTracingProvider(tp *tracing.Provider) ServerOption {
	return func(s *Server) {
		s.tp = tp
	}
}

// Server encapsulates the MCP server application.
// It holds all dependencies and manages the server lifecycle.
type Server struct {
	cfg           *config.Config       // Server configuration
	logger        *slog.Logger         // Structured logger
	registry      *mcpserver.Registry  // Tool registry
	mcpSrv        *mcpserver.MCPServer // MCP protocol server
	tp            *tracing.Provider    // OpenTelemetry tracer provider
	httpTransport *transport.HTTPTransport // HTTP transport (nil for stdio mode)
	toolProvider  ToolProvider         // Tool registration strategy
}

// NewServer creates a new Server instance with optional configuration.
// The server is initialized with default tool provider unless overridden.
// Options are applied in order, allowing customization of server behavior.
func NewServer(cfg *config.Config, logger *slog.Logger, opts ...ServerOption) *Server {
	s := &Server{
		cfg:          cfg,
		logger:       logger,
		toolProvider: &DefaultToolProvider{},
	}

	// Apply all options to customize the server
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Initialize sets up all server components in the correct order.
// Initialization sequence:
//  1. Tracing provider (OpenTelemetry)
//  2. Tool registry with registered tools
//  3. MCP server with tool bindings
//
// Returns error if any component fails to initialize.
func (s *Server) Initialize(ctx context.Context) error {
	if err := s.initTracing(ctx); err != nil {
		return err
	}

	s.initToolRegistry()
	s.initMCPServer()

	return nil
}

// initTracing initializes the OpenTelemetry tracing provider.
// If a tracing provider was injected via WithTracingProvider, this is a no-op.
// Otherwise, creates a new provider based on configuration.
func (s *Server) initTracing(ctx context.Context) error {
	// Skip initialization if provider was injected
	if s.tp != nil {
		return nil
	}

	tp, err := tracing.Init(ctx, tracing.Config{
		Enabled:     s.cfg.TracingEnabled,
		ServiceName: s.cfg.ServerName,
		Version:     s.cfg.ServerVersion,
		Endpoint:    s.cfg.TracingEndpoint,
		Insecure:    s.cfg.TracingInsecure,
		Logger:      s.logger,
	})
	if err != nil {
		return err
	}
	s.tp = tp
	return nil
}

// initToolRegistry creates a new tool registry and registers all tools.
// The tool provider (default or custom) determines which tools are available.
func (s *Server) initToolRegistry() {
	s.registry = mcpserver.NewRegistry()
	s.toolProvider.RegisterTools(s.registry)
}

// initMCPServer creates the MCP protocol server and binds all registered tools.
// This server handles the JSON-RPC protocol and tool invocation.
func (s *Server) initMCPServer() {
	s.mcpSrv = mcpserver.NewMCPServer(s.cfg.ServerName, s.cfg.ServerVersion, s.logger)
	s.registry.RegisterAll(s.mcpSrv)
}

// Run starts the server in the configured transport mode.
// For stdio mode, this blocks until the connection is closed.
// For HTTP mode, this starts the server in a goroutine and returns immediately.
// For multi mode, this starts both stdio and HTTP servers simultaneously.
func (s *Server) Run() error {
	switch s.cfg.Transport {
	case config.TransportStdio:
		return s.runStdio()
	case config.TransportHTTP:
		return s.runHTTP()
	case config.TransportMulti:
		return s.runMulti()
	default:
		return fmt.Errorf("invalid transport mode: %s", s.cfg.Transport)
	}
}

// runStdio starts the server in stdio mode for local client communication.
// This is a blocking call that runs until stdin is closed or an error occurs.
func (s *Server) runStdio() error {
	return transport.ServeStdio(s.mcpSrv.Server(), s.logger)
}

// runHTTP starts the server in HTTP mode for remote client communication.
// The server runs in a goroutine and errors are sent to a channel.
// Returns immediately after startup, or returns error if startup fails.
func (s *Server) runHTTP() error {
	router := s.setupRouter()
	s.httpTransport = transport.NewHTTPTransport(router, s.mcpSrv.Server(), s.cfg.Addr, s.cfg.MaxBodySize, s.logger)

	// Start server in goroutine and capture errors
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpTransport.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	s.logger.Info("MCP server started",
		"transport", "http",
		"addr", s.cfg.Addr,
		"auth", s.cfg.AuthEnabled,
		"tracing", s.cfg.TracingEnabled,
		"metrics", s.cfg.MetricsEnabled,
		"version", s.cfg.ServerVersion,
	)

	// Wait briefly to catch immediate startup errors
	select {
	case err := <-errCh:
		return err
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// runMulti starts the server in multi-transport mode (stdio + HTTP simultaneously).
// The HTTP server runs in a goroutine while stdio blocks the main thread.
// Returns error if HTTP server fails to start.
func (s *Server) runMulti() error {
	// Start HTTP server in goroutine
	if err := s.runHTTP(); err != nil {
		return err
	}

	s.logger.Info("Multi-transport mode: stdio + HTTP running simultaneously")

	// Run stdio server in main thread (blocking)
	return s.runStdio()
}

// setupRouter creates and configures the Gin HTTP router.
// Applies middleware in the correct order and registers all routes.
func (s *Server) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Apply middleware in order: core -> cors -> observability -> auth
	s.setupCoreMiddleware(router)
	s.setupCORS(router)
	s.setupObservabilityMiddleware(router)
	s.setupAuthMiddleware(router)
	s.setupRoutes(router)

	return router
}

// setupCoreMiddleware adds essential middleware that always runs.
// Order matters: Recovery -> Timeout -> RequestID -> Logger
func (s *Server) setupCoreMiddleware(router *gin.Engine) {
	router.Use(gin.Recovery())                                    // Panic recovery
	router.Use(middleware.TimeoutMiddleware(s.cfg.RequestTimeout)) // Request timeout
	router.Use(middleware.RequestIDMiddleware())                   // Generate/preserve request ID
	router.Use(middleware.LoggerMiddleware(s.logger))              // Structured request logging
}

// setupObservabilityMiddleware adds tracing and metrics middleware when enabled.
// These are optional and controlled by configuration.
func (s *Server) setupObservabilityMiddleware(router *gin.Engine) {
	if s.cfg.TracingEnabled {
		router.Use(middleware.TracingMiddleware())
	}
	if s.cfg.MetricsEnabled {
		router.Use(middleware.MetricsMiddleware())
	}
}

// setupAuthMiddleware adds authentication middleware when enabled.
// Logs a warning if auth is disabled in HTTP mode (security risk).
func (s *Server) setupAuthMiddleware(router *gin.Engine) {
	if s.cfg.AuthEnabled {
		router.Use(middleware.AuthMiddleware(s.cfg.TokenValidator(), true))
	} else {
		s.logger.Warn("authentication is DISABLED — all requests are accepted without verification")
	}
}

// setupCORS adds CORS middleware when enabled.
func (s *Server) setupCORS(router *gin.Engine) {
	if s.cfg.CORSEnabled {
		router.Use(middleware.CORSMiddleware(middleware.CORSConfig{
			AllowedOrigins:   s.cfg.CORSOrigins,
			AllowedMethods:   s.cfg.CORSMethods,
			AllowedHeaders:   s.cfg.CORSHeaders,
			AllowCredentials: s.cfg.CORSCredentials,
		}))
		s.logger.Info("CORS enabled", "origins", s.cfg.CORSOrigins)
	}
}

// setupRoutes registers all HTTP endpoint handlers.
// Always registers /healthz and /status, conditionally registers /metrics and /debug/pprof/*.
func (s *Server) setupRoutes(router *gin.Engine) {
	// Health and status endpoints (always available)
	router.GET("/healthz", s.healthzHandler)
	router.GET("/status", s.statusHandler)

	// Metrics endpoint (when enabled)
	if s.cfg.MetricsEnabled {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
		s.logger.Info("Prometheus metrics enabled at /metrics")
	}

	// Pprof endpoints (when enabled)
	if s.cfg.PprofEnabled {
		pprofGroup := router.Group("/debug/pprof")
		{
			pprofGroup.GET("/", gin.WrapF(pprof.Index))
			pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
			pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
			pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
			pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
			pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
			pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
			pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
			pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
			pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		}
		s.logger.Info("Pprof endpoints enabled at /debug/pprof")
	}
}

// healthzHandler returns server health status with dependency checks.
// Includes version, transport mode, uptime, and status of dependencies (tracing, metrics, auth).
func (s *Server) healthzHandler(c *gin.Context) {
	uptime := time.Since(serverStartTime)

	// Check dependency status
	dependencies := gin.H{
		"tracing": s.cfg.TracingEnabled,
		"metrics": s.cfg.MetricsEnabled,
		"auth":    s.cfg.AuthEnabled,
	}

	// Determine overall health status
	status := "ok"
	httpStatus := http.StatusOK

	// Check if server has been running for too short a time (still initializing)
	if uptime.Seconds() < 2 {
		status = "starting"
	}

	c.JSON(httpStatus, gin.H{
		"status":       status,
		"version":      s.cfg.ServerVersion,
		"transport":    s.cfg.Transport,
		"uptime_s":     int(uptime.Seconds()),
		"dependencies": dependencies,
	})
}

// statusHandler returns detailed runtime status information.
// Includes memory usage, goroutine count, and server configuration.
func (s *Server) statusHandler(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.JSON(http.StatusOK, gin.H{
		"status":      "running",
		"version":     s.cfg.ServerVersion,
		"server_name": s.cfg.ServerName,
		"transport":   s.cfg.Transport,
		"addr":        s.cfg.Addr,
		"uptime_s":    int(time.Since(serverStartTime).Seconds()),

		// Runtime statistics
		"goroutines":    runtime.NumGoroutine(),
		"memory_alloc":  memStats.Alloc / 1024 / 1024,      // MB
		"memory_sys":    memStats.Sys / 1024 / 1024,         // MB
		"memory_gc_runs": memStats.NumGC,

		// Configuration
		"auth_enabled":    s.cfg.AuthEnabled,
		"tracing_enabled": s.cfg.TracingEnabled,
		"metrics_enabled": s.cfg.MetricsEnabled,
		"pprof_enabled":   s.cfg.PprofEnabled,
		"log_level":       s.cfg.LogLevel,
	})
}

// Shutdown gracefully stops the server with configurable timeout.
// It stops accepting new connections, waits for in-flight requests to complete,
// then closes all resources (tracing, HTTP transport).
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("initiating graceful shutdown", "timeout", s.cfg.ShutdownTimeout)

	// Create a context with the configured shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()

	// Shutdown tracing provider first to flush any pending spans
	if s.tp != nil {
		if err := s.tp.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("tracing shutdown error", "err", err)
		}
	}

	// Shutdown HTTP transport (drains connections)
	if s.httpTransport != nil {
		if err := s.httpTransport.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("HTTP transport shutdown error", "err", err)
			return err
		}
	}

	s.logger.Info("graceful shutdown complete")
	return nil
}

// WaitForShutdown returns a channel that receives OS shutdown signals.
func (s *Server) WaitForShutdown() <-chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	return quit
}
