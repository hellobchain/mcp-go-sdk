// Package transport provides HTTP and stdio transport implementations for the MCP server.
// HTTP transport wraps the StreamableHTTPServer in a Gin router, while stdio transport
// provides local command-line interface support with environment variable injection.
package transport

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

type HTTPTransport struct {
	httpServer *http.Server
	mcpHandler *server.StreamableHTTPServer
	router     *gin.Engine
	logger     *slog.Logger
}

func NewHTTPTransport(router *gin.Engine, mcpServer *server.MCPServer, addr string, maxBodySize int64, logger *slog.Logger) *HTTPTransport {
	mcpHandler := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithDisableLocalhostProtection(true),
		server.WithMaxRequestBodyBytes(maxBodySize),
	)

	router.Any("/mcp", func(c *gin.Context) {
		mcpHandler.ServeHTTP(c.Writer, c.Request)
	})

	return &HTTPTransport{
		httpServer: &http.Server{Addr: addr, Handler: router},
		mcpHandler: mcpHandler,
		router:     router,
		logger:     logger,
	}
}

func (t *HTTPTransport) Start() error {
	t.logger.Info("starting HTTP transport", "addr", t.httpServer.Addr)
	return t.httpServer.ListenAndServe()
}

func (t *HTTPTransport) Shutdown(ctx context.Context) error {
	if err := t.mcpHandler.Shutdown(ctx); err != nil {
		t.logger.Error("MCP handler shutdown error", "err", err)
	}
	return t.httpServer.Shutdown(ctx)
}
