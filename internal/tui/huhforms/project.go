package huhforms

import (
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/git"
)

// buildGitBranchOptions builds select options for git branches.
// When editing, ensures the saved branch appears even if deleted from git.
// Displays branches with markers: "* " (current), "+ " (worktree), "  " (normal)
func buildGitBranchOptions(gitBranches []git.BranchInfo, currentBranch *string, isEditing bool) []huh.Option[string] {
	options := []huh.Option[string]{
		huh.NewOption("(none)", ""),
	}

	seen := make(map[string]bool)
	seen[""] = true

	// If editing and saved branch no longer exists, add it with indicator
	if isEditing && currentBranch != nil && *currentBranch != "" {
		found := false
		for _, branch := range gitBranches {
			if branch.Name == *currentBranch {
				found = true
				break
			}
		}
		if !found {
			options = append(options, huh.NewOption("  "+*currentBranch+" (deleted)", *currentBranch))
			seen[*currentBranch] = true
		}
	}

	// Add all git branches with their markers
	for _, branch := range gitBranches {
		if !seen[branch.Name] {
			displayText := branch.Marker + branch.Name
			options = append(options, huh.NewOption(displayText, branch.Name))
			seen[branch.Name] = true
		}
	}

	return options
}

type ProjectFormProps struct {
	Name        *string
	Description *string
	GitBranch   *string
	Confirm     *bool
	IsEditing   bool
	GitBranches []git.BranchInfo
	IsGitRepo   bool
}

func CreateProjectForm(props ProjectFormProps) *huh.Form {
	confirmTitle := "Create this project?"
	if props.IsEditing {
		confirmTitle = "Save changes?"
	}

	fields := []huh.Field{
		huh.NewInput().
			Key("name").
			Title("Project Name").
			Placeholder("Enter project name...").
			Value(props.Name),

		huh.NewText().
			Key("description").
			Title("Description (optional)").
			Placeholder("Enter project description...").
			CharLimit(500).
			Lines(3).
			Value(props.Description),
	}

	// Only add git branch field if in a git repository
	if props.IsGitRepo {
		options := buildGitBranchOptions(props.GitBranches, props.GitBranch, props.IsEditing)
		fields = append(fields, huh.NewSelect[string]().
			Key("gitbranch").
			Title("Git Branch (optional)").
			Options(options...).
			Value(props.GitBranch))
	}

	fields = append(fields, huh.NewConfirm().
		Key("confirm").
		Title(confirmTitle).
		Affirmative("Yes").
		Negative("No").
		Value(props.Confirm))

	form := huh.NewForm(huh.NewGroup(fields...))
	return form.WithKeyMap(CreateKeyMapWithShiftEnter())
}
