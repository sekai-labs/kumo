package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		notCont  string
	}{
		{
			name:     "Bearer token redaction",
			input:    "Request Authorization: Bearer c3RyaW5nX3Rva2VuX3ZhbHVlMTIzNA",
			contains: "Bearer [REDACTED]",
			notCont:  "c3RyaW5nX3Rva2VuX3ZhbHVlMTIzNA",
		},
		{
			name:     "Token key in json/query",
			input:    `{"token": "abcdefghijklmnopqrstuvwxyz123456"}`,
			contains: `{"token": "[REDACTED]"}`,
			notCont:  "abcdefghijklmnopqrstuvwxyz123456",
		},
		{
			name:     "Authorization header assignment",
			input:    "header authorization: secret_key_12345678",
			contains: "authorization: [REDACTED]",
			notCont:  "secret_key_12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := Sanitize(tt.input)
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected output to contain '%s', got '%s'", tt.contains, output)
			}
			if strings.Contains(output, tt.notCont) {
				t.Errorf("expected output NOT to contain '%s', got '%s'", tt.notCont, output)
			}
		})
	}
}

func TestLogger_Levels(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewWriterLogger(buf, LevelInfo)

	logger.Debug("debug message %d", 1)
	logger.Info("info message %d", 2)
	logger.Warn("warn message %d", 3)
	logger.Error("error message %d", 4)

	output := buf.String()
	if strings.Contains(output, "debug message 1") {
		t.Errorf("expected debug message to be filtered out")
	}
	if !strings.Contains(output, "info message 2") {
		t.Errorf("expected info message 2 to be logged")
	}
	if !strings.Contains(output, "warn message 3") {
		t.Errorf("expected warn message 3 to be logged")
	}
	if !strings.Contains(output, "error message 4") {
		t.Errorf("expected error message 4 to be logged")
	}
}
