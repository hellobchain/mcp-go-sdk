// Package audit provides structured audit logging for security events.
// It captures important actions like authentication, tool calls, and errors
// for security monitoring and compliance.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

// EventType represents the type of audit event.
type EventType string

const (
	// EventAuth represents authentication events (login, logout, token validation)
	EventAuth EventType = "auth"
	// EventToolCall represents tool invocation events
	EventToolCall EventType = "tool_call"
	// EventError represents error events
	EventError EventType = "error"
	// EventConfigChange represents configuration change events
	EventConfigChange EventType = "config_change"
	// EventAccess represents access control events
	EventAccess EventType = "access"
)

// Event represents a single audit log entry.
type Event struct {
	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`
	// Type is the event type
	Type EventType `json:"type"`
	// Action is the specific action within the event type
	Action string `json:"action"`
	// UserID is the identifier of the user who triggered the event (if authenticated)
	UserID string `json:"user_id,omitempty"`
	// RequestID is the unique identifier for the request
	RequestID string `json:"request_id,omitempty"`
	// Resource is the resource being accessed (e.g., tool name, endpoint)
	Resource string `json:"resource,omitempty"`
	// Metadata contains additional event-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Success indicates whether the action was successful
	Success bool `json:"success"`
	// ErrorMessage contains error details if Success is false
	ErrorMessage string `json:"error_message,omitempty"`
	// ClientIP is the IP address of the client
	ClientIP string `json:"client_ip,omitempty"`
}

// Logger defines the interface for audit logging.
type Logger interface {
	// Log records an audit event
	Log(ctx context.Context, event Event)
}

// SlogLogger implements Logger using slog.
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new slog-based audit logger.
func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	return &SlogLogger{logger: logger}
}

// Log records an audit event using slog.
func (l *SlogLogger) Log(ctx context.Context, event Event) {
	// Convert event to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		l.logger.ErrorContext(ctx, "failed to marshal audit event", "error", err)
		return
	}

	attrs := []any{
		"audit_event", string(eventJSON),
		"event_type", string(event.Type),
		"action", event.Action,
		"success", event.Success,
	}

	if event.UserID != "" {
		attrs = append(attrs, "user_id", event.UserID)
	}
	if event.RequestID != "" {
		attrs = append(attrs, "request_id", event.RequestID)
	}
	if event.Resource != "" {
		attrs = append(attrs, "resource", event.Resource)
	}
	if event.ClientIP != "" {
		attrs = append(attrs, "client_ip", event.ClientIP)
	}
	if event.ErrorMessage != "" {
		attrs = append(attrs, "error_message", event.ErrorMessage)
	}

	if event.Success {
		l.logger.InfoContext(ctx, "audit", attrs...)
	} else {
		l.logger.WarnContext(ctx, "audit", attrs...)
	}
}

// NopLogger is a no-op audit logger that discards all events.
type NopLogger struct{}

// NewNopLogger creates a new no-op audit logger.
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// Log discards the audit event.
func (l *NopLogger) Log(ctx context.Context, event Event) {
	// No-op
}

// Global audit logger instance
var globalLogger Logger = NewNopLogger()

// SetLogger sets the global audit logger.
func SetLogger(logger Logger) {
	if logger != nil {
		globalLogger = logger
	}
}

// GetLogger returns the global audit logger.
func GetLogger() Logger {
	return globalLogger
}

// Log is a convenience function to log an audit event using the global logger.
func Log(ctx context.Context, event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	globalLogger.Log(ctx, event)
}

// LogAuth logs an authentication event.
func LogAuth(ctx context.Context, action, userID, requestID, clientIP string, success bool, errMsg string) {
	Log(ctx, Event{
		Type:         EventAuth,
		Action:       action,
		UserID:       userID,
		RequestID:    requestID,
		ClientIP:     clientIP,
		Success:      success,
		ErrorMessage: errMsg,
	})
}

// LogToolCall logs a tool call event.
func LogToolCall(ctx context.Context, toolName, userID, requestID string, success bool, duration time.Duration, errMsg string) {
	metadata := map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
	}
	Log(ctx, Event{
		Type:         EventToolCall,
		Action:       "call",
		Resource:     toolName,
		UserID:       userID,
		RequestID:    requestID,
		Success:      success,
		ErrorMessage: errMsg,
		Metadata:     metadata,
	})
}

// LogAccess logs an access control event.
func LogAccess(ctx context.Context, action, resource, userID, requestID, clientIP string, success bool, errMsg string) {
	Log(ctx, Event{
		Type:         EventAccess,
		Action:       action,
		Resource:     resource,
		UserID:       userID,
		RequestID:    requestID,
		ClientIP:     clientIP,
		Success:      success,
		ErrorMessage: errMsg,
	})
}
