package input

import (
	"testing"

	"bookmux/internal/model"
)

func TestNaturalSort(t *testing.T) {
	tracks := []model.InputTrack{
		{BaseName: "chapter10.mp3"},
		{BaseName: "chapter2.mp3"},
		{BaseName: "chapter1.mp3"},
	}

	NaturalSort(tracks)

	want := []string{"chapter1.mp3", "chapter2.mp3", "chapter10.mp3"}
	for i, expected := range want {
		if tracks[i].BaseName != expected {
			t.Fatalf("track[%d] = %q, want %q", i, tracks[i].BaseName, expected)
		}
	}
}
