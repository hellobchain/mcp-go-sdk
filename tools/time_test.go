package tools

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/hellobchain/mcp-go-sdk/middleware"
)

func TestTimeTool(t *testing.T) {
	tool := NewTimeTool()
	if tool.Name != "time" {
		t.Errorf("name = %q, want %q", tool.Name, "time")
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = "time"

	result, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}

	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	_, err = time.Parse(time.RFC3339, tc.Text)
	if err != nil {
		t.Errorf("result %q is not valid RFC3339: %v", tc.Text, err)
	}
}

func TestTimeTool_WithAuthUser(t *testing.T) {
	tool := NewTimeTool()

	req := mcp.CallToolRequest{}
	req.Params.Name = "time"

	user := &middleware.AuthUser{UserID: "alice", Roles: []string{"user"}}
	ctx := context.WithValue(context.Background(), contextKey("test"), user)

	result, err := tool.Handler(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
}

type contextKey string
