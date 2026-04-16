package audio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bookmux/internal/model"
)

func TestGenerateMetadata(t *testing.T) {
	cfg := &model.BuildConfig{
		Title:  "Book Title",
		Author: "Book Author",
		Album:  "Book Album",
	}
	tracks := []model.InputTrack{
		{BaseName: "01-intro.mp3", DurationMs: 1000},
		{BaseName: "02-chapter.mp3", Chapter: "Custom Chapter", DurationMs: 2500},
	}

	path, err := GenerateMetadata(t.TempDir(), cfg, tracks)
	if err != nil {
		t.Fatalf("GenerateMetadata returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read metadata: %v", err)
	}

	got := string(content)
	for _, want := range []string{
		";FFMETADATA1",
		"title=Book Title",
		"artist=Book Author",
		"album_artist=Book Author",
		"album=Book Album",
		"START=0",
		"END=1000",
		"title=01-intro",
		"START=1000",
		"END=3500",
		"title=Custom Chapter",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("metadata missing %q in %q", want, got)
		}
	}
}
