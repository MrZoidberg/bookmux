# BookMux

A CLI tool for merging audio tracks into audiobooks.

## Features
- Merges multiple audio files into a single `.m4b`.
- Support for chapters from source files.
- Loudness normalization.
- Mono/Bitrate control.
- Built-in TUI progress bar.

## Installation

### macOS & Linux
```bash
curl -sSfL https://raw.githubusercontent.com/MrZoidberg/bookmux/main/install.sh | sh
```
Works across `bash`, `zsh`, and `fish`.

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/MrZoidberg/bookmux/main/install.ps1 | iex
```

### Shell Completions

BookMux can generate shell completions for `bash`, `zsh`, and `fish`.

To install completions for your current shell, run:
```bash
bookmux --completion <bash|zsh|fish> --install
```

After installation, restart your shell or source your configuration file.

## Usage

Merge a directory of audio files into an M4B audiobook:
```bash
bookmux --input ./my-book --output my-book.m4b --title "The Title" --author "The Author"
```

### Options
- `--input`: Directoy or comma-separated list of files.
- `--output`: Resulting M4B file path.
- `--normalize`: Apply loudness normalization to all tracks.
- `--mono`: Convert output to mono (saves space).
- `--bitrate`: Override AAC bitrate (e.g., `64k`, `96k`).
- `--chapters`: `from-files` (default) or `none`.

Run `bookmux --help` for the full list of options.

## License
MIT
