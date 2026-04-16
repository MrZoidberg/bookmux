package audio

import (
	"bookmux/internal/model"
	"fmt"
	"strings"
	"time"
)

type ProbeSummary struct {
	TrackCount       int
	TotalDurationMs  int64
	TotalSizeBytes   int64
	CodecCounts      map[string]int
	ChannelCounts    map[string]int
	SampleRateCounts map[string]int
	BitrateCounts    map[string]int
}

func SummarizeTracks(tracks []model.InputTrack) ProbeSummary {
	summary := ProbeSummary{
		TrackCount:       len(tracks),
		CodecCounts:      make(map[string]int),
		ChannelCounts:    make(map[string]int),
		SampleRateCounts: make(map[string]int),
		BitrateCounts:    make(map[string]int),
	}

	for _, track := range tracks {
		summary.TotalDurationMs += track.DurationMs
		summary.TotalSizeBytes += track.Size
		summary.CodecCounts[fallback(track.Codec, "unknown")]++
		summary.ChannelCounts[fallback(track.ChannelMode, "unknown")]++
		summary.SampleRateCounts[formatSampleRate(track.SampleRate)]++
		summary.BitrateCounts[fallback(track.Bitrate, "unknown")]++
	}

	return summary
}

func FormatDuration(ms int64) string {
	if ms <= 0 {
		return "0s"
	}

	return (time.Duration(ms) * time.Millisecond).Round(time.Second).String()
}

func FormatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func FormatSummaryCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return "n/a"
	}

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}

	// Small maps; simple insertion sort avoids another import.
	for i := 1; i < len(keys); i++ {
		j := i
		for j > 0 && keys[j-1] > keys[j] {
			keys[j-1], keys[j] = keys[j], keys[j-1]
			j--
		}
	}

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s x%d", key, counts[key]))
	}

	return strings.Join(parts, ", ")
}

func formatSampleRate(rate int) string {
	if rate <= 0 {
		return "unknown"
	}

	if rate%1000 == 0 {
		return fmt.Sprintf("%dkHz", rate/1000)
	}

	return fmt.Sprintf("%.1fkHz", float64(rate)/1000)
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}
