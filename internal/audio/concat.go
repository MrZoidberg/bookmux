package audio

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"bookmux/internal/ffmpeg"
	"bookmux/internal/model"
	"bookmux/internal/util"
)

func selectTargetBitrate(cfg *model.BuildConfig, tracks []model.InputTrack) string {
	if cfg.Bitrate != "" {
		return cfg.Bitrate
	}

	for _, track := range tracks {
		if track.Bitrate != "" {
			return track.Bitrate
		}
	}

	return "96k"
}

func tracksAlreadyProbed(tracks []model.InputTrack) bool {
	if len(tracks) == 0 {
		return false
	}

	for _, track := range tracks {
		if track.DurationMs <= 0 {
			return false
		}
	}

	return true
}

func totalInputDuration(tracks []model.InputTrack) int64 {
	var total int64
	for _, track := range tracks {
		total += track.DurationMs
	}
	return total
}

// ConcatFiles takes the sorted tracks and uses ffmpeg to merge them.
// It pre-processes tracks in parallel (transcoding/normalizing) and then merges them.
func ConcatFiles(_ io.Writer, cfg *model.BuildConfig, tracks []model.InputTrack, progressCallback func(int64)) error {
	if !cfg.Overwrite {
		if _, err := os.Stat(cfg.OutputPath); err == nil {
			return fmt.Errorf("output file %s already exists; use --overwrite to replace it", cfg.OutputPath)
		}
	}

	// 1. Probe tracks if durations are not already available.
	var meta model.BookMetadata
	if tracksAlreadyProbed(tracks) {
		if cfg.Verbose {
			log.Printf("Using existing probe data for %d tracks", len(tracks))
		}
		info, err := ffmpeg.GetAudioInfo(tracks[0].Path)
		if err != nil {
			return err
		}
		meta = info.Metadata
	} else {
		if cfg.Verbose {
			log.Printf("Probing tracks...")
		}
		probedMeta, err := ProbeTracks(tracks, cfg.Verbose)
		if err != nil {
			return err
		}
		meta = probedMeta
	}

	summary := SummarizeTracks(tracks)
	targetBitrate := selectTargetBitrate(cfg, tracks)
	totalDurationMs := totalInputDuration(tracks)

	tempDir := cfg.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	workDir, err := os.MkdirTemp(tempDir, "bookmux-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	if cfg.Verbose {
		log.Printf("Input summary: files=%d total_duration=%s total_size=%s codecs=[%s] source_bitrates=[%s] sample_rates=[%s] channels=[%s]",
			summary.TrackCount,
			FormatDuration(summary.TotalDurationMs),
			FormatBytes(summary.TotalSizeBytes),
			FormatSummaryCounts(summary.CodecCounts),
			FormatSummaryCounts(summary.BitrateCounts),
			FormatSummaryCounts(summary.SampleRateCounts),
			FormatSummaryCounts(summary.ChannelCounts),
		)
		log.Printf("Build settings: output=%s temp_dir=%s target_bitrate=%s normalize=%t mono=%t chapters=%s overwrite=%t",
			cfg.OutputPath,
			workDir,
			targetBitrate,
			cfg.Normalize,
			cfg.Mono,
			cfg.Chapters,
			cfg.Overwrite,
		)
		log.Printf("Source metadata: title=%q author=%q album=%q embedded_cover=%t explicit_cover=%t",
			meta.Title,
			meta.Author,
			meta.Album,
			meta.HasCover,
			cfg.CoverPath != "",
		)
	}

	// Auto-extract cover if not provided and first track has one
	if cfg.CoverPath == "" && meta.HasCover {
		if cfg.Verbose {
			log.Printf("Found cover art in source, extracting...")
		}
		autoCover := filepath.Join(workDir, "auto_cover.jpg")
		if err := ffmpeg.ExtractCover(tracks[0].Path, autoCover); err == nil {
			cfg.CoverPath = autoCover
		} else if cfg.Verbose {
			log.Printf("Warning: Failed to auto-extract cover: %v", err)
		}
	}

	// 2. Stage 1: Pre-process tracks in parallel
	processedTracks := make([]model.InputTrack, len(tracks))
	copy(processedTracks, tracks)

	type taskResult struct {
		idx int
		err error
	}
	results := make(chan taskResult, len(tracks))
	semaphore := make(chan struct{}, 8) // Limit concurrency to 8

	// Track progress for each track
	trackProgress := make([]int64, len(tracks))
	var progressMu sync.Mutex

	updateGlobalProgress := func() {
		progressMu.Lock()
		var currentProgress int64
		for _, p := range trackProgress {
			currentProgress += p
		}
		progressMu.Unlock()
		if progressCallback != nil {
			progressCallback(currentProgress)
		}
	}

	for i := range tracks {
		go func(idx int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			track := tracks[idx]
			ext := ".m4a" // Standardize on m4a for concat
			outPath := filepath.Join(workDir, fmt.Sprintf("track_%05d%s", idx, ext))

			trans := ffmpeg.GetTranscoder()
			if err := trans.InitializeEmptyTranscoder(); err != nil {
				results <- taskResult{idx: idx, err: err}
				return
			}
			trans.MediaFile().SetInputPath(track.Path)
			trans.MediaFile().SetOutputPath(outPath)

			// Configure transcoding
			trans.MediaFile().SetSkipVideo(true)
			trans.MediaFile().SetAudioCodec("aac")
			trans.MediaFile().SetAudioBitRate(targetBitrate)
			trans.MediaFile().SetAudioRate(44100)

			if cfg.Mono {
				trans.MediaFile().SetAudioChannels(1)
			}

			if cfg.Normalize {
				trans.MediaFile().SetAudioFilter("loudnorm")
			}

			done := trans.Run(true)
			progress := trans.Output()

			go func() {
				for p := range progress {
					// Parse goffmpeg's CurrentTime (HH:MM:SS.ms) to milliseconds
					ms := ffmpeg.ParseDurationToMs(p.CurrentTime)
					progressMu.Lock()
					trackProgress[idx] = ms
					progressMu.Unlock()
					updateGlobalProgress()
				}
			}()

			err := <-done
			if err == nil {
				processedTracks[idx].Path = outPath
				// Mark as 100% done for this track
				progressMu.Lock()
				trackProgress[idx] = track.DurationMs
				progressMu.Unlock()
				updateGlobalProgress()
			}
			results <- taskResult{idx: idx, err: err}
		}(i)
	}

	for range tracks {
		res := <-results
		if res.err != nil {
			return fmt.Errorf("failed to process track %d: %w", res.idx, res.err)
		}
	}

	// 3. Handle Chapters (optional, usually from-files)
	var metadataPath string
	if cfg.Chapters == "from-files" {
		var err error
		metadataPath, err = GenerateMetadata(workDir, cfg, tracks)
		if err != nil {
			return fmt.Errorf("failed to generate metadata: %w", err)
		}
	}

	// 4. Stage 2: Merge processed tracks
	manifestPath := filepath.Join(workDir, "concat.txt")
	var lines []string
	for _, t := range processedTracks {
		absPath, _ := filepath.Abs(t.Path)
		safePath := strings.ReplaceAll(absPath, "'", "'\\''")
		lines = append(lines, fmt.Sprintf("file '%s'", safePath))
	}

	if err := util.WriteLines(manifestPath, lines); err != nil {
		return fmt.Errorf("failed to write concat manifest: %w", err)
	}

	trans := ffmpeg.GetTranscoder()
	// Use the manifest as the first input
	if err := trans.InitializeEmptyTranscoder(); err != nil {
		return err
	}

	// Configure inputs and global flags
	trans.MediaFile().SetInputPath(manifestPath)

	inputIdx := 0
	var inputArgs []string
	var metadataIdx, coverIdx, concatIdx int

	if metadataPath != "" {
		inputArgs = append(inputArgs, "-i", metadataPath)
		metadataIdx = inputIdx
		inputIdx++
	} else {
		metadataIdx = -1
	}

	if cfg.CoverPath != "" {
		inputArgs = append(inputArgs, "-i", cfg.CoverPath)
		coverIdx = inputIdx
		inputIdx++
	} else {
		coverIdx = -1
	}

	inputArgs = append(inputArgs, "-f", "concat", "-safe", "0")
	concatIdx = inputIdx

	trans.MediaFile().SetRawInputArgs(inputArgs)

	// Use raw parameters for complex mapping and metadata
	var raw []string
	if metadataPath != "" {
		raw = append(raw, "-map_metadata", fmt.Sprintf("%d", metadataIdx))
	}
	if cfg.CoverPath != "" {
		raw = append(raw, "-map", fmt.Sprintf("%d:a", concatIdx), "-map", fmt.Sprintf("%d:v", coverIdx), "-c:v", "copy", "-disposition:v:0", "attached_pic")
	} else {
		raw = append(raw, "-map", fmt.Sprintf("%d:a", concatIdx), "-vn")
	}

	if cfg.Title != "" {
		raw = append(raw, "-metadata", fmt.Sprintf("title=%s", cfg.Title))
	}
	if cfg.Author != "" {
		raw = append(raw, "-metadata", fmt.Sprintf("artist=%s", cfg.Author), "-metadata", fmt.Sprintf("album_artist=%s", cfg.Author))
	}
	if cfg.Album != "" {
		raw = append(raw, "-metadata", fmt.Sprintf("album=%s", cfg.Album))
	}

	raw = append(raw, "-c:a", "aac", "-b:a", targetBitrate)

	trans.MediaFile().SetOutputPath(cfg.OutputPath)
	trans.MediaFile().SetRawOutputArgs(raw)

	if cfg.Verbose {
		log.Printf("Executing final merge via goffmpeg")
		log.Printf("FFmpeg command: ffmpeg %v", trans.GetCommand())
	}

	done := trans.Run(true)
	progress := trans.Output()

	go func() {
		for p := range progress {
			if progressCallback == nil {
				continue
			}

			mergeProgressMs := ffmpeg.ParseDurationToMs(p.CurrentTime)
			if mergeProgressMs > totalDurationMs {
				mergeProgressMs = totalDurationMs
			}
			progressCallback(totalDurationMs + mergeProgressMs)
		}
	}()

	err = <-done
	if err == nil && progressCallback != nil {
		progressCallback(totalDurationMs * 2)
	}
	return err
}
