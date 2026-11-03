package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hellobchain/mcp-go-sdk/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockValidator struct {
	tokens map[string]string
}

func (v *mockValidator) ValidateToken(token string) (string, []string, bool) {
	userID, ok := v.tokens[token]
	if !ok {
		return "", nil, false
	}
	return userID, []string{"user"}, true
}

func TestAuthMiddleware_NoAuth(t *testing.T) {
	router := gin.New()
	validator := &mockValidator{tokens: map[string]string{"valid": "alice"}}
	router.Use(AuthMiddleware(validator, false))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := gin.New()
	validator := &mockValidator{tokens: map[string]string{"valid": "alice"}}
	router.Use(AuthMiddleware(validator, true))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := gin.New()
	validator := &mockValidator{tokens: map[string]string{"valid-token": "alice"}}
	router.Use(AuthMiddleware(validator, true))

	var capturedUser *AuthUser
	router.GET("/test", func(c *gin.Context) {
		capturedUser = UserFromContext(c.Request.Context())
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if capturedUser == nil {
		t.Fatal("expected user in context")
	}
	if capturedUser.UserID != "alice" {
		t.Errorf("userID = %q, want %q", capturedUser.UserID, "alice")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	validator := &mockValidator{tokens: map[string]string{"valid": "alice"}}
	router.Use(AuthMiddleware(validator, true))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_WrongScheme(t *testing.T) {
	router := gin.New()
	validator := &mockValidator{tokens: map[string]string{"valid": "alice"}}
	router.Use(AuthMiddleware(validator, true))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic wrong-scheme")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUserFromContext_Nil(t *testing.T) {
	user := UserFromContext(context.Background())
	if user != nil {
		t.Error("expected nil user from empty context")
	}
}

func TestLoggerMiddleware(t *testing.T) {
	logger := slog.Default()
	router := gin.New()
	router.Use(LoggerMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestLoggerMiddleware_SkipsHealthz(t *testing.T) {
	logger := slog.Default()
	router := gin.New()
	router.Use(LoggerMiddleware(logger))
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestMaskSensitiveData(t *testing.T) {
	data := map[string]any{
		"username":   "alice",
		"password":   "secret123",
		"api_key":    "key-abc",
		"message":    "hello",
		"auth_token": "tok-123",
	}

	masked := MaskSensitiveData(data, defaultSensitiveFields)

	if masked["username"] != "alice" {
		t.Errorf("username should not be masked, got %v", masked["username"])
	}
	if masked["message"] != "hello" {
		t.Errorf("message should not be masked, got %v", masked["message"])
	}
	if masked["password"] != "***" {
		t.Errorf("password should be maskeded, got %v", masked["password"])
	}
	if masked["api_key"] != "***" {
		t.Errorf("api_key should be masked, got %v", masked["api_key"])
	}
	if masked["auth_token"] != "***" {
		t.Errorf("auth_token should be masked, got %v", masked["auth_token"])
	}
}

func TestIsSensitive(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"password", true},
		{"user_password_hash", true},
		{"Authorization", true},
		{"api_key", true},
		{"apikey", true},
		{"secret", true},
		{"token", true},
		{"username", false},
		{"message", false},
		{"id", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isSensitive(tt.key, defaultSensitiveFields); got != tt.want {
				t.Errorf("isSensitive(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(3, time.Second)

	for i := 0; i < 3; i++ {
		if !limiter.Allow("test-ip") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	if limiter.Allow("test-ip") {
		t.Error("4th request should be rate limited")
	}

	if !limiter.Allow("other-ip") {
		t.Error("request from different IP should be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)
	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("rate limited request: status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestAuthMiddleware_WithConfigValidator(t *testing.T) {
	router := gin.New()
	validator := &config.MapTokenValidator{
		Tokens: map[string]string{"cfg-token": "cfg-user"},
	}
	router.Use(AuthMiddleware(validator, true))

	var capturedUser *AuthUser
	router.GET("/test", func(c *gin.Context) {
		capturedUser = UserFromContext(c.Request.Context())
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer cfg-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if capturedUser == nil || capturedUser.UserID != "cfg-user" {
		t.Errorf("user = %v, want cfg-user", capturedUser)
	}
}
