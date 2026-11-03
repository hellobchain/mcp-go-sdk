// Package middleware provides Gin HTTP middleware for the MCP server.
// Includes authentication, logging, rate limiting, request ID, tracing, and metrics.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/mcp-go-sdk/metrics"
)

// MetricsMiddleware returns a Gin middleware that records HTTP request metrics.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		metrics.RecordHTTPRequest(c.Request.Method, c.Request.URL.Path, status, duration)
	}
}
