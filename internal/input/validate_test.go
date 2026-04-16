package input

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bookmux/internal/model"
)

func TestValidate(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "book.m4b")
	validCover := filepath.Join(tmpDir, "cover.jpg")
	if err := os.WriteFile(validCover, []byte("cover"), 0o600); err != nil {
		t.Fatalf("failed to create cover: %v", err)
	}

	existingOutput := filepath.Join(tmpDir, "existing.m4b")
	if err := os.WriteFile(existingOutput, []byte("output"), 0o600); err != nil {
		t.Fatalf("failed to create output: %v", err)
	}

	tests := []struct {
		name    string
		cfg     model.BuildConfig
		wantErr string
	}{
		{
			name:    "missing input",
			cfg:     model.BuildConfig{OutputPath: outputPath},
			wantErr: "input path is required",
		},
		{
			name:    "missing output",
			cfg:     model.BuildConfig{InputPath: tmpDir},
			wantErr: "output path is required",
		},
		{
			name:    "invalid output extension",
			cfg:     model.BuildConfig{InputPath: tmpDir, OutputPath: filepath.Join(tmpDir, "book.mp3")},
			wantErr: "output file must end with .m4b",
		},
		{
			name:    "existing output without overwrite",
			cfg:     model.BuildConfig{InputPath: tmpDir, OutputPath: existingOutput},
			wantErr: "already exists",
		},
		{
			name: "invalid cover extension",
			cfg: model.BuildConfig{
				InputPath:  tmpDir,
				OutputPath: outputPath,
				CoverPath:  filepath.Join(tmpDir, "cover.gif"),
				Overwrite:  true,
			},
			wantErr: "cover file must be a .jpg or .png",
		},
		{
			name: "missing cover file",
			cfg: model.BuildConfig{
				InputPath:  tmpDir,
				OutputPath: outputPath,
				CoverPath:  filepath.Join(tmpDir, "missing.png"),
				Overwrite:  true,
			},
			wantErr: "cover file not accessible",
		},
		{
			name: "valid config",
			cfg: model.BuildConfig{
				InputPath:  tmpDir,
				OutputPath: outputPath,
				CoverPath:  validCover,
				Overwrite:  true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(&tc.cfg)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("Validate returned nil error, want %q", tc.wantErr)
				}
				if got := err.Error(); got == "" || !strings.Contains(got, tc.wantErr) {
					t.Fatalf("Validate error = %q, want substring %q", got, tc.wantErr)
				}
			}
		})
	}
}
