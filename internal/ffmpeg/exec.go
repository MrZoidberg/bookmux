package ffmpeg

import (
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"unsafe"

	"github.com/xfrr/goffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
)

// Progress represents the progress information from ffmpeg.
// We keep this for compatibility with the TUI.
type Progress struct {
	Frames      int64
	FPS         float64
	Bitrate     string
	TotalSize   int64
	OutTimeMs   int64
	Speed       string
	ProgressStr string
}

// Run executes a command and captures its output.
// Note: goffmpeg is usually used via its Transcoder API, but we keep this for simple utility calls.
func Run(logger io.Writer, name string, args ...string) error {
	path := name
	if name == "ffmpeg" {
		path = FFmpegPath
	} else if name == "ffprobe" {
		path = FFprobePath
	}

	// #nosec G204
	cmd := exec.Command(path, args...)
	if logger != nil {
		cmd.Stderr = logger
		cmd.Stdout = logger
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	return nil
}

// RunWithProgress is being phased out in favor of direct goffmpeg.Transcoder usage
// in the audio package.
func RunWithProgress(_ io.Writer, _ func(Progress), _ string, _ ...string) error {
	// For now, keep the old implementation or a wrapper if possible.
	// But since goffmpeg is a builder, it's better to refactor the callers.
	// I'll leave this here as a placeholder during refactor.
	return fmt.Errorf("RunWithProgress is deprecated, use goffmpeg.Transcoder directly")
}

// GetTranscoder returns a new goffmpeg transcoder with appropriate binary paths.
func GetTranscoder() *transcoder.Transcoder {
	t := new(transcoder.Transcoder)
	cfg := goffmpeg.Configuration{}

	// Use reflect to set unexported fields of goffmpeg.Configuration
	v := reflect.ValueOf(&cfg).Elem()

	fFfmpeg := v.FieldByName("ffmpegBinPath")
	if fFfmpeg.IsValid() && fFfmpeg.CanAddr() {
		ptr := unsafe.Pointer(fFfmpeg.UnsafeAddr())
		*(*string)(ptr) = FFmpegPath
	}

	fFfprobe := v.FieldByName("ffprobeBinPath")
	if fFfprobe.IsValid() && fFfprobe.CanAddr() {
		ptr := unsafe.Pointer(fFfprobe.UnsafeAddr())
		*(*string)(ptr) = FFprobePath
	}

	t.SetConfiguration(cfg)
	return t
}
