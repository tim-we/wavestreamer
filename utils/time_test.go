package utils

import (
	"testing"
	"time"
)

func TestPrettyDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{10 * time.Second, "just now"},
		{59 * time.Second, "just now"},
		{1 * time.Minute, "1 minute"},
		{5 * time.Minute, "5 minutes"},
		{59 * time.Minute, "59 minutes"},
		{1 * time.Hour, "1 hour"},
		{5 * time.Hour, "5 hours"},
		{23 * time.Hour, "23 hours"},
		{25 * time.Hour, "yesterday"},
		{48 * time.Hour, "2 days"},
		{7 * 24 * time.Hour, "7 days"},
	}

	for _, tt := range tests {
		got := PrettyDuration(tt.duration, "")
		if got != tt.expected {
			t.Errorf("PrettyDuration(%v) = %q, want %q", tt.duration, got, tt.expected)
		}
	}
}

func TestPrettyDurationWithSuffix(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{10 * time.Second, "just now"},
		{59 * time.Second, "just now"},
		{1 * time.Minute, "1 minute ago"},
		{5 * time.Minute, "5 minutes ago"},
		{59 * time.Minute, "59 minutes ago"},
		{1 * time.Hour, "1 hour ago"},
		{5 * time.Hour, "5 hours ago"},
		{23 * time.Hour, "23 hours ago"},
		{25 * time.Hour, "yesterday"},
		{48 * time.Hour, "2 days ago"},
		{7 * 24 * time.Hour, "7 days ago"},
	}

	for _, tt := range tests {
		got := PrettyDuration(tt.duration, " ago")
		if got != tt.expected {
			t.Errorf("PrettyDuration(%v) = %q, want %q", tt.duration, got, tt.expected)
		}
	}
}
