//go:build darwin && amd64

package ffmpeg

import ff "github.com/go-ffstatic/darwin-amd64"

func getStaticFFmpegPath() string {
	return ff.FFmpegPath()
}

func getStaticFFprobePath() string {
	return ff.FFprobePath()
}
