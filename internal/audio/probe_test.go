package audio

import (
	"os"
	"path/filepath"
	"testing"

	"bookmux/internal/ffmpeg"
	"bookmux/internal/model"
)

func TestProbeTracks(t *testing.T) {
	dir := t.TempDir()
	ffprobePath := filepath.Join(dir, "ffprobe")
	if err := os.WriteFile(ffprobePath, []byte(`#!/bin/sh
cat <<'EOF'
{"format":{"duration":"2.500","bit_rate":"96000","tags":{"album":"Album"}},"streams":[{"codec_type":"audio","tags":{"title":"Chapter Title","artist":"Author"}}]}
EOF
`), 0o700); err != nil {
		t.Fatalf("failed to create ffprobe stub: %v", err)
	}

	oldFFprobePath := ffmpeg.FFprobePath
	defer func() {
		ffmpeg.FFprobePath = oldFFprobePath
	}()
	ffmpeg.FFprobePath = ffprobePath

	tracks := []model.InputTrack{
		{Path: "track1.mp3", BaseName: "track1.mp3"},
		{Path: "track2.mp3", BaseName: "track2.mp3"},
	}

	meta, err := ProbeTracks(tracks, false)
	if err != nil {
		t.Fatalf("ProbeTracks returned error: %v", err)
	}

	for i, track := range tracks {
		if track.DurationMs != 2500 {
			t.Fatalf("track[%d].DurationMs = %d, want 2500", i, track.DurationMs)
		}
		if track.Bitrate != "96k" {
			t.Fatalf("track[%d].Bitrate = %q, want %q", i, track.Bitrate, "96k")
		}
		if track.Chapter != "Chapter Title" {
			t.Fatalf("track[%d].Chapter = %q, want %q", i, track.Chapter, "Chapter Title")
		}
	}

	if meta.Author != "Author" || meta.Album != "Album" {
		t.Fatalf("returned metadata = %+v", meta)
	}
}

func TestProbeTracksReturnsProbeError(t *testing.T) {
	dir := t.TempDir()
	ffprobePath := filepath.Join(dir, "ffprobe")
	if err := os.WriteFile(ffprobePath, []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
		t.Fatalf("failed to create ffprobe stub: %v", err)
	}

	oldFFprobePath := ffmpeg.FFprobePath
	defer func() {
		ffmpeg.FFprobePath = oldFFprobePath
	}()
	ffmpeg.FFprobePath = ffprobePath

	tracks := []model.InputTrack{{Path: "track1.mp3", BaseName: "track1.mp3"}}
	if _, err := ProbeTracks(tracks, false); err == nil {
		t.Fatal("ProbeTracks returned nil error")
	}
}
