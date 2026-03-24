package input

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bookmux/internal/model"
)

// Validate checks the configuration and basic file conditions.
func Validate(cfg *model.BuildConfig) error {
	if cfg.InputPath == "" {
		return fmt.Errorf("input path is required")
	}
	if cfg.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	if strings.ToLower(filepath.Ext(cfg.OutputPath)) != ".m4b" {
		return fmt.Errorf("output file must end with .m4b")
	}

	if !cfg.Overwrite {
		if _, err := os.Stat(cfg.OutputPath); err == nil {
			return fmt.Errorf("output file %s already exists and --overwrite not set", cfg.OutputPath)
		}
	}

	// Validate cover if provided
	if cfg.CoverPath != "" {
		ext := strings.ToLower(filepath.Ext(cfg.CoverPath))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			return fmt.Errorf("cover file must be a .jpg or .png")
		}
		if _, err := os.Stat(cfg.CoverPath); err != nil {
			return fmt.Errorf("cover file not accessible: %w", err)
		}
	}

	return nil
}
