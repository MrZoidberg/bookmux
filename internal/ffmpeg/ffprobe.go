package ffmpeg

import (
	"bookmux/internal/model"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type AudioInfo struct {
	DurationMs  int64
	Bitrate     string
	Codec       string
	SampleRate  int
	ChannelMode string
	Metadata    model.BookMetadata
}

// GetAudioInfo returns the duration (ms), bitrate, and available metadata of the audio file.
func GetAudioInfo(path string) (AudioInfo, error) {
	// Bypass goffmpeg which strips tags during unmarshaling. We execute ffprobe directly.
	cmd := exec.Command(FFprobePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path)
	out, err := cmd.Output()
	if err != nil {
		return AudioInfo{}, fmt.Errorf("ffprobe failed: %w", err)
	}

	var parsed struct {
		Format struct {
			Duration string            `json:"duration"`
			BitRate  string            `json:"bit_rate"`
			Tags     map[string]string `json:"tags"`
		} `json:"format"`
		Streams []struct {
			CodecType     string            `json:"codec_type"`
			CodecName     string            `json:"codec_name"`
			Channels      int               `json:"channels"`
			ChannelLayout string            `json:"channel_layout"`
			SampleRate    string            `json:"sample_rate"`
			BitRate       string            `json:"bit_rate"`
			Tags          map[string]string `json:"tags"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(out, &parsed); err != nil {
		return AudioInfo{}, fmt.Errorf("failed to parse ffprobe JSON: %w", err)
	}

	durationSec, _ := strconv.ParseFloat(parsed.Format.Duration, 64)
	bitrate := formatBitrate(parsed.Format.BitRate)
	meta := model.BookMetadata{}

	// Normalize keys to lowercase for robust lookup
	tags := make(map[string]string)

	// First load from global format tags
	for k, v := range parsed.Format.Tags {
		tags[strings.ToLower(k)] = v
	}

	var (
		audioCodec   string
		audioBitrate string
		sampleRate   int
		channelMode  string
		hasCover     bool
	)
	// Then append/fallback to tags from the first audio stream if present
	for _, stream := range parsed.Streams {
		if stream.CodecType == "audio" {
			if audioCodec == "" {
				audioCodec = stream.CodecName
			}
			if audioBitrate == "" {
				audioBitrate = formatBitrate(stream.BitRate)
			}
			if sampleRate == 0 && stream.SampleRate != "" {
				if parsedRate, err := strconv.Atoi(stream.SampleRate); err == nil {
					sampleRate = parsedRate
				}
			}
			if channelMode == "" {
				channelMode = describeChannels(stream.Channels, stream.ChannelLayout)
			}
			for k, v := range stream.Tags {
				lk := strings.ToLower(k)
				if _, exists := tags[lk]; !exists {
					tags[lk] = v
				}
			}
		}
		if stream.CodecType == "video" {
			hasCover = true
		}
	}

	meta = model.BookMetadata{
		Title:    tags["title"],
		Author:   tags["artist"],
		Album:    tags["album"],
		HasCover: hasCover,
	}

	// Popular alternatives
	if meta.Author == "" {
		meta.Author = tags["album_artist"]
	}
	if meta.Author == "" {
		meta.Author = tags["authors"] // sometimes used in ID3
	}
	if meta.Author == "" {
		meta.Author = tags["composer"]
	}

	if bitrate == "" {
		bitrate = audioBitrate
	}

	return AudioInfo{
		DurationMs:  int64(durationSec * 1000),
		Bitrate:     bitrate,
		Codec:       audioCodec,
		SampleRate:  sampleRate,
		ChannelMode: channelMode,
		Metadata:    meta,
	}, nil
}

// ExtractCover extracts the first video stream (usually cover art) to the specified path.
func ExtractCover(src, dst string) error {
	// ffmpeg -i src -map 0:v:0 -c copy dst
	cmd := exec.Command(FFmpegPath, "-i", src, "-map", "0:v:0", "-c", "copy", "-y", dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg cover extraction failed: %w (output: %s)", err, string(out))
	}
	return nil
}

func formatBitrate(raw string) string {
	if raw == "" {
		return ""
	}

	brInt, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return raw
	}

	return fmt.Sprintf("%dk", brInt/1000)
}

func describeChannels(channels int, layout string) string {
	if layout != "" {
		return layout
	}

	switch channels {
	case 1:
		return "mono"
	case 2:
		return "stereo"
	case 0:
		return ""
	default:
		return fmt.Sprintf("%d channels", channels)
	}
}
