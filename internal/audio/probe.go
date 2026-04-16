package audio

import (
	"bookmux/internal/ffmpeg"
	"bookmux/internal/model"
	"fmt"
	"log"
	"sync"
)

// ProbeTracks populates the DurationMs field for each track using ffprobe.
// It uses a pool of workers to speed up the process.
func ProbeTracks(tracks []model.InputTrack, verbose bool) (model.BookMetadata, error) {
	var wg sync.WaitGroup
	trackChan := make(chan int, len(tracks))
	errChan := make(chan error, len(tracks))

	// Use a worker pool (e.g., 8 workers)
	workerCount := min(8, len(tracks))

	for i := range workerCount {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for idx := range trackChan {
				if verbose {
					log.Printf("[Worker %d] Probing track %d/%d: %s", workerID, idx+1, len(tracks), tracks[idx].Path)
				}
				info, err := ffmpeg.GetAudioInfo(tracks[idx].Path)
				if err != nil {
					errChan <- fmt.Errorf("failed to probe %s: %w", tracks[idx].Path, err)
					return
				}
				tracks[idx].DurationMs = info.DurationMs
				tracks[idx].Bitrate = info.Bitrate
				tracks[idx].Codec = info.Codec
				tracks[idx].SampleRate = info.SampleRate
				tracks[idx].ChannelMode = info.ChannelMode
				if info.Metadata.Title != "" {
					tracks[idx].Chapter = info.Metadata.Title
				}
				if verbose {
					log.Printf(
						"[Worker %d] Track %d/%d info: codec=%s bitrate=%s sample_rate=%s channels=%s duration=%s size=%s",
						workerID,
						idx+1,
						len(tracks),
						fallback(tracks[idx].Codec, "unknown"),
						fallback(tracks[idx].Bitrate, "unknown"),
						formatSampleRate(tracks[idx].SampleRate),
						fallback(tracks[idx].ChannelMode, "unknown"),
						FormatDuration(tracks[idx].DurationMs),
						FormatBytes(tracks[idx].Size),
					)
				}
			}
		}(i)
	}

	for i := range tracks {
		trackChan <- i
	}
	close(trackChan)

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return model.BookMetadata{}, <-errChan
	}

	// For convenience, return the metadata of the first track
	info, _ := ffmpeg.GetAudioInfo(tracks[0].Path)
	return info.Metadata, nil
}
