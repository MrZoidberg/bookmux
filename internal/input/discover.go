package input

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bookmux/internal/model"
)

func isSupportedAudioFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".m4a":
		return true
	default:
		return false
	}
}

// DiscoverFiles finds supported audio files in the given path.
func DiscoverFiles(cfg *model.BuildConfig) ([]model.InputTrack, error) {
	if strings.Contains(cfg.InputPath, ",") {
		return parseFileList(cfg.InputPath)
	}

	info, err := os.Stat(cfg.InputPath)
	if err != nil {
		return nil, fmt.Errorf("invalid input path: %w", err)
	}

	if !info.IsDir() {
		if !isSupportedAudioFile(cfg.InputPath) {
			return nil, fmt.Errorf("input file must be an .mp3 or .m4a")
		}
		return []model.InputTrack{{
			Path:     cfg.InputPath,
			BaseName: filepath.Base(cfg.InputPath),
			Size:     info.Size(),
		}}, nil
	}

	var tracks []model.InputTrack
	if cfg.Recursive {
		tracks, err = scanRecursive(cfg.InputPath)
	} else {
		tracks, err = scanFlat(cfg.InputPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan for audio files: %w", err)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no supported audio files found in %s", cfg.InputPath)
	}

	if cfg.Verbose {
		log.Printf("Discovered %d files in %s", len(tracks), cfg.InputPath)
	}

	return tracks, nil
}

func parseFileList(input string) ([]model.InputTrack, error) {
	var tracks []model.InputTrack
	for f := range strings.SplitSeq(input, ",") {
		f = strings.TrimSpace(f)
		if isSupportedAudioFile(f) {
			if info, err := os.Stat(f); err == nil {
				tracks = append(tracks, model.InputTrack{
					Path:     f,
					BaseName: filepath.Base(f),
					Size:     info.Size(),
				})
			}
		}
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no valid audio files found in input list")
	}
	return tracks, nil
}

func scanRecursive(root string) ([]model.InputTrack, error) {
	var tracks []model.InputTrack
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && isSupportedAudioFile(path) {
			if info, err := d.Info(); err == nil {
				tracks = append(tracks, model.InputTrack{
					Path:     path,
					BaseName: filepath.Base(path),
					Size:     info.Size(),
				})
			}
		}
		return nil
	})
	return tracks, err
}

func scanFlat(root string) ([]model.InputTrack, error) {
	var tracks []model.InputTrack
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() && isSupportedAudioFile(e.Name()) {
			if info, err := e.Info(); err == nil {
				path := filepath.Join(root, e.Name())
				tracks = append(tracks, model.InputTrack{
					Path:     path,
					BaseName: filepath.Base(path),
					Size:     info.Size(),
				})
			}
		}
	}
	return tracks, nil
}
