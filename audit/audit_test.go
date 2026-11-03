package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

func TestSlogLogger_Log(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	auditLogger := NewSlogLogger(logger)

	event := Event{
		Timestamp: time.Now(),
		Type:      EventAuth,
		Action:    "login",
		UserID:    "user123",
		RequestID: "req456",
		Success:   true,
		ClientIP:  "192.168.1.1",
	}

	auditLogger.Log(context.Background(), event)

	output := buf.String()
	if output == "" {
		t.Error("expected log output, got empty string")
	}

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("failed to parse log output as JSON: %v", err)
	}

	// Check key fields
	if logEntry["event_type"] != string(EventAuth) {
		t.Errorf("event_type = %v, want %v", logEntry["event_type"], EventAuth)
	}
	if logEntry["action"] != "login" {
		t.Errorf("action = %v, want %v", logEntry["action"], "login")
	}
	if logEntry["success"] != true {
		t.Errorf("success = %v, want true", logEntry["success"])
	}
}

func TestNopLogger_Log(t *testing.T) {
	logger := NewNopLogger()

	event := Event{
		Type:   EventAuth,
		Action: "login",
	}

	// Should not panic
	logger.Log(context.Background(), event)
}

func TestSetLogger(t *testing.T) {
	// Save original logger
	original := GetLogger()
	defer SetLogger(original)

	// Set new logger
	newLogger := NewNopLogger()
	SetLogger(newLogger)

	// Verify it was set
	if GetLogger() != newLogger {
		t.Error("SetLogger did not set the global logger")
	}

	// Setting nil should not change the logger
	SetLogger(nil)
	if GetLogger() != newLogger {
		t.Error("SetLogger(nil) should not change the global logger")
	}
}

func TestLogAuth(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	auditLogger := NewSlogLogger(logger)

	// Save original logger and restore after test
	original := GetLogger()
	defer SetLogger(original)

	SetLogger(auditLogger)

	LogAuth(context.Background(), "login", "user123", "req456", "192.168.1.1", true, "")

	output := buf.String()
	if output == "" {
		t.Error("expected log output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("failed to parse log output as JSON: %v", err)
	}

	if logEntry["user_id"] != "user123" {
		t.Errorf("user_id = %v, want user123", logEntry["user_id"])
	}
}

func TestLogToolCall(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	auditLogger := NewSlogLogger(logger)

	// Save original logger and restore after test
	original := GetLogger()
	defer SetLogger(original)

	SetLogger(auditLogger)

	LogToolCall(context.Background(), "echo", "user123", "req456", true, 100*time.Millisecond, "")

	output := buf.String()
	if output == "" {
		t.Error("expected log output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("failed to parse log output as JSON: %v", err)
	}

	if logEntry["resource"] != "echo" {
		t.Errorf("resource = %v, want echo", logEntry["resource"])
	}
}

func TestLogAccess(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	auditLogger := NewSlogLogger(logger)

	// Save original logger and restore after test
	original := GetLogger()
	defer SetLogger(original)

	SetLogger(auditLogger)

	LogAccess(context.Background(), "read", "/api/data", "user123", "req456", "192.168.1.1", false, "permission denied")

	output := buf.String()
	if output == "" {
		t.Error("expected log output, got empty string")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("failed to parse log output as JSON: %v", err)
	}

	if logEntry["success"] != false {
		t.Errorf("success = %v, want false", logEntry["success"])
	}
	if logEntry["error_message"] != "permission denied" {
		t.Errorf("error_message = %v, want permission denied", logEntry["error_message"])
	}
}

func TestEvent_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	auditLogger := NewSlogLogger(logger)

	// Save original logger and restore after test
	original := GetLogger()
	defer SetLogger(original)

	SetLogger(auditLogger)

	// Log without timestamp - should be auto-filled
	Log(context.Background(), Event{
		Type:   EventAuth,
		Action: "test",
	})

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("failed to parse log output as JSON: %v", err)
	}

	// Verify audit_event contains timestamp
	auditEventStr, ok := logEntry["audit_event"].(string)
	if !ok {
		t.Fatal("audit_event is not a string")
	}

	var auditEvent Event
	if err := json.Unmarshal([]byte(auditEventStr), &auditEvent); err != nil {
		t.Errorf("failed to parse audit_event: %v", err)
	}

	if auditEvent.Timestamp.IsZero() {
		t.Error("timestamp should be auto-filled if not provided")
	}
}
