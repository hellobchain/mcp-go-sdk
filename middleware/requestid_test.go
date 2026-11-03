package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		rid := GetRequestID(c)
		if rid == "" {
			t.Error("expected request ID to be set")
		}
		c.String(http.StatusOK, rid)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	responseID := w.Body.String()
	if responseID == "" {
		t.Error("expected non-empty request ID in response")
	}

	headerID := w.Header().Get(RequestIDHeader)
	if headerID != responseID {
		t.Errorf("header ID %q != body ID %q", headerID, responseID)
	}
}

func TestRequestIDMiddleware_PreservesExisting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	var capturedID string
	router.GET("/test", func(c *gin.Context) {
		capturedID = GetRequestID(c)
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, "my-custom-id-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedID != "my-custom-id-123" {
		t.Errorf("request ID = %q, want %q", capturedID, "my-custom-id-123")
	}
}

func TestGetRequestID_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	rid := GetRequestID(c)
	if rid != "" {
		t.Errorf("expected empty request ID, got %q", rid)
	}
}

func TestGenerateRequestID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()
		if len(id) != 32 {
			t.Errorf("request ID length = %d, want 32", len(id))
		}
		if seen[id] {
			t.Errorf("duplicate request ID: %s", id)
		}
		seen[id] = true
	}
}
