package cli

import (
	"bookmux/internal/ffmpeg"
	"bookmux/internal/input"
	"bookmux/internal/model"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// RunInteractiveMode runs a multi-step configuration wizard.
func RunInteractiveMode(cfg *model.BuildConfig) (bool, error) {
	if err := ffmpeg.CheckDependencies(); err != nil {
		return false, fmt.Errorf("failed to initialize dependencies for interactive mode: %w", err)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(1, 2).
		MarginBottom(1).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("205")).
		Render("  BookMux Interactive Mode  ")

	fmt.Println(header)

	var inputPath string
	var outputPath string

	// Step 1: Input Directory
	inputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Input Directory").
				Description("Where are your audio files located?").
				Placeholder("./my_audiobook").
				Value(&inputPath).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("input directory is required")
					}
					if _, err := os.Stat(s); os.IsNotExist(err) {
						return fmt.Errorf("directory does not exist")
					}
					return nil
				}),
		),
	)

	err := inputForm.Run()
	if err != nil {
		return false, err
	}
	cfg.InputPath = inputPath

	// Scanning feedback (simple)
	fmt.Println("Scanning files...")
	tracks, err := input.DiscoverFiles(cfg)
	if err != nil {
		return false, fmt.Errorf("scan error: %v", err)
	}

	if len(tracks) == 0 {
		return false, fmt.Errorf("no audio files found in %s", cfg.InputPath)
	}

	fmt.Printf("Found %d audio files.\n\n", len(tracks))

	// Pre-fill metadata from the first track
	fmt.Printf("Probing first track for metadata: %s\n", tracks[0].Path)
	info, probeErr := ffmpeg.GetAudioInfo(tracks[0].Path)
	if probeErr != nil {
		fmt.Printf("Warning: Failed to probe meta from first track: %v\n", probeErr)
	} else {
		sourceBitrate := info.Bitrate
		meta := info.Metadata
		if meta.Title != "" {
			fmt.Printf("Found title: %s\n", meta.Title)
			if cfg.Title == "" {
				cfg.Title = meta.Title
			}
		}
		if meta.Author != "" {
			fmt.Printf("Found author: %s\n", meta.Author)
			if cfg.Author == "" {
				cfg.Author = meta.Author
			}
		}
		if meta.Album != "" {
			fmt.Printf("Found album: %s\n", meta.Album)
			if cfg.Album == "" {
				cfg.Album = meta.Album
			}
		}

		if sourceBitrate != "" {
			fmt.Printf("Found source bitrate: %s\n", sourceBitrate)
			if cfg.Bitrate == "" {
				cfg.Bitrate = sourceBitrate
			}
		}

		if meta.HasCover && cfg.CoverPath == "" {
			fmt.Println("Found cover art in first track.")
			tempCover := filepath.Join(os.TempDir(), "bookmux_cover.jpg")
			if err := ffmpeg.ExtractCover(tracks[0].Path, tempCover); err == nil {
				cfg.CoverPath = tempCover
				fmt.Printf("Defaulting to source cover: %s\n", cfg.CoverPath)
			} else {
				fmt.Printf("Warning: Failed to extract cover: %v\n", err)
			}
		}
	}

	// Fallback to directory name for title if still empty
	if cfg.Title == "" {
		dirName := filepath.Base(cfg.InputPath)
		if dirName != "." && dirName != "" {
			cfg.Title = dirName
		}
	}

	// Step 2: Output Path
	outputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Output File").
				Description("Path to the resulting .m4b file").
				Placeholder("audiobook.m4b").
				Value(&outputPath).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("output path is required")
					}
					return nil
				}),
		),
	)

	err = outputForm.Run()
	if err != nil {
		return false, err
	}
	cfg.OutputPath = outputPath

	// Step 3: Settings Form
	settingsForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&cfg.Title),
			huh.NewInput().
				Title("Author").
				Value(&cfg.Author),
			huh.NewInput().
				Title("Album").
				Value(&cfg.Album),
			huh.NewInput().
				Title("Cover Path").
				Placeholder("None").
				Value(&cfg.CoverPath),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Bitrate").
				Options(
					huh.NewOption("44k", "44k"),
					huh.NewOption("64k", "64k"),
					huh.NewOption("96k", "96k"),
					huh.NewOption("128k", "128k"),
				).
				Value(&cfg.Bitrate),
			huh.NewSelect[string]().
				Title("Chapters").
				Options(
					huh.NewOption("From Files", "from-files"),
					huh.NewOption("None", "none"),
				).
				Value(&cfg.Chapters),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Loudness Normalization").
				Value(&cfg.Normalize),
			huh.NewConfirm().
				Title("Mono Conversion").
				Value(&cfg.Mono),
			huh.NewConfirm().
				Title("Recursive Scan").
				Value(&cfg.Recursive),
			huh.NewConfirm().
				Title("Overwrite Existing").
				Value(&cfg.Overwrite),
		),
	)

	err = settingsForm.Run()
	if err != nil {
		return false, err
	}

	return true, nil
}
