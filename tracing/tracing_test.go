package tracing

import (
	"context"
	"log/slog"
	"testing"
)

func TestInit_Disabled(t *testing.T) {
	logger := slog.Default()
	p, err := Init(context.Background(), Config{
		Enabled:     false,
		ServiceName: "test",
		Version:     "1.0",
		Logger:      logger,
	})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if p == nil {
		t.Fatal("Init() returned nil provider")
	}
	if p.Tracer() == nil {
		t.Error("Tracer() returned nil")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestInit_Enabled_BadEndpoint(t *testing.T) {
	logger := slog.Default()
	p, err := Init(context.Background(), Config{
		Enabled:     true,
		ServiceName: "test",
		Version:     "1.0",
		Endpoint:    "localhost:99999",
		Insecure:    true,
		Logger:      logger,
	})
	if err != nil {
		t.Fatalf("Init() with bad endpoint should not fail at creation: %v", err)
	}
	if p == nil {
		t.Fatal("Init() returned nil provider")
	}
	if p.Tracer() == nil {
		t.Error("Tracer() returned nil")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Logf("Shutdown() error (expected with bad endpoint): %v", err)
	}
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()
	span := SpanFromContext(ctx)
	if span == nil {
		t.Error("SpanFromContext() returned nil")
	}
}

func TestShutdown_NilProvider(t *testing.T) {
	p := &Provider{}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown() with nil tp error = %v", err)
	}
}
