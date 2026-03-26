//go:build e2e

package e2e_test

import (
	"archive/zip"
	"bookmux/internal/ffmpeg"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

type BookChapter struct {
	Section         int    `json:"section"`
	Chapter         string `json:"chapter"`
	Reader          string `json:"reader"`
	Duration        string `json:"duration"`
	DurationSeconds int    `json:"duration_seconds"`
}

type BookDef struct {
	Title    string        `json:"title"`
	Author   string        `json:"author"`
	Chapters []BookChapter `json:"chapters"`
}

type FFprobeOutput struct {
	Chapters []struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Tags      struct {
			Title string `json:"title"`
		} `json:"tags"`
	} `json:"chapters"`
}

func downloadFile(t *testing.T, url, dest string) {
	t.Helper()
	t.Logf("Downloading %s to %s", url, dest)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to download file %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Bad status for %s: %s", url, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}
}

func unzipFile(t *testing.T, src, dest string) {
	t.Helper()
	t.Logf("Unzipping %s to %s", src, dest)
	r, err := zip.OpenReader(src)
	if err != nil {
		t.Fatalf("Failed to open zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			t.Fatalf("Failed to open output file: %v", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			t.Fatalf("Failed to open zip file content: %v", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			t.Fatalf("Failed to copy zip file content: %v", err)
		}
	}
}

func fetchAudiobooks() ([]BookDef, error) {
	resp, err := http.Get("http://bookmux-demo.s3.eu-central-1.amazonaws.com/audiobooks.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch audiobooks.json: %w", err)
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to read json: %w", err)
	}

	var books []BookDef
	if strings.HasPrefix(strings.TrimSpace(string(raw)), "[") {
		err = json.Unmarshal(raw, &books)
	} else {
		var single BookDef
		err = json.Unmarshal(raw, &single)
		books = append(books, single)
	}
	return books, err
}

func TestE2E_BookMuxConversion(t *testing.T) {
	// First, fetch the audiobook specs
	books, err := fetchAudiobooks()
	if err != nil {
		t.Fatalf("Could not fetch test specs: %v", err)
	}

	// Prepare one binary for all subtests
	tmpDir := t.TempDir()
	exeName := "bookmux"
	if runtime.GOOS == "windows" {
		exeName = "bookmux.exe"
	}
	exePath := filepath.Join(tmpDir, exeName)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	repoRoot := cwd
	if filepath.Base(cwd) == "e2e" {
		repoRoot = filepath.Dir(cwd)
	}

	mainGoPath := filepath.Join(repoRoot, "cmd", "bookmux")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Fatalf("Could not find cmd dir at %s", mainGoPath)
	}

	t.Log("Building bookmux binary for tests...")
	buildCmd := exec.Command("go", "build", "-o", exePath, "./cmd/bookmux")
	buildCmd.Dir = repoRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build bookmux: %v (output: %s)", err, string(out))
	}

	for _, book := range books {
		book := book // capture for loop
		t.Run(book.Title, func(t *testing.T) {
			t.Parallel()

			subTmpDir := t.TempDir()

			zipSlug := strings.ReplaceAll(strings.ToLower(book.Title), " ", "_")
			zipUrl := fmt.Sprintf("http://bookmux-demo.s3-website.eu-central-1.amazonaws.com/%s_librivox.zip", zipSlug)
			zipPath := filepath.Join(subTmpDir, "book.zip")

			downloadFile(t, zipUrl, zipPath)

			inputDir := filepath.Join(subTmpDir, "input")
			unzipFile(t, zipPath, inputDir)

			outputM4b := filepath.Join(subTmpDir, "result.m4b")

			runCmd := exec.Command(exePath,
				"--input", inputDir,
				"--output", outputM4b,
				"--title", book.Title,
				"--author", book.Author,
			)

			t.Logf("Running command: %v", runCmd.Args)
			if out, err := runCmd.CombinedOutput(); err != nil {
				t.Fatalf("bookmux command failed: %v (output: %s)", err, string(out))
			}

			// Verify general metadata
			if err := ffmpeg.CheckDependencies(); err != nil {
				t.Fatalf("Failed to initialize ffmpeg dependencies: %v", err)
			}
			_, _, meta, err := ffmpeg.GetAudioInfo(outputM4b)
			if err != nil {
				t.Fatalf("Failed to probe output m4b file: %v", err)
			}
			if meta.Title != book.Title {
				t.Errorf("Expected title %q, got %q", book.Title, meta.Title)
			}
			if meta.Author != book.Author {
				t.Errorf("Expected author %q, got %q", book.Author, meta.Author)
			}

			// Verify chapters
			probeCmd := exec.Command(ffmpeg.FFprobePath, "-v", "quiet", "-show_chapters", "-print_format", "json", outputM4b)
			probeOut, err := probeCmd.Output()
			if err != nil {
				t.Fatalf("ffprobe show_chapters failed: %v", err)
			}

			var probeData FFprobeOutput
			if err := json.Unmarshal(probeOut, &probeData); err != nil {
				t.Fatalf("failed to decode ffprobe chapters output: %v", err)
			}

			if len(probeData.Chapters) != len(book.Chapters) {
				t.Fatalf("Chapter count mismatch: got %d, want %d", len(probeData.Chapters), len(book.Chapters))
			}

			for i, exp := range book.Chapters {
				got := probeData.Chapters[i]

				if got.Tags.Title != exp.Chapter {
					t.Errorf("Chapter %d title mismatch: got %q, want %q", i, got.Tags.Title, exp.Chapter)
				}

				start, _ := strconv.ParseFloat(got.StartTime, 64)
				end, _ := strconv.ParseFloat(got.EndTime, 64)
				gotDur := end - start

				diff := gotDur - float64(exp.DurationSeconds)
				tolerance := 2.0
				
				if diff < -tolerance || diff > tolerance {
					t.Errorf("Chapter %d (Section %d) duration mismatch: got %.2f, want %d", i, exp.Section, gotDur, exp.DurationSeconds)
				}
			}

			t.Logf("E2E subtest for %s passed successfully. Validated %d chapters.", book.Title, len(book.Chapters))
		})
	}
}
