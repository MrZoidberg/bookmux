//go:build !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (windows && amd64))

package ffmpeg

func getStaticFFmpegPath() string {
	return ""
}

func getStaticFFprobePath() string {
	return ""
}
