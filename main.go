package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli/column"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/launcher"
)

var (
	// Version information (set via ldflags during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "paso",
	Short: "Terminal-based Kanban board with CLI and TUI",
	Long: `Paso is a zero-setup, terminal-based kanban board for personal task management.

Use 'paso tui' to launch the interactive TUI.
Use 'paso task create ...' for CLI commands.`,
	Version: version,
	// No Run function - shows help text by default
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Set version template to include build info
	rootCmd.SetVersionTemplate(fmt.Sprintf("paso version %s\n  commit: %s\n  built: %s\n", version, commit, date))

	// Add CLI subcommands
	rootCmd.AddCommand(task.TaskCmd())
	rootCmd.AddCommand(project.ProjectCmd())
	rootCmd.AddCommand(column.ColumnCmd())
	rootCmd.AddCommand(label.LabelCmd())

	// Add TUI subcommand
	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		Long:  "Launch the interactive terminal user interface for managing tasks visually.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return launcher.Launch()
		},
	}
	rootCmd.AddCommand(tuiCmd)

	// Add completion command
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for paso.

To load completions:

Bash:
  $ source <(paso completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ paso completion bash > /etc/bash_completion.d/paso
  # macOS:
  $ paso completion bash > $(brew --prefix)/etc/bash_completion.d/paso

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ paso completion zsh > "${fpath[1]}/_paso"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ paso completion fish | source

  # To load completions for each session, execute once:
  $ paso completion fish > ~/.config/fish/completions/paso.fish

PowerShell:
  PS> paso completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> paso completion powershell > paso.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell type: %s", args[0])
			}
		},
	}
	rootCmd.AddCommand(completionCmd)
}
