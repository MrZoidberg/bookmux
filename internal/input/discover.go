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

// DiscoverFiles finds MP3 files in the given path.
func DiscoverFiles(cfg *model.BuildConfig) ([]model.InputTrack, error) {
	if strings.Contains(cfg.InputPath, ",") {
		return parseFileList(cfg.InputPath)
	}

	info, err := os.Stat(cfg.InputPath)
	if err != nil {
		return nil, fmt.Errorf("invalid input path: %w", err)
	}

	if !info.IsDir() {
		if strings.ToLower(filepath.Ext(cfg.InputPath)) != ".mp3" {
			return nil, fmt.Errorf("input file is not an mp3")
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
		return nil, fmt.Errorf("failed to scan for mp3s: %w", err)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no mp3 files found in %s", cfg.InputPath)
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
		if strings.ToLower(filepath.Ext(f)) == ".mp3" {
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
		return nil, fmt.Errorf("no valid mp3 files found in input list")
	}
	return tracks, nil
}

func scanRecursive(root string) ([]model.InputTrack, error) {
	var tracks []model.InputTrack
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".mp3" {
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
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".mp3" {
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
