package ffmpeg

import (
	"testing"
)

func TestParseDurationToMs(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"00:00:01.00", 1000},
		{"00:00:01.50", 1500},
		{"00:01:00.00", 60000},
		{"01:00:00.00", 3600000},
		{"00:00:32.000", 32000},
		{"00:00:32.5", 32500},
	}

	for _, tc := range tests {
		got := ParseDurationToMs(tc.input)
		if got != tc.expected {
			t.Errorf("ParseDurationToMs(%q) = %d; want %d", tc.input, got, tc.expected)
		}
	}
}
