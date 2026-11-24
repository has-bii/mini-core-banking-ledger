package middleware

import (
	"encoding/json/v2"
	"testing"
)

func capturePanic(fn func()) (value any) {
	defer func() { value = recover() }()
	fn()
	return nil
}

func decodeLog(t *testing.T, logs []byte) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(logs, &record); err != nil {
		t.Fatalf("decode single log record: %v; log: %s", err, logs)
	}
	return record
}
