package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestDivideTool_Success(t *testing.T) {
	tool := NewDivideTool()

	req := mcp.CallToolRequest{}
	req.Params.Name = "divide"
	req.Params.Arguments = map[string]any{"a": float64(10), "b": float64(4)}

	result, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if tc.Text != "2.5" {
		t.Errorf("result = %q, want %q", tc.Text, "2.5")
	}
}

func TestDivideTool_ByZero(t *testing.T) {
	tool := NewDivideTool()

	req := mcp.CallToolRequest{}
	req.Params.Name = "divide"
	req.Params.Arguments = map[string]any{"a": float64(10), "b": float64(0)}

	result, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for division by zero")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected error content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if tc.Text != "division by zero" {
		t.Errorf("error = %q, want %q", tc.Text, "division by zero")
	}
}

func TestDivideTool_InvalidInput(t *testing.T) {
	tool := NewDivideTool()

	req := mcp.CallToolRequest{}
	req.Params.Name = "divide"
	req.Params.Arguments = map[string]any{"a": "not-a-number", "b": float64(2)}

	result, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid input")
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    float64
		wantErr bool
	}{
		{"float64", float64(3.14), 3.14, false},
		{"float32", float32(2.5), 2.5, false},
		{"int", 42, 42.0, false},
		{"string", "1.5", 1.5, false},
		{"bad string", "abc", 0, true},
		{"nil", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toFloat64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("toFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}
