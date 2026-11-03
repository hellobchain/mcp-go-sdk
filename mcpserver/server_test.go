package mcpserver

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewMCPServer(t *testing.T) {
	logger := slog.Default()
	srv := NewMCPServer("test", "1.0", logger)
	if srv == nil {
		t.Fatal("NewMCPServer() returned nil")
	}
	if srv.Server() == nil {
		t.Fatal("Server() returned nil")
	}
}

func TestAddTool(t *testing.T) {
	logger := slog.Default()
	srv := NewMCPServer("test", "1.0", logger)

	tool := ToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		Options: []mcp.ToolOption{
			mcp.WithDescription("A test tool"),
		},
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		},
	}

	srv.AddTool(tool)
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if len(r.List()) != 0 {
		t.Errorf("new registry should be empty, got %d tools", len(r.List()))
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()

	tool := ToolDefinition{
		Name:        "my_tool",
		Description: "test",
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		},
	}

	if err := r.Register(tool); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	got, ok := r.Get("my_tool")
	if !ok {
		t.Fatal("Get() returned false for registered tool")
	}
	if got.Name != "my_tool" {
		t.Errorf("tool name = %q, want %q", got.Name, "my_tool")
	}
}

func TestRegistryDuplicate(t *testing.T) {
	r := NewRegistry()

	tool := ToolDefinition{Name: "dup"}
	r.Register(tool)

	if err := r.Register(tool); err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for unregistered tool")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(ToolDefinition{Name: "a"})
	r.Register(ToolDefinition{Name: "b"})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("List() returned %d tools, want 2", len(list))
	}
}

func TestRegistryRegisterAll(t *testing.T) {
	r := NewRegistry()
	r.Register(ToolDefinition{
		Name: "tool1",
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		},
	})
	r.Register(ToolDefinition{
		Name: "tool2",
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		},
	})

	logger := slog.Default()
	srv := NewMCPServer("test", "1.0", logger)
	r.RegisterAll(srv)
}

func TestContextKeyMap(t *testing.T) {
	cm := &contextKeyMap{}
	ctx := context.Background()
	now := time.Now()

	cm.Set(ctx, now)
	got := cm.Get(ctx)
	if !got.Equal(now) {
		t.Errorf("Get() = %v, want %v", got, now)
	}
}

func TestContextKeyMap_Missing(t *testing.T) {
	cm := &contextKeyMap{}
	ctx := context.Background()
	before := time.Now()

	got := cm.Get(ctx)
	if got.Before(before) {
		t.Errorf("Get() for missing key returned time before test start")
	}
}
