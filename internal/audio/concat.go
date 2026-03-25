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

// ConcatFiles takes the sorted tracks and uses ffmpeg to merge them.
// It pre-processes tracks in parallel (transcoding/normalizing) and then merges them.
func ConcatFiles(_ io.Writer, cfg *model.BuildConfig, tracks []model.InputTrack, progressCallback func(int64)) error {
	if !cfg.Overwrite {
		if _, err := os.Stat(cfg.OutputPath); err == nil {
			return fmt.Errorf("output file %s already exists; use --overwrite to replace it", cfg.OutputPath)
		}
	}

	// 1. Probe all tracks to get durations (needed for progress)
	if cfg.Verbose {
		log.Printf("Probing tracks...")
	}
	if err := ProbeTracks(tracks, cfg.Verbose); err != nil {
		return err
	}

	var totalDurationMs int64
	for _, t := range tracks {
		totalDurationMs += t.DurationMs
	}

	tempDir := cfg.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	workDir, err := os.MkdirTemp(tempDir, "bookmux-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

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
			if err := trans.Initialize(track.Path, outPath); err != nil {
				results <- taskResult{idx: idx, err: err}
				return
			}

			// Configure transcoding
			trans.MediaFile().SetSkipVideo(true)
			trans.MediaFile().SetAudioCodec("aac")
			
			bitrate := cfg.Bitrate
			if bitrate == "" {
				bitrate = track.Bitrate
			}
			if bitrate == "" {
				bitrate = "96k"
			}
			trans.MediaFile().SetAudioBitRate(bitrate)

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

	raw = append(raw, "-c:a", "copy")
	
	trans.MediaFile().SetOutputPath(cfg.OutputPath)
	trans.MediaFile().SetRawOutputArgs(raw)

	if cfg.Verbose {
		log.Printf("Executing final merge via goffmpeg")
		log.Printf("FFmpeg command: ffmpeg %v", trans.GetCommand())
	}

	done := trans.Run(false)
	return <-done
}
