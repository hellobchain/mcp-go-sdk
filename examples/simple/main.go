// Example: Simple MCP Server using the SDK
//
// This example demonstrates how to create a simple MCP server using the SDK.
// Run with: go run examples/simple/main.go
package main

import (
	"context"
	"log"

	"github.com/hellobchain/mcp-go-sdk/config"
	"github.com/hellobchain/mcp-go-sdk/mcpserver"
	"github.com/hellobchain/mcp-go-sdk/sdk"
	"github.com/mark3labs/mcp-go/mcp"
)

// CustomToolProvider provides custom tools
type CustomToolProvider struct{}

// RegisterTools registers custom tools with the server
func (p *CustomToolProvider) RegisterTools(registry *mcpserver.Registry) {
	// Register a simple greeting tool
	registry.Register(mcpserver.ToolDefinition{
		Name:        "greet",
		Description: "Greets a user by name",
		Options: []mcp.ToolOption{
			mcp.WithDescription("Greets a user by name"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name to greet")),
		},
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			name, _ := args["name"].(string)
			return mcp.NewToolResultText("Hello, " + name + "!"), nil
		},
	})

	// Register a calculator tool
	registry.Register(mcpserver.ToolDefinition{
		Name:        "add",
		Description: "Adds two numbers",
		Options: []mcp.ToolOption{
			mcp.WithDescription("Adds two numbers"),
			mcp.WithNumber("a", mcp.Required(), mcp.Description("First number")),
			mcp.WithNumber("b", mcp.Required(), mcp.Description("Second number")),
		},
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			a, _ := args["a"].(float64)
			b, _ := args["b"].(float64)
			result := a + b
			return mcp.NewToolResultText("Result: " + string(rune(result))), nil
		},
	})
}

func main() {
	// Option 1: Quick start with defaults
	// sdk.QuickStart()

	// Option 2: Create server with custom configuration
	cfg := &config.Config{
		Transport:     config.TransportHTTP,
		Addr:          ":8080",
		ServerName:    "my-mcp-server",
		ServerVersion: "1.0.0",
		LogLevel:      config.LogLevelInfo,
	}

	server, err := sdk.NewServer(
		sdk.WithConfig(cfg),
		sdk.WithToolProvider(&CustomToolProvider{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the server
	ctx := context.Background()
	if err := server.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	// Run the server
	log.Printf("Starting MCP server on %s", cfg.Addr)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
