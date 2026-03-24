package ffmpeg

import (
	"strconv"
	"strings"
)

// ParseDurationToMs converts FFmpeg duration string (HH:MM:SS.ms) to milliseconds.
func ParseDurationToMs(s string) int64 {
	// Format: HH:MM:SS.ms
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.ParseFloat(parts[0], 64)
	m, _ := strconv.ParseFloat(parts[1], 64)
	sParts := strings.Split(parts[2], ".")
	sec, _ := strconv.ParseFloat(sParts[0], 64)
	var ms float64
	if len(sParts) > 1 {
		ms, _ = strconv.ParseFloat(sParts[1], 64)
		// goffmpeg might return different precision. 
		// If it's SS.ms (2 digits), then multiply by 10.
		if len(sParts[1]) == 2 {
			ms *= 10
		} else if len(sParts[1]) == 1 {
			ms *= 100
		}
	}
	return int64(h*3600000 + m*60000 + sec*1000 + ms)
}
