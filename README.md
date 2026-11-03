# MCP Go SDK

一个简单易用的 MCP (Model Context Protocol) Go SDK，让你几行代码即可启动生产级 MCP 服务。

## 特性

- **快速启动**: 最少代码即可运行 MCP 服务器
- **多传输模式**: 支持 stdio、HTTP、multi 三种模式
- **生产级特性**: 认证、限流、超时、CORS、审计日志、链路追踪、指标监控
- **灵活配置**: YAML 配置文件、环境变量、命令行参数
- **易于扩展**: 简单的 Tool 注册接口

## 快速开始

### 安装

```bash
go get github.com/hellobchain/mcp-go-sdk/sdk
```

### 最简示例

```go
package main

import "github.com/hellobchain/mcp-go-sdk/sdk"

func main() {
    // 一行代码启动 MCP 服务
    sdk.QuickStart()
}
```

### 自定义配置

```go
package main

import (
    "log"
    "github.com/hellobchain/mcp-go-sdk/sdk"
    "github.com/hellobchain/mcp-go-sdk/config"
)

func main() {
    cfg := &config.Config{
        Transport:     config.TransportHTTP,
        Addr:          ":8080",
        ServerName:    "my-mcp-server",
        ServerVersion: "1.0.0",
        LogLevel:      config.LogLevelInfo,
    }

    server, err := sdk.NewServer(sdk.WithConfig(cfg))
    if err != nil {
        log.Fatal(err)
    }

    if err := server.Initialize(context.Background()); err != nil {
        log.Fatal(err)
    }

    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### 注册自定义 Tool

```go
package main

import (
    "context"
    "log"
    "github.com/hellobchain/mcp-go-sdk/sdk"
    "github.com/hellobchain/mcp-go-sdk/mcpserver"
    "github.com/mark3labs/mcp-go/mcp"
)

type MyToolProvider struct{}

func (p *MyToolProvider) RegisterTools(registry *mcpserver.Registry) {
    registry.Register(mcpserver.ToolDefinition{
        Name:        "greet",
        Description: "Greets a user",
        Options: []mcp.ToolOption{
            mcp.WithDescription("Greets a user by name"),
            mcp.WithString("name", mcp.Required(), mcp.Description("Name to greet")),
        },
        Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            name := req.GetArguments()["name"].(string)
            return mcp.NewToolResultText("Hello, " + name + "!"), nil
        },
    })
}

func main() {
    server, err := sdk.NewServer(
        sdk.WithToolProvider(&MyToolProvider{}),
    )
    if err != nil {
        log.Fatal(err)
    }

    server.Initialize(context.Background())
    server.Run()
}
```

## 配置

### 配置文件 (config.yaml)

```yaml
# 传输模式: stdio, http, multi
transport: http
addr: ":8080"

# 认证
auth_enabled: true
auth_tokens:
  "token1": "alice"
  "token2": "bob"

# 超时
request_timeout: 30s
tool_timeout: 60s
shutdown_timeout: 30s

# 速率限制
rate_limit_enabled: true
requests_per_second: 100
burst_size: 200

# CORS
cors_enabled: true
cors_origins: ["http://localhost:3000"]

# 可观测性
tracing_enabled: true
tracing_endpoint: "localhost:4318"
metrics_enabled: true
audit_log_enabled: true
pprof_enabled: true

# 日志
log_level: info
```

### 使用配置文件

```go
// 方式 1: 命令行参数
./your-app -config config.yaml

// 方式 2: 环境变量
MCP_CONFIG_FILE=config.yaml ./your-app

// 方式 3: 代码加载
cfg, _ := config.LoadFromFile("config.yaml")
server, _ := sdk.NewServer(sdk.WithConfig(cfg))
```

### 配置优先级

1. 环境变量 (最高)
2. 命令行参数
3. 配置文件
4. 默认值 (最低)

## 传输模式

### stdio 模式

适用于本地桌面客户端（Claude Desktop、Cursor）：

```yaml
transport: stdio
```

```bash
# 运行
./your-app -transport stdio
```

### HTTP 模式

适用于远程客户端接入：

```yaml
transport: http
addr: ":8080"
```

```bash
./your-app -transport http -addr :8080
```

### Multi 模式

同时运行 stdio + HTTP：

```yaml
transport: multi
addr: ":8080"
```

```bash
./your-app -transport multi -addr :8080
```

## API 端点

| 路径 | 方法 | 说明 |
|---|---|---|
| `/mcp` | POST | MCP JSON-RPC 消息入口 |
| `/mcp` | GET | SSE 流（如支持） |
| `/mcp` | DELETE | 会话终止 |
| `/healthz` | GET | 健康检查 |
| `/status` | GET | 运行时状态 |
| `/metrics` | GET | Prometheus 指标 |
| `/debug/pprof/*` | GET | pprof 性能分析 |

## 测试

```bash
# 使用 MCP Inspector 测试
npx @modelcontextprotocol/inspector http://localhost:8080/mcp

# 或 stdio 模式
npx @modelcontextprotocol/inspector ./your-app -transport stdio
```

## 生产部署

### Docker

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/mcp-server /mcp-server
COPY config.yaml /config.yaml
EXPOSE 8080
CMD ["/mcp-server", "-config", "/config.yaml"]
```

### 推荐配置

```yaml
# 生产环境配置
transport: http
addr: ":8080"

# 启用认证
auth_enabled: true
auth_tokens:
  "secure-token": "user"

# 配置超时
request_timeout: 30s
tool_timeout: 60s
shutdown_timeout: 30s

# 启用限流
rate_limit_enabled: true
requests_per_second: 100
burst_size: 200
per_user_limit: 50

# 启用可观测性
tracing_enabled: true
tracing_endpoint: "otel-collector:4318"
metrics_enabled: true
audit_log_enabled: true

# 日志级别
log_level: info
```

## 示例

查看 `examples/` 目录获取完整示例：

- [simple](examples/simple/main.go) - 最简示例
- 更多示例持续添加中...

## 架构

```
┌─────────────────────────────────────┐
│  SDK Layer (sdk.go)                 │
│  - NewServer()                      │
│  - QuickStart()                     │
│  - WithConfig() / WithToolProvider()│
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  App Layer (app/)                   │
│  - Server lifecycle                 │
│  - Middleware orchestration         │
│  - Route setup                      │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Domain Layers                      │
│  - config/   - middleware/          │
│  - mcpserver/ - transport/          │
│  - tools/    - metrics/             │
│  - audit/    - tracing/             │
└─────────────────────────────────────┘
```

## 许可证

MIT License

## 支持

- GitHub Issues: 报告 bug 或提出功能请求
- Discussions: 提问和分享使用经验
