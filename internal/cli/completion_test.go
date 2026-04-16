package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteCompletion(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	os.Args = []string{"/tmp/bookmux"}

	tests := []struct {
		shell  string
		wantIn string
	}{
		{shell: "bash", wantIn: "complete -F _bookmux bookmux"},
		{shell: "zsh", wantIn: "#compdef bookmux"},
		{shell: "fish", wantIn: "complete -c bookmux -f"},
	}

	for _, tc := range tests {
		t.Run(tc.shell, func(t *testing.T) {
			output := captureStdout(t, func() {
				if err := WriteCompletion(tc.shell); err != nil {
					t.Fatalf("WriteCompletion returned error: %v", err)
				}
			})
			if !strings.Contains(output, tc.wantIn) {
				t.Fatalf("WriteCompletion output = %q, want substring %q", output, tc.wantIn)
			}
		})
	}
}

func TestWriteCompletionUnsupportedShell(t *testing.T) {
	if err := WriteCompletion("pwsh"); err == nil {
		t.Fatal("WriteCompletion returned nil error for unsupported shell")
	}
}

func TestInstallCompletion(t *testing.T) {
	oldArgs := os.Args
	oldHome := os.Getenv("HOME")
	defer func() {
		os.Args = oldArgs
		if err := os.Setenv("HOME", oldHome); err != nil {
			t.Fatalf("failed to restore HOME: %v", err)
		}
	}()

	tmpHome := t.TempDir()
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	os.Args = []string{"/usr/local/bin/bookmux"}

	tests := []struct {
		shell string
		path  string
	}{
		{shell: "bash", path: filepath.Join(tmpHome, ".bash_completion")},
		{shell: "zsh", path: filepath.Join(tmpHome, ".zsh", "_bookmux")},
		{shell: "fish", path: filepath.Join(tmpHome, ".config", "fish", "completions", "bookmux.fish")},
	}

	for _, tc := range tests {
		t.Run(tc.shell, func(t *testing.T) {
			output := captureStdout(t, func() {
				if err := InstallCompletion(tc.shell); err != nil {
					t.Fatalf("InstallCompletion returned error: %v", err)
				}
			})

			content, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("failed to read completion script: %v", err)
			}
			if len(content) == 0 {
				t.Fatalf("completion script at %s is empty", tc.path)
			}
			if !strings.Contains(output, tc.path) {
				t.Fatalf("InstallCompletion output = %q, want path %q", output, tc.path)
			}
		})
	}
}
