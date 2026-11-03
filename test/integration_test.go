package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/hellobchain/mcp-go-sdk/config"
	"github.com/hellobchain/mcp-go-sdk/mcpserver"
	"github.com/hellobchain/mcp-go-sdk/middleware"
	"github.com/hellobchain/mcp-go-sdk/tools"
	"github.com/hellobchain/mcp-go-sdk/transport"
)

func setupTestRouter(t *testing.T, authEnabled bool) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	registry := mcpserver.NewRegistry()
	registry.Register(tools.NewEchoTool())
	registry.Register(tools.NewTimeTool())
	registry.Register(tools.NewDivideTool())

	mcpSrv := mcpserver.NewMCPServer("test-server", "0.1.0", logger)
	registry.RegisterAll(mcpSrv)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))

	if authEnabled {
		validator := &config.MapTokenValidator{
			Tokens: map[string]string{"test-token": "test-user"},
		}
		router.Use(middleware.AuthMiddleware(validator, true))
	}

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"version":   "0.1.0",
			"transport": "http",
			"uptime_s":  0,
		})
	})

	transport.NewHTTPTransport(router, mcpSrv.Server(), ":0", 10*1024*1024, logger)

	return router
}

func doMCPPost(router *gin.Engine, payload map[string]any, sessionID string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func initializeSession(t *testing.T, router *gin.Engine) string {
	t.Helper()

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "test-client", "version": "1.0.0"},
		},
	}

	w := doMCPPost(router, initReq, "")
	if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
		t.Fatalf("initialize: status = %d, body = %s", w.Code, w.Body.String())
	}

	sessionID := w.Header().Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("initialize: no Mcp-Session-Id in response")
	}

	notifReq := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	_ = doMCPPost(router, notifReq, sessionID)

	return sessionID
}

func TestHealthzEndpoint(t *testing.T) {
	router := setupTestRouter(t, false)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
}

func TestMCPInitialize(t *testing.T) {
	router := setupTestRouter(t, false)

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	w := doMCPPost(router, initReq, "")

	if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 200 or 202", w.Code)
		t.Logf("body: %s", w.Body.String())
	}

	sessionID := w.Header().Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Error("expected Mcp-Session-Id header in response")
	}
}

