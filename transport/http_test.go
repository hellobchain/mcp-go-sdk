package transport

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

func TestNewHTTPTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	mcpSrv := server.NewMCPServer("test", "1.0")
	router := gin.New()

	ht := NewHTTPTransport(router, mcpSrv, ":0", 10*1024*1024, logger)
	if ht == nil {
		t.Fatal("NewHTTPTransport() returned nil")
	}
	if ht.httpServer == nil {
		t.Error("httpServer is nil")
	}
	if ht.mcpHandler == nil {
		t.Error("mcpHandler is nil")
	}
}

func TestHTTPTransport_StartAndShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	mcpSrv := server.NewMCPServer("test", "1.0")
	router := gin.New()

	ht := NewHTTPTransport(router, mcpSrv, ":0", 10*1024*1024, logger)

	errCh := make(chan error, 1)
	go func() {
		errCh <- ht.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ht.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("Start() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down in time")
	}
}
