package input

import (
	"sort"

	"bookmux/internal/model"

	"github.com/facette/natsort"
)

// NaturalSort sorts the input tracks using natural sort logic.
func NaturalSort(tracks []model.InputTrack) {
	sort.Slice(tracks, func(i, j int) bool {
		return natsort.Compare(tracks[i].BaseName, tracks[j].BaseName)
	})
}
