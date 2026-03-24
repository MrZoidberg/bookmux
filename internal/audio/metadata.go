package audio

import (
	"bookmux/internal/model"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateMetadata creates an FFMETADATA file with chapters and basic metadata.
func GenerateMetadata(tempDir string, cfg *model.BuildConfig, tracks []model.InputTrack) (string, error) {
	metaPath := filepath.Join(tempDir, "metadata.txt")
	f, err := os.Create(metaPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Write header
	fmt.Fprintln(f, ";FFMETADATA1")
	if cfg.Title != "" {
		fmt.Fprintf(f, "title=%s\n", cfg.Title)
	}
	if cfg.Author != "" {
		fmt.Fprintf(f, "artist=%s\n", cfg.Author)
		fmt.Fprintf(f, "album_artist=%s\n", cfg.Author)
	}
	if cfg.Album != "" {
		fmt.Fprintf(f, "album=%s\n", cfg.Album)
	}

	// Calculate and write chapters
	var currentMs int64 = 0
	for _, track := range tracks {
		fmt.Fprintln(f, "[CHAPTER]")
		fmt.Fprintln(f, "TIMEBASE=1/1000")
		fmt.Fprintf(f, "START=%d\n", currentMs)
		fmt.Fprintf(f, "END=%d\n", currentMs+track.DurationMs)

		// Clean up title: remove extension
		title := track.BaseName
		if ext := filepath.Ext(title); ext != "" {
			title = title[:len(title)-len(ext)]
		}
		fmt.Fprintf(f, "title=%s\n", title)

		currentMs += track.DurationMs
	}

	return metaPath, nil
}
