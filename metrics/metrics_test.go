package metrics

import (
	"testing"
	"time"
)

func TestRecordHTTPRequest(t *testing.T) {
	RecordHTTPRequest("GET", "/test", 200, 100*time.Millisecond)
	RecordHTTPRequest("POST", "/test", 500, 50*time.Millisecond)
}

func TestRecordToolCall(t *testing.T) {
	RecordToolCall("test_tool", false, 100*time.Millisecond)
	RecordToolCall("test_tool", true, 50*time.Millisecond)
}
