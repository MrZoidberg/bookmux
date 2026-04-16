package cli

import (
	"os"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	os.Args = []string{
		"bookmux",
		"--input", "/tmp/in",
		"--output", "/tmp/out.m4b",
		"--title", "Example",
		"--recursive",
		"--overwrite",
	}

	cfg, err := ParseFlags("")
	if err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}

	if cfg.InputPath != "/tmp/in" || cfg.OutputPath != "/tmp/out.m4b" || cfg.Title != "Example" {
		t.Fatalf("ParseFlags returned unexpected config: %+v", cfg)
	}
	if !cfg.Recursive || !cfg.Overwrite {
		t.Fatalf("ParseFlags did not set boolean flags: %+v", cfg)
	}
}

func TestParseFlagsHelpWritesUsage(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	os.Args = []string{"bookmux", "--help"}
	output := captureStdout(t, func() {
		cfg, err := ParseFlags("")
		if cfg != nil {
			t.Fatalf("ParseFlags returned config on help: %+v", cfg)
		}
		if err == nil {
			t.Fatal("ParseFlags returned nil error for help")
		}
	})

	if !strings.Contains(output, "A CLI tool for merging audio tracks into M4B audiobooks.") {
		t.Fatalf("help output missing long description: %q", output)
	}
}
