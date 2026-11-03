// Package middleware provides HTTP middleware for the MCP server.
// This file implements CORS (Cross-Origin Resource Sharing) middleware.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration settings.
type CORSConfig struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// Use ["*"] to allow any origin.
	AllowedOrigins []string
	// AllowedMethods is a list of methods the client is allowed to use.
	AllowedMethods []string
	// AllowedHeaders is a list of headers the client is allowed to use.
	AllowedHeaders []string
	// AllowCredentials indicates whether the request can include user credentials.
	AllowCredentials bool
}

// CORSMiddleware creates a CORS middleware with the given configuration.
// It handles preflight requests and sets appropriate headers for cross-origin requests.
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	// Default methods if not specified
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	// Default headers if not specified
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Origin", "Content-Type", "Authorization"}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range config.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if !allowed {
			// Origin not allowed, continue without CORS headers
			c.Next()
			return
		}

		// Set CORS headers
		if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight request
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
