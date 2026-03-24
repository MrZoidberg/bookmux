package ffmpeg

import (
	"fmt"
	"os/exec"
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
	return nil
}
