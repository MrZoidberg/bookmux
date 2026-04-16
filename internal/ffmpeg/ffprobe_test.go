package ffmpeg

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetAudioInfo(t *testing.T) {
	dir := t.TempDir()
	ffprobe := fakeExecutable(t, dir, "ffprobe", `#!/bin/sh
cat <<'EOF'
{"format":{"duration":"12.345","bit_rate":"64000","tags":{"TITLE":"Book Title","ALBUM":"Book Album"}},"streams":[{"codec_type":"audio","tags":{"title":"Chapter 1","ALBUM_ARTIST":"Primary Author"}},{"codec_type":"video","tags":{}}]}
EOF
`)

	oldPath := FFprobePath
	defer func() {
		FFprobePath = oldPath
	}()
	FFprobePath = ffprobe

	durationMs, bitrate, meta, err := GetAudioInfo("input.m4a")
	if err != nil {
		t.Fatalf("GetAudioInfo returned error: %v", err)
	}

	if durationMs != 12345 {
		t.Fatalf("durationMs = %d, want 12345", durationMs)
	}
	if bitrate != "64k" {
		t.Fatalf("bitrate = %q, want %q", bitrate, "64k")
	}
	if meta.Title != "Book Title" || meta.Author != "Primary Author" || meta.Album != "Book Album" || !meta.HasCover {
		t.Fatalf("metadata = %+v", meta)
	}
}

func TestGetAudioInfoInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	ffprobe := fakeExecutable(t, dir, "ffprobe", "#!/bin/sh\nprintf 'not-json'\n")

	oldPath := FFprobePath
	defer func() {
		FFprobePath = oldPath
	}()
	FFprobePath = ffprobe

	if _, _, _, err := GetAudioInfo("input.m4a"); err == nil {
		t.Fatal("GetAudioInfo returned nil error for invalid JSON")
	}
}

func TestExtractCover(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "cover.jpg")

	t.Run("success", func(t *testing.T) {
		ffmpegPath := fakeExecutable(t, dir, "ffmpeg-success", "#!/bin/sh\nprintf 'cover' > \"$8\"\n")

		oldPath := FFmpegPath
		defer func() {
			FFmpegPath = oldPath
		}()
		FFmpegPath = ffmpegPath

		if err := ExtractCover("in.m4a", dst); err != nil {
			t.Fatalf("ExtractCover returned error: %v", err)
		}
		if _, err := os.Stat(dst); err != nil {
			t.Fatalf("cover file was not created: %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		ffmpegPath := fakeExecutable(t, dir, "ffmpeg-fail", "#!/bin/sh\nprintf 'boom' >&2\nexit 1\n")

		oldPath := FFmpegPath
		defer func() {
			FFmpegPath = oldPath
		}()
		FFmpegPath = ffmpegPath

		err := ExtractCover("in.m4a", dst)
		if err == nil {
			t.Fatal("ExtractCover returned nil error")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("ExtractCover error = %q, want stderr output", err.Error())
		}
	})
}

func TestRun(t *testing.T) {
	var log strings.Builder

	if err := Run(&log, "/bin/sh", "-c", "printf out; printf err >&2"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if got := log.String(); got != "outerr" {
		t.Fatalf("logger output = %q, want %q", got, "outerr")
	}

	if err := Run(nil, "/bin/sh", "-c", "exit 7"); err == nil {
		t.Fatal("Run returned nil error for failure")
	}
}

func TestRunWithProgressDeprecated(t *testing.T) {
	if err := RunWithProgress(nil, nil, "ffmpeg"); err == nil {
		t.Fatal("RunWithProgress returned nil error")
	}
}

func TestGetTranscoder(t *testing.T) {
	oldFFmpegPath := FFmpegPath
	oldFFprobePath := FFprobePath
	defer func() {
		FFmpegPath = oldFFmpegPath
		FFprobePath = oldFFprobePath
	}()

	FFmpegPath = "/tmp/ffmpeg"
	FFprobePath = "/tmp/ffprobe"

	if trans := GetTranscoder(); trans == nil {
		t.Fatal("GetTranscoder returned nil")
	}
}

func TestEnsureWindowsExe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ffmpeg")
	if err := os.WriteFile(path, []byte("bin"), 0o600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	got := ensureWindowsExe(path)
	if want := path + ".exe"; got != want {
		t.Fatalf("ensureWindowsExe returned %q, want %q", got, want)
	}
	if _, err := os.Stat(path + ".exe"); err != nil {
		t.Fatalf("renamed file missing: %v", err)
	}
}

func TestCheckDependencies(t *testing.T) {
	if err := CheckDependencies(); err != nil {
		t.Fatalf("CheckDependencies returned error: %v", err)
	}
	if FFmpegPath == "" || FFprobePath == "" {
		t.Fatalf("dependency paths not populated: ffmpeg=%q ffprobe=%q", FFmpegPath, FFprobePath)
	}
}

func fakeExecutable(t *testing.T, dir, name, script string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		path += ".bat"
		script = "@echo off\r\n" + script
	}
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatalf("failed to write fake executable: %v", err)
	}
	return path
}
