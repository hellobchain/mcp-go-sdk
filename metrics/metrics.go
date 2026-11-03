// Package metrics provides Prometheus metrics for MCP server observability.
// It tracks HTTP requests and tool calls with configurable labels.
package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	toolCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_tool_calls_total",
			Help: "Total number of MCP tool calls",
		},
		[]string{"tool", "is_error"},
	)

	toolCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_tool_call_duration_seconds",
			Help:    "MCP tool call duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tool"},
	)
)

// RecordHTTPRequest records an HTTP request metric.
func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusStr := strconv.Itoa(status)
	httpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordToolCall records a tool call metric.
func RecordToolCall(tool string, isError bool, duration time.Duration) {
	toolCallsTotal.WithLabelValues(tool, strconv.FormatBool(isError)).Inc()
	toolCallDuration.WithLabelValues(tool).Observe(duration.Seconds())
}
