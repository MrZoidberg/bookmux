package ffmpeg

import (
	"bookmux/internal/model"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GetAudioInfo returns the duration (ms), bitrate, and available metadata of the audio file.
func GetAudioInfo(path string) (durationMs int64, bitrate string, meta model.BookMetadata, err error) {
	// Bypass goffmpeg which strips tags during unmarshaling. We execute ffprobe directly.
	cmd := exec.Command(FFprobePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, "", model.BookMetadata{}, fmt.Errorf("ffprobe failed: %w", err)
	}

	var parsed struct {
		Format struct {
			Duration string            `json:"duration"`
			BitRate  string            `json:"bit_rate"`
			Tags     map[string]string `json:"tags"`
		} `json:"format"`
		Streams []struct {
			CodecType string            `json:"codec_type"`
			Tags      map[string]string `json:"tags"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(out, &parsed); err != nil {
		return 0, "", model.BookMetadata{}, fmt.Errorf("failed to parse ffprobe JSON: %w", err)
	}

	durationSec, _ := strconv.ParseFloat(parsed.Format.Duration, 64)
	bitrate = parsed.Format.BitRate
	if bitrate != "" {
		if brInt, err := strconv.ParseInt(bitrate, 10, 64); err == nil {
			bitrate = fmt.Sprintf("%dk", brInt/1000)
		}
	}

	// Normalize keys to lowercase for robust lookup
	tags := make(map[string]string)

	// First load from global format tags
	for k, v := range parsed.Format.Tags {
		tags[strings.ToLower(k)] = v
	}

	// Then append/fallback to tags from the first audio stream if present
	for _, stream := range parsed.Streams {
		if stream.CodecType == "audio" {
			for k, v := range stream.Tags {
				lk := strings.ToLower(k)
				if _, exists := tags[lk]; !exists {
					tags[lk] = v
				}
			}
			break
		}
	}

	meta = model.BookMetadata{
		Title:  tags["title"],
		Author: tags["artist"],
		Album:  tags["album"],
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

	return int64(durationSec * 1000), bitrate, meta, nil
}
