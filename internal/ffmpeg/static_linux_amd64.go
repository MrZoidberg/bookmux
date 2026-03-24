//go:build linux && amd64

package ffmpeg

import ff "github.com/go-ffstatic/linux-amd64"

func getStaticFFmpegPath() string {
	return ff.FFmpegPath()
}

func getStaticFFprobePath() string {
	return ff.FFprobePath()
}
