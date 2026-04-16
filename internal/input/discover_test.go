package input

import (
	"os"
	"path/filepath"
	"testing"

	"bookmux/internal/model"
)

func TestDiscoverFilesSupportsSingleM4AInput(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "chapter01.m4a")
	if err := os.WriteFile(inputPath, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	tracks, err := DiscoverFiles(&model.BuildConfig{InputPath: inputPath})
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("DiscoverFiles returned %d tracks, want 1", len(tracks))
	}
	if tracks[0].BaseName != "chapter01.m4a" {
		t.Fatalf("DiscoverFiles returned basename %q, want %q", tracks[0].BaseName, "chapter01.m4a")
	}
}

func TestDiscoverFilesSupportsMixedAudioDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	files := []string{"chapter01.mp3", "chapter02.m4a", "notes.txt"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(name), 0o600); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	tracks, err := DiscoverFiles(&model.BuildConfig{InputPath: tmpDir})
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("DiscoverFiles returned %d tracks, want 2", len(tracks))
	}
}

func TestDiscoverFilesSupportsM4AInInputList(t *testing.T) {
	tmpDir := t.TempDir()
	mp3Path := filepath.Join(tmpDir, "chapter01.mp3")
	m4aPath := filepath.Join(tmpDir, "chapter02.m4a")
	for _, path := range []string{mp3Path, m4aPath} {
		if err := os.WriteFile(path, []byte(filepath.Base(path)), 0o600); err != nil {
			t.Fatalf("failed to create %s: %v", path, err)
		}
	}

	inputList := mp3Path + "," + m4aPath
	tracks, err := DiscoverFiles(&model.BuildConfig{InputPath: inputList})
	if err != nil {
		t.Fatalf("DiscoverFiles returned error: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("DiscoverFiles returned %d tracks, want 2", len(tracks))
	}
}
