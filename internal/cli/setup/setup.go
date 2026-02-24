package setup

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
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

func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up paso skills and commands for AI tools",
	}

	for _, t := range targets {
		cmd.AddCommand(targetCmd(t))
	}

	return cmd
}

func targetCmd(t target) *cobra.Command {
	var checkFlag, removeFlag, allFlag, forceFlag bool

	cmd := &cobra.Command{
		Use:   t.key,
		Short: fmt.Sprintf("Set up paso for %s", t.label),
		RunE: func(cmd *cobra.Command, args []string) error {
			if checkFlag {
				return runCheck(t)
			}
			if removeFlag {
				return runRemove(t)
			}
			if allFlag {
				return setupAll(t, forceFlag)
			}
			return runInteractive(t, forceFlag)
		},
	}

	cmd.Flags().BoolVar(&checkFlag, "check", false, "Show setup status")
	cmd.Flags().BoolVar(&removeFlag, "remove", false, "Remove files set up by paso")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Set up all skills and commands")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing files without prompting")

	return cmd
}

func runInteractive(t target, force bool) error {
	var selectedSkills []string
	var selectedCommands []string

	skillOpts := make([]huh.Option[string], len(skills))
	for i, s := range skills {
		skillOpts[i] = huh.NewOption(s.label, s.key).Selected(true)
	}

	cmdOpts := make([]huh.Option[string], len(commands))
	for i, c := range commands {
		cmdOpts[i] = huh.NewOption(c.label, c.key).Selected(true)
	}

	theme := huhforms.CreatePasoTheme(cli.GetColorScheme())

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Skills").
				Options(skillOpts...).
				Height(len(skills)+2).
				Value(&selectedSkills),

			huh.NewMultiSelect[string]().
				Title("Commands").
				Options(cmdOpts...).
				Height(len(commands)+2).
				Value(&selectedCommands),
		),
	).WithTheme(theme)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("form error: %v", err))
	}

	selected := make([]string, 0, len(selectedSkills)+len(selectedCommands))
	selected = append(selected, selectedSkills...)
	selected = append(selected, selectedCommands...)
	if len(selected) == 0 {
		fmt.Println("Nothing selected.")
		return nil
	}

	return setupSelected(selected, t, force)
}

func setupAll(t target, force bool) error {
	return setupSelected(allKeys(), t, force)
}

func setupSelected(selectedKeys []string, t target, force bool) error {
	items := resolveItems(selectedKeys)

	baseDir, err := t.baseDir()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to resolve directory for %s: %v", t.key, err))
	}

	for _, item := range items {
		dest := item.pathFunc(baseDir)

		if FileExists(dest) && !force {
			overwrite := true
			theme := huhforms.CreatePasoTheme(cli.GetColorScheme())
			confirm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("%s already exists. Overwrite?", dest)).
						Value(&overwrite),
				),
			).WithTheme(theme)

			if err := confirm.Run(); err != nil {
				return cli.NewExitErr(cli.ExitError, fmt.Sprintf("prompt error: %v", err))
			}

			if !overwrite {
				fmt.Printf("Skipped %s\n", item.key)
				continue
			}
		}

		if err := EnsureDir(filepath.Dir(dest), 0755); err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to create directory for %s: %v", dest, err))
		}

		if err := atomicWriteFile(dest, item.content); err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to write %s: %v", dest, err))
		}

		cli.PrintSuccess(fmt.Sprintf("Installed %s → %s", item.key, dest))
	}

	return nil
}

func runCheck(t target) error {
	allItems := allInstallables()

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
	fmt.Println()

	return nil
}

func runRemove(t target) error {
	allItems := allInstallables()
	removed := 0

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
