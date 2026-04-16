package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.txt")
	lines := []string{"alpha", "beta", "gamma"}

	if err := WriteLines(path, lines); err != nil {
		t.Fatalf("WriteLines returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if got, want := string(content), strings.Join(lines, "\n")+"\n"; got != want {
		t.Fatalf("file contents = %q, want %q", got, want)
	}
}
