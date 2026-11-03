// Package tools provides example MCP tool implementations.
// Includes echo, time, and divide tools demonstrating various patterns
// such as error handling with IsError and context usage.
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/hellobchain/mcp-go-sdk/mcpserver"
)

func NewEchoTool() mcpserver.ToolDefinition {
	return mcpserver.ToolDefinition{
		Name:        "echo",
		Description: "Echoes back the input message",
		Options: []mcp.ToolOption{
			mcp.WithDescription("Echoes back the input message"),
			mcp.WithString("message", mcp.Required(), mcp.Description("The message to echo")),
		},
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			msg, _ := args["message"].(string)
			if msg == "" {
				msg = "(empty)"
			}
			return mcp.NewToolResultText(fmt.Sprintf("Echo: %s", msg)), nil
		},
	}
}
