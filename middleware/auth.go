package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/mcp-go-sdk/config"
)

type contextKey int

const authUserKey contextKey = iota

type AuthUser struct {
	UserID string
	Roles  []string
	Token  string
}

func UserFromContext(ctx context.Context) *AuthUser {
	if user, ok := ctx.Value(authUserKey).(*AuthUser); ok {
		return user
	}
	return nil
}

func AuthMiddleware(validator config.TokenValidator, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization scheme"})
			return
		}

		userID, roles, ok := validator.ValidateToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user := &AuthUser{UserID: userID, Roles: roles, Token: token}
		ctx := context.WithValue(c.Request.Context(), authUserKey, user)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