func TestMCPToolsList(t *testing.T) {
	router := setupTestRouter(t, false)
	sessionID := initializeSession(t, router)

	listReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}
	w := doMCPPost(router, listReq, sessionID)

	if w.Code != http.StatusOK {
		t.Errorf("tools/list status = %d, want 200", w.Code)
		t.Logf("body: %s", w.Body.String())
		return
	}

	var resp struct {
		Result struct {
			Tools []mcp.Tool `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v\nbody: %s", err, w.Body.String())
	}

	if len(resp.Result.Tools) < 2 {
		t.Errorf("expected at least 2 tools, got %d", len(resp.Result.Tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range resp.Result.Tools {
		toolNames[tool.Name] = true
	}
	if !toolNames["echo"] {
		t.Error("missing 'echo' tool")
	}
	if !toolNames["time"] {
		t.Error("missing 'time' tool")
	}
}

func TestMCPToolCall_Echo(t *testing.T) {
	router := setupTestRouter(t, false)
	sessionID := initializeSession(t, router)

	callReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "echo",
			"arguments": map[string]any{"message": "hello world"},
		},
	}
	w := doMCPPost(router, callReq, sessionID)

	if w.Code != http.StatusOK {
		t.Errorf("tools/call status = %d, want 200", w.Code)
		t.Logf("body: %s", w.Body.String())
		return
	}

	var resp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v\nbody: %s", err, w.Body.String())
	}

	if resp.Result.IsError {
		t.Error("expected no error from echo tool")
	}
	if len(resp.Result.Content) == 0 {
		t.Fatal("expected content in response")
	}
	if resp.Result.Content[0].Text != "Echo: hello world" {
		t.Errorf("echo result = %q, want %q", resp.Result.Content[0].Text, "Echo: hello world")
	}
}

func TestAuthRequired(t *testing.T) {
	router := setupTestRouter(t, true)

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}

	w := doMCPPost(router, initReq, "")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthWithValidToken(t *testing.T) {
	router := setupTestRouter(t, true)

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	}
	body, _ := json.Marshal(initReq)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 200 or 202", w.Code)
		t.Logf("body: %s", w.Body.String())
	}
}

func TestRegistry(t *testing.T) {
	registry := mcpserver.NewRegistry()

	tool := tools.NewEchoTool()
	if err := registry.Register(tool); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	if err := registry.Register(tool); err == nil {
		t.Error("expected error for duplicate registration")
	}

	got, ok := registry.Get("echo")
	if !ok {
		t.Fatal("Get() returned false for registered tool")
	}
	if got.Name != "echo" {
		t.Errorf("tool name = %q, want %q", got.Name, "echo")
	}

	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for unregistered tool")
	}

	list := registry.List()
	if len(list) != 1 {
		t.Errorf("List() returned %d tools, want 1", len(list))
	}
}

func TestMCPServerCreation(t *testing.T) {
	logger := slog.Default()
	srv := mcpserver.NewMCPServer("test", "1.0", logger)
	if srv == nil {
		t.Fatal("NewMCPServer() returned nil")
	}
	if srv.Server() == nil {
		t.Fatal("Server() returned nil")
	}
}

func TestContextPropagation(t *testing.T) {
	ctx := context.Background()
	user := middleware.UserFromContext(ctx)
	if user != nil {
		t.Error("expected nil user from empty context")
	}
}

func TestHealthzEnhanced(t *testing.T) {
	router := setupTestRouter(t, false)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("status = %v, want %q", resp["status"], "ok")
	}
	if resp["version"] == nil {
		t.Error("expected version in healthz response")
	}
	if resp["transport"] == nil {
		t.Error("expected transport in healthz response")
	}
	if resp["uptime_s"] == nil {
		t.Error("expected uptime_s in healthz response")
	}
}

func TestRequestIDInResponse(t *testing.T) {
	router := setupTestRouter(t, false)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	rid := w.Header().Get("X-Request-Id")
	if rid == "" {
		t.Error("expected X-Request-Id header in response")
	}
}

func TestMCPToolCall_Divide_Success(t *testing.T) {
	router := setupTestRouter(t, false)
	sessionID := initializeSession(t, router)

	callReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "divide",
			"arguments": map[string]any{"a": 10, "b": 3},
		},
	}
	w := doMCPPost(router, callReq, sessionID)

	if w.Code != http.StatusOK {
		t.Errorf("tools/call status = %d, want 200", w.Code)
		t.Logf("body: %s", w.Body.String())
		return
	}

	var resp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Result.IsError {
		t.Error("expected no error from divide tool")
	}
	if len(resp.Result.Content) == 0 {
		t.Fatal("expected content in response")
	}
	if resp.Result.Content[0].Text != "3.3333333333333335" {
		t.Errorf("divide result = %q, want %q", resp.Result.Content[0].Text, "3.3333333333333335")
	}
}

func TestMCPToolCall_Divide_ByZero(t *testing.T) {
	router := setupTestRouter(t, false)
	sessionID := initializeSession(t, router)

	callReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      11,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "divide",
			"arguments": map[string]any{"a": 10, "b": 0},
		},
	}
	w := doMCPPost(router, callReq, sessionID)

	if w.Code != http.StatusOK {
		t.Errorf("tools/call status = %d, want 200", w.Code)
		return
	}

	var resp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Result.IsError {
		t.Error("expected IsError=true for division by zero")
	}
	if len(resp.Result.Content) == 0 {
		t.Fatal("expected error content in response")
	}
	if resp.Result.Content[0].Text != "division by zero" {
		t.Errorf("error text = %q, want %q", resp.Result.Content[0].Text, "division by zero")
	}
}

func TestMCPToolsList_IncludesDivide(t *testing.T) {
	router := setupTestRouter(t, false)
	sessionID := initializeSession(t, router)

	listReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      20,
		"method":  "tools/list",
	}
	w := doMCPPost(router, listReq, sessionID)

	if w.Code != http.StatusOK {
		t.Errorf("tools/list status = %d, want 200", w.Code)
		return
	}

	var resp struct {
		Result struct {
			Tools []mcp.Tool `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	toolNames := make(map[string]bool)
	for _, tool := range resp.Result.Tools {
		toolNames[tool.Name] = true
	}

	if !toolNames["divide"] {
		t.Error("missing 'divide' tool in tools/list")
	}
	if len(resp.Result.Tools) < 3 {
		t.Errorf("expected at least 3 tools, got %d", len(resp.Result.Tools))
	}
}
