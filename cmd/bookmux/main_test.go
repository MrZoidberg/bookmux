package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bookmux/internal/ffmpeg"
	"bookmux/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func captureStream(t *testing.T, target **os.File, fn func()) string {
	t.Helper()

	old := *target
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	*target = w
	defer func() {
		*target = old
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}
	return buf.String()
}

func TestInitialModel(t *testing.T) {
	cfg := &model.BuildConfig{Title: "Example"}
	m := initialModel(cfg)

	if m.config != cfg {
		t.Fatal("initialModel did not keep config pointer")
	}
	if m.status != "Starting..." {
		t.Fatalf("status = %q, want %q", m.status, "Starting...")
	}
	if m.startTime.IsZero() {
		t.Fatal("startTime was not initialized")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{input: 42, want: "42 B"},
		{input: 1024, want: "1.00 KB"},
		{input: 5 * 1024 * 1024, want: "5.00 MB"},
	}

	for _, tc := range tests {
		if got := formatSize(tc.input); got != tc.want {
			t.Fatalf("formatSize(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestRenderHeader(t *testing.T) {
	oldVersion := version
	defer func() {
		version = oldVersion
	}()
	version = "1.2.3"

	if got := renderHeader(); got != "BookMux 1.2.3" {
		t.Fatalf("renderHeader() = %q", got)
	}
}

func TestViewStates(t *testing.T) {
	errView := modelTUI{err: os.ErrNotExist, width: 60}.View()
	if !strings.Contains(errView, "Error:") {
		t.Fatalf("error view = %q", errView)
	}

	doneView := modelTUI{
		status:       "Done! Successfully generated m4b.",
		done:         true,
		originalSize: 1024,
		resultSize:   2048,
		elapsedTime:  3 * time.Second,
	}.View()
	if !strings.Contains(doneView, "Original size:  1.00 KB") || !strings.Contains(doneView, "Resulting size: 2.00 KB") {
		t.Fatalf("done view = %q", doneView)
	}

	progressView := modelTUI{status: "Working...", totalDuration: 1000}.View()
	if !strings.Contains(progressView, "Overall Status: Working...") {
		t.Fatalf("progress view = %q", progressView)
	}
}

func TestUpdateTransitions(t *testing.T) {
	t.Run("key quit", func(t *testing.T) {
		m := initialModel(&model.BuildConfig{})
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatal("expected quit command")
		}
		if updated.(modelTUI).err != nil {
			t.Fatalf("unexpected error state: %+v", updated)
		}
	})

	t.Run("track discovery", func(t *testing.T) {
		m := initialModel(&model.BuildConfig{})
		tracks := []model.InputTrack{{Size: 10}, {Size: 20}}
		updated, cmd := m.Update(tracks)
		next := updated.(modelTUI)
		if next.originalSize != 30 {
			t.Fatalf("originalSize = %d, want 30", next.originalSize)
		}
		if !strings.Contains(next.status, "Found 2 files") {
			t.Fatalf("status = %q", next.status)
		}
		if cmd == nil {
			t.Fatal("expected follow-up sort command")
		}
	})

	t.Run("progress caps at one", func(t *testing.T) {
		m := initialModel(&model.BuildConfig{})
		m.totalDuration = 100
		updated, cmd := m.Update(progressMsg(150))
		if cmd == nil {
			t.Fatal("expected progress animation command")
		}
		if updated.(modelTUI).totalDuration != 100 {
			t.Fatalf("unexpected totalDuration change: %+v", updated)
		}
	})

	t.Run("done message", func(t *testing.T) {
		outputPath := filepath.Join(t.TempDir(), "out.m4b")
		if err := os.WriteFile(outputPath, []byte("done"), 0o600); err != nil {
			t.Fatalf("failed to create output: %v", err)
		}
		m := initialModel(&model.BuildConfig{OutputPath: outputPath})
		updated, cmd := m.Update(doneMsg{})
		next := updated.(modelTUI)
		if !next.done || next.resultSize == 0 {
			t.Fatalf("done state not populated: %+v", next)
		}
		if cmd == nil {
			t.Fatal("expected quit command")
		}
	})
}

func TestRunVersionAndCompletion(t *testing.T) {
	oldArgs := os.Args
	oldVersion := version
	defer func() {
		os.Args = oldArgs
		version = oldVersion
	}()

	version = "9.9.9"
	os.Args = []string{"bookmux", "--version"}
	stdout := captureStream(t, &os.Stdout, func() {
		if err := run(); err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	})
	if !strings.Contains(stdout, "BookMux 9.9.9") {
		t.Fatalf("version output = %q", stdout)
	}

	os.Args = []string{"bookmux", "--completion", "bash"}
	stdout = captureStream(t, &os.Stdout, func() {
		if err := run(); err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	})
	if !strings.Contains(stdout, "complete -F _bookmux bookmux") {
		t.Fatalf("completion output = %q", stdout)
	}
}

func TestInitAndCommands(t *testing.T) {
	t.Run("init validates config", func(t *testing.T) {
		cfg := &model.BuildConfig{InputPath: "/tmp/in", OutputPath: "/tmp/out.m4b"}
		msg := initialModel(cfg).Init()()
		if _, ok := msg.(statusMsg); !ok {
			t.Fatalf("Init returned %T, want statusMsg", msg)
		}
	})

	t.Run("init returns validation error", func(t *testing.T) {
		cfg := &model.BuildConfig{InputPath: "/tmp/in", OutputPath: "/tmp/out.mp3"}
		msg := initialModel(cfg).Init()()
		if _, ok := msg.(errMsg); !ok {
			t.Fatalf("Init returned %T, want errMsg", msg)
		}
	})

	t.Run("discover command", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := filepath.Join(dir, "chapter01.mp3")
		if err := os.WriteFile(inputPath, []byte("audio"), 0o600); err != nil {
			t.Fatalf("failed to create input: %v", err)
		}

		msg := discoverCmd(&model.BuildConfig{InputPath: dir})()
		tracks, ok := msg.([]model.InputTrack)
		if !ok {
			t.Fatalf("discoverCmd returned %T", msg)
		}
		if len(tracks) != 1 || tracks[0].BaseName != "chapter01.mp3" {
			t.Fatalf("discoverCmd returned %+v", tracks)
		}
	})

	t.Run("sort command", func(t *testing.T) {
		dir := t.TempDir()
		ffprobePath := filepath.Join(dir, "ffprobe")
		if err := os.WriteFile(ffprobePath, []byte(`#!/bin/sh
cat <<'EOF'
{"format":{"duration":"1.000","bit_rate":"96000","tags":{}},"streams":[{"codec_type":"audio","tags":{"title":"Sorted Chapter"}}]}
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
			{Path: "chapter10.mp3", BaseName: "chapter10.mp3"},
			{Path: "chapter2.mp3", BaseName: "chapter2.mp3"},
		}
		msg := sortCmd(tracks, false)()
		if got, ok := msg.(statusMsg); !ok || got != "Probing complete. Merging..." {
			t.Fatalf("sortCmd returned %T (%v)", msg, msg)
		}
		if tracks[0].BaseName != "chapter2.mp3" {
			t.Fatalf("tracks were not naturally sorted: %+v", tracks)
		}
		if tracks[0].DurationMs != 1000 || tracks[0].Chapter != "Sorted Chapter" {
			t.Fatalf("tracks were not probed: %+v", tracks)
		}
	})
}

func TestRunHeadlessValidationError(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	os.Args = []string{"bookmux", "--input", "/tmp/in", "--output", "/tmp/out.mp3"}
	err := run()
	if err == nil || !strings.Contains(err.Error(), "output file must end with .m4b") {
		t.Fatalf("run error = %v", err)
	}
}

func TestRunRequiresFlagsInCI(t *testing.T) {
	oldArgs := os.Args
	oldCI := os.Getenv("CI")
	defer func() {
		os.Args = oldArgs
		if err := os.Setenv("CI", oldCI); err != nil {
			t.Fatalf("failed to restore CI env: %v", err)
		}
	}()

	if err := os.Setenv("CI", "true"); err != nil {
		t.Fatalf("failed to set CI env: %v", err)
	}
	os.Args = []string{"bookmux"}

	err := run()
	if err == nil || !strings.Contains(err.Error(), "--input and --output flags are required in CI environment") {
		t.Fatalf("run error = %v", err)
	}
}
