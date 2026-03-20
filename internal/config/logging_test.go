package config

import (
	"testing"

	"github.com/pterm/pterm"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected pterm.LogLevel
	}{
		{"trace", pterm.LogLevelTrace},
		{"debug", pterm.LogLevelDebug},
		{"info", pterm.LogLevelInfo},
		{"warn", pterm.LogLevelWarn},
		{"warning", pterm.LogLevelWarn},
		{"error", pterm.LogLevelError},
		{"fatal", pterm.LogLevelFatal},
		{"unknown", pterm.LogLevelInfo},
		{"", pterm.LogLevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := logLevel(tt.input)
			if result != tt.expected {
				t.Errorf("logLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
