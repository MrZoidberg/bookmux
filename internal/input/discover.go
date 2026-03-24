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
// If recursive is true, it uses WalkDir, else it just reads the flat directory.
func DiscoverFiles(cfg *model.BuildConfig) ([]model.InputTrack, error) {
	// For MVP: input could be a directory or a comma-separated list of files
	// If it's a comma-separated list, bypass scanning.
	if strings.Contains(cfg.InputPath, ",") {
		var tracks []model.InputTrack
		files := strings.Split(cfg.InputPath, ",")
		for _, f := range files {
			f = strings.TrimSpace(f)
			if strings.ToLower(filepath.Ext(f)) == ".mp3" {
				tracks = append(tracks, model.InputTrack{
					Path:     f,
					BaseName: filepath.Base(f),
				})
			}
		}
		if len(tracks) == 0 {
			return nil, fmt.Errorf("no valid mp3 files found in input list")
		}
		return tracks, nil
	}

	info, err := os.Stat(cfg.InputPath)
	if err != nil {
		return nil, fmt.Errorf("invalid input path: %w", err)
	}

	if !info.IsDir() {
		// Single file?
		if strings.ToLower(filepath.Ext(cfg.InputPath)) != ".mp3" {
			return nil, fmt.Errorf("input file is not an mp3")
		}
		return []model.InputTrack{{
			Path:     cfg.InputPath,
			BaseName: filepath.Base(cfg.InputPath),
		}}, nil
	}

	var tracks []model.InputTrack
	if cfg.Recursive {
		err = filepath.WalkDir(cfg.InputPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".mp3" {
				tracks = append(tracks, model.InputTrack{
					Path:     path,
					BaseName: filepath.Base(path),
				})
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(cfg.InputPath)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".mp3" {
					path := filepath.Join(cfg.InputPath, e.Name())
					tracks = append(tracks, model.InputTrack{
						Path:     path,
						BaseName: filepath.Base(path),
					})
				}
			}
		}
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
