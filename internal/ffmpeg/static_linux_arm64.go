//go:build linux && arm64

package ffmpeg

import ff "github.com/go-ffstatic/linux-arm64"

func getStaticFFmpegPath() string {
	return ff.FFmpegPath()
}

func getStaticFFprobePath() string {
	return ff.FFprobePath()
}
