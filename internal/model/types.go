package model

// InputTrack represents a single audio file to be merged.
type InputTrack struct {
	Path       string
	BaseName   string
	Chapter    string
	DurationMs int64
	Bitrate    string
}

// Chapter represents a chapter's metadata in the resulting audiobook.
type Chapter struct {
	Title      string
	StartMs    int64
	DurationMs int64
}

// BookMetadata contains descriptive information for the audiobook.
type BookMetadata struct {
	Title  string
	Author string
	Album  string
	Cover  string
}

// BuildConfig holds all command-line flags and derived configuration.
type BuildConfig struct {
	InputPath  string `long:"input" description:"input directory or comma-separated files"`
	OutputPath string `long:"output" description:"output .m4b file"`
	Title      string `long:"title" description:"audiobook title"`
	Author     string `long:"author" description:"author / artist"`
	Album      string `long:"album" description:"album name"`
	CoverPath  string `long:"cover" description:"optional cover image"`
	Recursive  bool   `long:"recursive" description:"scan directories recursively"`
	Normalize  bool   `long:"normalize" description:"enable loudness normalization"`
	Chapters   string `long:"chapters" description:"none | from-files" default:"from-files"`
	Mono       bool   `long:"mono" description:"convert output to mono"`
	Bitrate    string `long:"bitrate" description:"AAC bitrate, e.g. 64k / 96k / 128k"`
	DryRun     bool   `long:"dry-run" description:"print planned actions only"`
	Overwrite  bool   `long:"overwrite" description:"overwrite output if exists"`
	TempDir    string `long:"temp-dir" description:"custom temp location"`
	LogFile    string `long:"log" description:"write logs to file"`
	Verbose    bool   `long:"verbose" description:"verbose logs"`
	Version    bool   `short:"v" long:"version" description:"show version and exit"`
	Shell      string `long:"completion" description:"generate completion script for bash, zsh, or fish"`
	Install    bool   `long:"install" description:"install completion script for the current shell"`
}
