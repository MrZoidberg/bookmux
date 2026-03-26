package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	FFmpegPath  string
	FFprobePath string
)

// CheckDependencies validates if required binaries are in the path.
func CheckDependencies() error {
	// Try static first
	FFmpegPath = getStaticFFmpegPath()
	FFprobePath = getStaticFFprobePath()

	if FFmpegPath == "" {
		p, err := exec.LookPath("ffmpeg")
		if err != nil {
			return fmt.Errorf("ffmpeg not found in PATH: %w", err)
		}
		FFmpegPath = p
	}

	if FFprobePath == "" {
		p, err := exec.LookPath("ffprobe")
		if err != nil {
			return fmt.Errorf("ffprobe not found in PATH: %w", err)
		}
		FFprobePath = p
	}

	// On Windows, extracted static binaries MUST have .exe extension to be executable
	// If the path doesn't have an extension, try to rename it.
	if runtime.GOOS == "windows" {
		FFmpegPath = ensureWindowsExe(FFmpegPath)
		FFprobePath = ensureWindowsExe(FFprobePath)
	}

	return nil
}

func ensureWindowsExe(path string) string {
	if path == "" {
		return ""
	}
	if filepath.Ext(path) == "" {
		newPath := path + ".exe"
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			_ = os.Rename(path, newPath)
		}
		return newPath
	}
	return path
}
