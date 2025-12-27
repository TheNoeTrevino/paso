package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// SqruffFormatter handles formatting of SQL files using sqruff
type SqruffFormatter struct{}

func (s *SqruffFormatter) Name() string {
	return "sqruff"
}

// GetStagedFiles returns a list of staged SQL files
func (s *SqruffFormatter) GetStagedFiles(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	var sqlFiles []string
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, file := range files {
		if file != "" && strings.HasSuffix(file, ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}

	return sqlFiles, nil
}

// Format formats a single SQL file using sqruff
func (s *SqruffFormatter) Format(ctx context.Context, file string) error {
	// Format the file
	cmd := exec.CommandContext(ctx, "sqruff", "fix", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sqruff fix failed: %w", err)
	}

	// Re-stage the formatted file
	cmd = exec.CommandContext(ctx, "git", "add", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	return nil
}
