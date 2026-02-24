package install

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

//go:embed skills/paso/SKILL.md
var pasoSkillContent []byte

//go:embed commands/paso-delegate.md
var pasoDelegateContent []byte

//go:embed commands/paso-setup.md
var pasoSetupContent []byte

type installable struct {
	key     string
	label   string
	content []byte
	// pathFunc returns the destination path for a given target's base directory.
	pathFunc func(baseDir string) string
}

var skills = []installable{
	{
		key:     "skill:paso",
		label:   "paso — Complete paso CLI knowledge for AI agents",
		content: pasoSkillContent,
		pathFunc: func(baseDir string) string {
			return filepath.Join(baseDir, "skills", "paso", "SKILL.md")
		},
	},
}

var commands = []installable{
	{
		key:     "cmd:paso-delegate",
		label:   "paso-delegate — Delegate tasks to subagents",
		content: pasoDelegateContent,
		pathFunc: func(baseDir string) string {
			return filepath.Join(baseDir, "commands", "paso-delegate.md")
		},
	},
	{
		key:     "cmd:paso-setup",
		label:   "paso-setup — Organize tasks into EPICs",
		content: pasoSetupContent,
		pathFunc: func(baseDir string) string {
			return filepath.Join(baseDir, "commands", "paso-setup.md")
		},
	},
}

type target struct {
	key     string
	label   string
	baseDir func() (string, error)
}

var targets = []target{
	{
		key:   "opencode",
		label: "OpenCode (~/.config/opencode/)",
		baseDir: func() (string, error) {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, ".config", "opencode"), nil
		},
	},
	{
		key:   "claude",
		label: "Claude Code (~/.claude/)",
		baseDir: func() (string, error) {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, ".claude"), nil
		},
	},
}

func InstallCmd() *cobra.Command {
	var checkFlag, removeFlag, allFlag bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install paso skills and commands for AI tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			if checkFlag {
				return runCheck()
			}
			if removeFlag {
				return runRemove()
			}

			if allFlag {
				return installAll()
			}

			return runInteractive()
		},
	}

	cmd.Flags().BoolVar(&checkFlag, "check", false, "Show installation status")
	cmd.Flags().BoolVar(&removeFlag, "remove", false, "Remove installed files")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Install everything for all targets")

	return cmd
}

func runInteractive() error {
	var selectedSkills []string
	var selectedCommands []string
	var selectedTargets []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Skills").
				Options(
					huh.NewOption("paso — Complete paso CLI knowledge for AI agents", "skill:paso").Selected(true),
				).
				Value(&selectedSkills),

			huh.NewMultiSelect[string]().
				Title("Commands").
				Options(
					huh.NewOption("paso-delegate — Delegate tasks to subagents", "cmd:paso-delegate"),
					huh.NewOption("paso-setup — Organize tasks into EPICs", "cmd:paso-setup"),
				).
				Value(&selectedCommands),

			huh.NewMultiSelect[string]().
				Title("Install for").
				Options(
					huh.NewOption("OpenCode (~/.config/opencode/)", "opencode").Selected(true),
					huh.NewOption("Claude Code (~/.claude/)", "claude"),
				).
				Value(&selectedTargets),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("form error: %v", err))
	}

	selected := append(selectedSkills, selectedCommands...)
	if len(selected) == 0 || len(selectedTargets) == 0 {
		fmt.Println("Nothing selected.")
		return nil
	}

	return installSelected(selected, selectedTargets)
}

func installAll() error {
	allItems := allKeys()
	allTargetKeys := make([]string, len(targets))
	for i, t := range targets {
		allTargetKeys[i] = t.key
	}
	return installSelected(allItems, allTargetKeys)
}

func installSelected(selectedKeys []string, targetKeys []string) error {
	items := resolveItems(selectedKeys)
	resolvedTargets := resolveTargets(targetKeys)

	for _, t := range resolvedTargets {
		baseDir, err := t.baseDir()
		if err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to resolve directory for %s: %v", t.key, err))
		}

		for _, item := range items {
			dest := item.pathFunc(baseDir)
			if err := EnsureDir(filepath.Dir(dest), 0755); err != nil {
				return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to create directory for %s: %v", dest, err))
			}

			if err := atomicWriteFile(dest, item.content); err != nil {
				return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to write %s: %v", dest, err))
			}

			cli.PrintSuccess(fmt.Sprintf("Installed %s → %s", item.key, dest))
		}
	}

	return nil
}

func runCheck() error {
	allItems := allInstallables()

	for _, t := range targets {
		baseDir, err := t.baseDir()
		if err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to resolve directory for %s: %v", t.key, err))
		}

		fmt.Printf("\n%s:\n", t.label)
		for _, item := range allItems {
			dest := item.pathFunc(baseDir)
			status := "not installed"
			if FileExists(dest) {
				status = "installed"
			}
			fmt.Printf("  %-25s %s\n", item.key, status)
		}
	}
	fmt.Println()

	return nil
}

func runRemove() error {
	allItems := allInstallables()
	removed := 0

	for _, t := range targets {
		baseDir, err := t.baseDir()
		if err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to resolve directory for %s: %v", t.key, err))
		}

		for _, item := range allItems {
			dest := item.pathFunc(baseDir)
			if !FileExists(dest) {
				continue
			}

			if err := os.Remove(dest); err != nil {
				return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to remove %s: %v", dest, err))
			}

			cli.PrintSuccess(fmt.Sprintf("Removed %s", dest))
			removed++
		}
	}

	if removed == 0 {
		fmt.Println("Nothing to remove.")
	}

	return nil
}

func allInstallables() []installable {
	all := make([]installable, 0, len(skills)+len(commands))
	all = append(all, skills...)
	all = append(all, commands...)
	return all
}

func allKeys() []string {
	items := allInstallables()
	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = item.key
	}
	return keys
}

func resolveItems(keys []string) []installable {
	all := allInstallables()
	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	resolved := make([]installable, 0, len(keys))
	for _, item := range all {
		if keySet[item.key] {
			resolved = append(resolved, item)
		}
	}
	return resolved
}

func resolveTargets(keys []string) []target {
	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	resolved := make([]target, 0, len(keys))
	for _, t := range targets {
		if keySet[t.key] {
			resolved = append(resolved, t)
		}
	}
	return resolved
}
