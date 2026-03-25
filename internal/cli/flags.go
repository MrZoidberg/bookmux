package cli

import (
	"bookmux/internal/model"
	"os"

	"github.com/jessevdk/go-flags"
)

// ParseFlags parses command line arguments and returns the BuildConfig
// or an error if arguments are invalid or if a user requests help.
func ParseFlags(_ string) (*model.BuildConfig, error) {
	var cfg model.BuildConfig
	parser := flags.NewParser(&cfg, flags.HelpFlag|flags.PassDoubleDash)
	parser.ShortDescription = "BookMux"
	parser.LongDescription = "A CLI tool for merging audio tracks into audiobooks."

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
		}
		return nil, err
	}

	return &cfg, nil
}
