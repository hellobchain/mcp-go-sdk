package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var defaultSensitiveFields = []string{"token", "password", "secret", "authorization", "api_key", "apikey"}

func LoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return LoggerMiddlewareWithSensitiveFields(logger, defaultSensitiveFields)
}

func LoggerMiddlewareWithSensitiveFields(logger *slog.Logger, sensitiveFields []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		if path == "/healthz" {
			c.Next()
			return
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID := GetRequestID(c)

		logger.Info("request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func MaskSensitiveData(data map[string]any, sensitiveFields []string) map[string]any {
	masked := make(map[string]any, len(data))
	for k, v := range data {
		if isSensitive(k, sensitiveFields) {
			masked[k] = "***"
		} else {
			masked[k] = v
		}
	}
	return masked
}

func isSensitive(key string, sensitiveFields []string) bool {
	lower := strings.ToLower(key)
	for _, field := range sensitiveFields {
		if strings.Contains(lower, field) {
			return true
		}
	}
	return false
}
