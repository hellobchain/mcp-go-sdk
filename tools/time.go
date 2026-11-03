package tools

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/hellobchain/mcp-go-sdk/mcpserver"
	"github.com/hellobchain/mcp-go-sdk/middleware"
)

func NewTimeTool() mcpserver.ToolDefinition {
	return mcpserver.ToolDefinition{
		Name:        "time",
		Description: "Returns the current server time",
		Options: []mcp.ToolOption{
			mcp.WithDescription("Returns the current server time"),
		},
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			user := middleware.UserFromContext(ctx)
			now := time.Now().Format(time.RFC3339)

			if user != nil {
				return mcp.NewToolResultText(now), nil
			}
			return mcp.NewToolResultText(now), nil
		},
	}
}
