package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// GoFmtFormatter handles formatting of Go files using gofmt
type GoFmtFormatter struct{}

func (g *GoFmtFormatter) Name() string {
	return "gofmt"
}

// GetStagedFiles returns a list of staged Go files
func (g *GoFmtFormatter) GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	var goFiles []string
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, file := range files {
		if file != "" && strings.HasSuffix(file, ".go") {
			goFiles = append(goFiles, file)
		}
	}

	return goFiles, nil
}

// Format formats a single Go file using gofmt
func (g *GoFmtFormatter) Format(file string) error {
	// Format the file
	cmd := exec.Command("gofmt", "-w", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}

	// Re-stage the formatted file
	cmd = exec.Command("git", "add", file)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	return nil
}
