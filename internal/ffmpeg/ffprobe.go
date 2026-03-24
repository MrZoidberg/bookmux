package ffmpeg

import (
	"fmt"
	"strconv"

	"github.com/xfrr/goffmpeg/transcoder"
)

// GetAudioInfo returns the duration (ms) and bitrate of the audio file.
func GetAudioInfo(path string) (durationMs int64, bitrate string, err error) {
	t := new(transcoder.Transcoder)
	// Initialize with empty output as we only need probing
	if err := t.Initialize(path, ""); err != nil {
		return 0, "", fmt.Errorf("transcoder initialize failed: %w", err)
	}

	metadata := t.MediaFile().Metadata()
	durationSec, _ := strconv.ParseFloat(metadata.Format.Duration, 64)
	bitrate = metadata.Format.BitRate
	if bitrate != "" {
		if brInt, err := strconv.ParseInt(bitrate, 10, 64); err == nil {
			bitrate = fmt.Sprintf("%dk", brInt/1000)
		}
	}

	return int64(durationSec * 1000), bitrate, nil
}
