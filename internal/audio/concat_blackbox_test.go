package audio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bookmux/internal/ffmpeg"
	"bookmux/internal/model"
)

func TestConcatFilesRejectsExistingOutputWithoutOverwrite(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "book.m4b")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o600); err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	err := ConcatFiles(nil, &model.BuildConfig{OutputPath: outputPath}, nil, nil)
	if err == nil {
		t.Fatal("ConcatFiles returned nil error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("ConcatFiles error = %q, want existing-output message", err.Error())
	}
}

func TestConcatFilesSuccess(t *testing.T) {
	dir := t.TempDir()

	ffmpegPath := filepath.Join(dir, "ffmpeg")
	if err := os.WriteFile(ffmpegPath, []byte(`#!/bin/sh
for last; do out="$last"; done
printf 'frame=1 fps=1 bitrate=96.0kbits/s time=00:00:01.00 speed=1x
' >&2
mkdir -p "$(dirname "$out")"
: > "$out"
`), 0o700); err != nil {
		t.Fatalf("failed to create ffmpeg stub: %v", err)
	}

	ffprobePath := filepath.Join(dir, "ffprobe")
	if err := os.WriteFile(ffprobePath, []byte(`#!/bin/sh
cat <<'EOF'
{"format":{"duration":"1.000","bit_rate":"96000","tags":{"album":"Album"}},"streams":[{"codec_type":"audio","tags":{"title":"Chapter","artist":"Author"}},{"codec_type":"video","tags":{}}]}
EOF
`), 0o700); err != nil {
		t.Fatalf("failed to create ffprobe stub: %v", err)
	}

	oldFFmpegPath := ffmpeg.FFmpegPath
	oldFFprobePath := ffmpeg.FFprobePath
	defer func() {
		ffmpeg.FFmpegPath = oldFFmpegPath
		ffmpeg.FFprobePath = oldFFprobePath
	}()
	ffmpeg.FFmpegPath = ffmpegPath
	ffmpeg.FFprobePath = ffprobePath

	input1 := filepath.Join(dir, "01.mp3")
	input2 := filepath.Join(dir, "02.mp3")
	for _, path := range []string{input1, input2} {
		if err := os.WriteFile(path, []byte("audio"), 0o600); err != nil {
			t.Fatalf("failed to create input %s: %v", path, err)
		}
	}

	outputPath := filepath.Join(dir, "output.m4b")
	cfg := &model.BuildConfig{
		OutputPath: outputPath,
		Title:      "Book",
		Author:     "Author",
		Album:      "Album",
		Mono:       true,
		Normalize:  true,
		Chapters:   "from-files",
		Overwrite:  true,
		TempDir:    dir,
	}
	tracks := []model.InputTrack{
		{Path: input1, BaseName: "01.mp3", Size: 5},
		{Path: input2, BaseName: "02.mp3", Size: 5},
	}

	var progressValues []int64
	if err := ConcatFiles(nil, cfg, tracks, func(v int64) {
		progressValues = append(progressValues, v)
	}); err != nil {
		t.Fatalf("ConcatFiles returned error: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file was not created: %v", err)
	}
	if len(progressValues) == 0 {
		t.Fatal("progress callback was not invoked")
	}
	if progressValues[len(progressValues)-1] != 4000 {
		t.Fatalf("final progress = %d, want 4000", progressValues[len(progressValues)-1])
	}
	if cfg.CoverPath == "" {
		t.Fatal("expected cover path to be auto-populated")
	}
}
