package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestEchoTool(t *testing.T) {
	tool := NewEchoTool()
	if tool.Name != "echo" {
		t.Errorf("name = %q, want %q", tool.Name, "echo")
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = "echo"
	req.Params.Arguments = map[string]any{"message": "hello"}

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
	if tc.Text != "Echo: hello" {
		t.Errorf("result = %q, want %q", tc.Text, "Echo: hello")
	}
}

func TestEchoTool_EmptyMessage(t *testing.T) {
	tool := NewEchoTool()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tc := result.Content[0].(mcp.TextContent)
	if tc.Text != "Echo: (empty)" {
		t.Errorf("result = %q, want %q", tc.Text, "Echo: (empty)")
	}
}
