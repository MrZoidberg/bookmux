package ffmpeg

import (
	"fmt"
	"os/exec"
)

// CheckDependencies validates if required binaries are in the path.
func CheckDependencies() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}
	_, err = exec.LookPath("ffprobe")
	if err != nil {
		return fmt.Errorf("ffprobe not found in PATH: %w", err)
	}
	return nil
}
