package audio

import (
	"testing"

	"bookmux/internal/model"
)

func TestSelectTargetBitrate(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *model.BuildConfig
		tracks []model.InputTrack
		want   string
	}{
		{
			name: "prefers explicit config bitrate",
			cfg: &model.BuildConfig{
				Bitrate: "128k",
			},
			tracks: []model.InputTrack{
				{Bitrate: "64k"},
			},
			want: "128k",
		},
		{
			name: "falls back to first discovered track bitrate",
			cfg:  &model.BuildConfig{},
			tracks: []model.InputTrack{
				{},
				{Bitrate: "112k"},
				{Bitrate: "96k"},
			},
			want: "112k",
		},
		{
			name:   "uses default when no bitrate is available",
			cfg:    &model.BuildConfig{},
			tracks: []model.InputTrack{{}, {}},
			want:   "96k",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := selectTargetBitrate(tc.cfg, tc.tracks)
			if got != tc.want {
				t.Fatalf("selectTargetBitrate() = %q, want %q", got, tc.want)
			}
		})
	}
}
