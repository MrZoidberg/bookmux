package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteCompletion prints the completion script for the given shell.
func WriteCompletion(shell string) error {
	binName := filepath.Base(os.Args[0])
	switch shell {
	case "bash":
		fmt.Printf(`_%s() {
    local args=("${COMP_WORDS[@]:1:$COMP_CWORD}")
    local IFS=$'\n'
    COMPREPLY=($(GO_FLAGS_COMPLETION=1 %s "${args[@]}"))
}
complete -F _%s %s
`, binName, binName, binName, binName)
	case "zsh":
		fmt.Printf(`#compdef %s
_%s() {
  local -a completions
  local -a args
  args=("${words[@]:1}")
  completions=("${(@f)$(GO_FLAGS_COMPLETION=1 %s "${args[@]}")}")
  compadd -a completions
}
_%s "$@"
`, binName, binName, binName, binName)
	case "fish":
		fmt.Printf(`complete -c %s -f -a "(set -lx GO_FLAGS_COMPLETION 1; %s (commandline -opc)[2..-1] (commandline -ct))"
`, binName, binName)
	default:
		return fmt.Errorf("unsupported shell: %s. Supported: bash, zsh, fish", shell)
	}
	return nil
}

// InstallCompletion attempts to install the completion script for the current shell.
func InstallCompletion(shell string) error {
	binName := filepath.Base(os.Args[0])
	var path string
	var script string
	var zshDir string

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %v", err)
	}

	bashScript := fmt.Sprintf(`_%s() {
    local args=("${COMP_WORDS[@]:1:$COMP_CWORD}")
    local IFS=$'\n'
    COMPREPLY=($(GO_FLAGS_COMPLETION=1 %s "${args[@]}"))
}
complete -F _%s %s
`, binName, binName, binName, binName)

	zshScript := fmt.Sprintf(`#compdef %s
_%s() {
  local -a completions
  local -a args
  args=("${words[@]:1}")
  completions=("${(@f)$(GO_FLAGS_COMPLETION=1 %s "${args[@]}")}")
  compadd -a completions
}
_%s "$@"
`, binName, binName, binName, binName)

	fishScript := fmt.Sprintf(`complete -c %s -f -a "(set -lx GO_FLAGS_COMPLETION 1; %s (commandline -opc)[2..-1] (commandline -ct))"
`, binName, binName)

	switch shell {
	case "bash":
		path = filepath.Join(home, ".bash_completion")
		script = bashScript
	case "zsh":
		// Zsh is tricky, we'll try to put it in ~/.zsh/_bookmux and tell user to add to fpath
		zshDir = filepath.Join(home, ".zsh")
		if err := os.MkdirAll(zshDir, 0755); err != nil {
			return fmt.Errorf("could not create %s: %v", zshDir, err)
		}
		path = filepath.Join(zshDir, "_"+binName)
		script = zshScript
	case "fish":
		path = filepath.Join(home, ".config", "fish", "completions", binName+".fish")
		script = fishScript
	default:
		return fmt.Errorf("unsupported shell: %s. Supported: bash, zsh, fish", shell)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory %s: %v", dir, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file %s: %v", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(script); err != nil {
		return fmt.Errorf("could not write to file %s: %v", path, err)
	}

	fmt.Printf("Completion script for %s installed to %s\n", shell, path)
	if shell == "zsh" {
		fmt.Println("Please add the following to your .zshrc if you haven't already:")
		fmt.Printf("  fpath=(%s $fpath)\n", zshDir)
		fmt.Println("  autoload -U compinit; compinit")
	} else if shell == "bash" {
		fmt.Println("Please ensure ~/.bash_completion is sourced in your .bashrc (default on many systems).")
	}

	return nil
}
