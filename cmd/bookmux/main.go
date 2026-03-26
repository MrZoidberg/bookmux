package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"bookmux/internal/audio"
	"bookmux/internal/cli"
	"bookmux/internal/ffmpeg"
	"bookmux/internal/input"
	"bookmux/internal/model"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jessevdk/go-flags"
)

var version = "dev"
var logFile *os.File

type modelTUI struct {
	config        *model.BuildConfig
	tracks        []model.InputTrack
	err           error
	status        string
	done          bool
	progress      progress.Model
	totalDuration int64
	startTime     time.Time
	elapsedTime   time.Duration
	originalSize  int64
	resultSize    int64
	width         int
}

type statusMsg string
type errMsg error
type doneMsg struct{}
type progressMsg int64

func initialModel(cfg *model.BuildConfig) modelTUI {
	return modelTUI{
		config:   cfg,
		status:    "Starting...",
		progress:  progress.New(progress.WithDefaultGradient()),
		startTime: time.Now(),
	}
}

func (m modelTUI) Init() tea.Cmd {
	return func() tea.Msg {
		if err := ffmpeg.CheckDependencies(); err != nil {
			return errMsg(err)
		}
		if err := input.Validate(m.config); err != nil {
			return errMsg(err)
		}
		return statusMsg("Discovering files...")
	}
}

func discoverCmd(cfg *model.BuildConfig) tea.Cmd {
	return func() tea.Msg {
		tracks, err := input.DiscoverFiles(cfg)
		if err != nil {
			return errMsg(err)
		}
		return tracks
	}
}

func sortCmd(tracks []model.InputTrack, verbose bool) tea.Cmd {
	return func() tea.Msg {
		input.NaturalSort(tracks)
		if _, err := audio.ProbeTracks(tracks, verbose); err != nil {
			return errMsg(err)
		}
		return statusMsg("Probing complete. Merging...")
	}
}

func concatCmd(cfg *model.BuildConfig, tracks []model.InputTrack, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		var logger io.Writer
		if logFile != nil {
			logger = logFile
		}
		err := audio.ConcatFiles(logger, cfg, tracks, func(current int64) {
			p.Send(progressMsg(current))
		})
		if err != nil {
			return errMsg(err)
		}
		return doneMsg{}
	}
}

func (m modelTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, tea.Quit
	case statusMsg:
		m.status = string(msg)
		if m.status == "Discovering files..." {
			return m, discoverCmd(m.config)
		}
		if m.status == "Probing complete. Merging..." {
			// Calculate total duration for progress bar
			var total int64
			for _, t := range m.tracks {
				total += t.DurationMs
			}
			m.totalDuration = total
			return m, concatCmd(m.config, m.tracks, p)
		}
	case []model.InputTrack:
		m.tracks = msg
		var totalSize int64
		for _, t := range m.tracks {
			totalSize += t.Size
		}
		m.originalSize = totalSize
		m.status = fmt.Sprintf("Found %d files. Probing durations...", len(m.tracks))
		return m, sortCmd(m.tracks, m.config.Verbose)
	case progressMsg:
		if m.totalDuration > 0 {
			pct := float64(msg) / float64(m.totalDuration)
			if pct > 1.0 {
				pct = 1.0
			}
			return m, m.progress.SetPercent(pct)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.progress.Width = msg.Width - 10
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		return m, nil
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	case doneMsg:
		m.status = "Done! Successfully generated m4b."
		m.done = true
		m.elapsedTime = time.Since(m.startTime)
		if fi, err := os.Stat(m.config.OutputPath); err == nil {
			m.resultSize = fi.Size()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m modelTUI) View() string {
	if m.err != nil {
		width := m.width
		if width == 0 {
			width = 80
		}
		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")).
			Bold(true).
			Width(width - 4).
			Padding(0, 2)

		return fmt.Sprintf("\n%s\n\nPress q to quit.\n", errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}
	if m.done {
		s := fmt.Sprintf("\n%s\n\n", m.status)
		s += fmt.Sprintf("Original size:  %s\n", formatSize(m.originalSize))
		s += fmt.Sprintf("Resulting size: %s\n", formatSize(m.resultSize))
		s += fmt.Sprintf("Total duration: %s\n", m.elapsedTime.Round(time.Second))
		return s + "\n"
	}
	s := fmt.Sprintf("\n%s\n\nOverall Status: %s\n", renderHeader(), m.status)
	if m.totalDuration > 0 {
		s += "\n" + m.progress.View() + "\n"
	}
	s += "\nPress q to quit.\n"
	return s
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

var p *tea.Program

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := cli.ParseFlags(version)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return nil
		}
		return err
	}

	if cfg.Version {
		fmt.Printf("BookMux %s\n", version)
		return nil
	}

	if cfg.Shell != "" {
		if cfg.Install {
			return cli.InstallCompletion(cfg.Shell)
		}
		return cli.WriteCompletion(cfg.Shell)
	}

	if cfg.InputPath == "" || cfg.OutputPath == "" {
		ok, err := cli.RunInteractiveMode(cfg)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	if cfg.LogFile != "" {
		var err error
		logFile, err = os.OpenFile(cfg.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			return fmt.Errorf("error opening log file %s: %v", cfg.LogFile, err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.Println("--- BookMux Session Start ---")
	} else {
		log.SetOutput(io.Discard)
	}

	m := initialModel(cfg)
	p = tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("bubbletea error: %v", err)
	}
	return nil
}

func renderHeader() string {
	return fmt.Sprintf("BookMux %s", version)
}
