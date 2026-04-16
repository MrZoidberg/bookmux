package audio

import (
	"testing"

	"bookmux/internal/model"
)

func TestSummarizeTracks(t *testing.T) {
	tracks := []model.InputTrack{
		{
			DurationMs:  61_000,
			Size:        2 * 1024 * 1024,
			Bitrate:     "64k",
			Codec:       "mp3",
			SampleRate:  44100,
			ChannelMode: "mono",
		},
		{
			DurationMs:  119_000,
			Size:        3 * 1024 * 1024,
			Bitrate:     "64k",
			Codec:       "mp3",
			SampleRate:  44100,
			ChannelMode: "mono",
		},
		{
			DurationMs:  180_000,
			Size:        5 * 1024 * 1024,
			Bitrate:     "96k",
			Codec:       "aac",
			SampleRate:  48000,
			ChannelMode: "stereo",
		},
	}

	summary := SummarizeTracks(tracks)

	if summary.TrackCount != 3 {
		t.Fatalf("TrackCount = %d, want 3", summary.TrackCount)
	}
	if summary.TotalDurationMs != 360_000 {
		t.Fatalf("TotalDurationMs = %d, want 360000", summary.TotalDurationMs)
	}
	if summary.TotalSizeBytes != 10*1024*1024 {
		t.Fatalf("TotalSizeBytes = %d, want %d", summary.TotalSizeBytes, 10*1024*1024)
	}
	if summary.CodecCounts["mp3"] != 2 || summary.CodecCounts["aac"] != 1 {
		t.Fatalf("unexpected codec counts: %#v", summary.CodecCounts)
	}
	if summary.BitrateCounts["64k"] != 2 || summary.BitrateCounts["96k"] != 1 {
		t.Fatalf("unexpected bitrate counts: %#v", summary.BitrateCounts)
	}
	if summary.SampleRateCounts["44.1kHz"] != 2 || summary.SampleRateCounts["48kHz"] != 1 {
		t.Fatalf("unexpected sample rate counts: %#v", summary.SampleRateCounts)
	}
	if summary.ChannelCounts["mono"] != 2 || summary.ChannelCounts["stereo"] != 1 {
		t.Fatalf("unexpected channel counts: %#v", summary.ChannelCounts)
	}
}

func TestFormatSummaryCounts(t *testing.T) {
	got := FormatSummaryCounts(map[string]int{
		"stereo": 1,
		"mono":   2,
		"aac":    3,
	})

	want := "aac x3, mono x2, stereo x1"
	if got != want {
		t.Fatalf("FormatSummaryCounts() = %q, want %q", got, want)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{name: "zero", ms: 0, want: "0s"},
		{name: "seconds", ms: 61_000, want: "1m1s"},
		{name: "hours", ms: 3_661_000, want: "1h1m1s"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatDuration(tc.ms); got != tc.want {
				t.Fatalf("FormatDuration(%d) = %q, want %q", tc.ms, got, tc.want)
			}
		})
	}
}
