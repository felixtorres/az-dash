package utils

import (
	"testing"
	"time"
)

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		offset time.Duration
		want   string
	}{
		{10 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{1 * time.Minute, "1m ago"},
		{2 * time.Hour, "2h ago"},
		{1 * time.Hour, "1h ago"},
		{3 * 24 * time.Hour, "3d ago"},
		{1 * 24 * time.Hour, "1d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := RelativeTime(time.Now().Add(-tt.offset))
			if got != tt.want {
				t.Errorf("RelativeTime(-%v) = %q, want %q", tt.offset, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "-"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m30s"},
		{3661 * time.Second, "1h1m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatDuration(tt.d)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"hello", 3, "hel"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
