package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/hellobchain/mcp-go-sdk/mcpserver"
)

func NewDivideTool() mcpserver.ToolDefinition {
	return mcpserver.ToolDefinition{
		Name:        "divide",
		Description: "Divides two numbers. Demonstrates error handling with IsError.",
		Options: []mcp.ToolOption{
			mcp.WithDescription("Divides two numbers. Demonstrates error handling with IsError."),
			mcp.WithNumber("a", mcp.Required(), mcp.Description("Dividend")),
			mcp.WithNumber("b", mcp.Required(), mcp.Description("Divisor")),
		},
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()

			a, err := toFloat64(args["a"])
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid parameter 'a': %v", err)), nil
			}

			b, err := toFloat64(args["b"])
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid parameter 'b': %v", err)), nil
			}

			if b == 0 {
				return mcp.NewToolResultError("division by zero"), nil
			}

			result := a / b
			return mcp.NewToolResultText(fmt.Sprintf("%g", result)), nil
		},
	}
}

func toFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}
