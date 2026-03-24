package cli

import (
	"bookmux/internal/model"

	"github.com/jessevdk/go-flags"
)

// ParseFlags parses command line arguments and returns the BuildConfig
// or an error if arguments are invalid or if a user requests help.
func ParseFlags(_ string) (*model.BuildConfig, error) {
	var cfg model.BuildConfig
	parser := flags.NewParser(&cfg, flags.Default)
	parser.ShortDescription = "BookMux"
	parser.LongDescription = "A CLI tool for merging audio tracks into audiobooks."

	_, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
