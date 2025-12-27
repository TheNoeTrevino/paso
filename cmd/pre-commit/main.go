package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// ANSI color codes for output
const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

// Formatter interface allows for extensible formatting support
type Formatter interface {
	Name() string
	GetStagedFiles(ctx context.Context) ([]string, error)
	Format(ctx context.Context, file string) error
}

// FormatterResult holds the result of a formatter run
type FormatterResult struct {
	Name           string
	FilesFormatted int
	Error          error
}

// runFormatter runs a single formatter concurrently
func runFormatter(ctx context.Context, formatter Formatter, resultCh chan<- FormatterResult) {
	result := FormatterResult{Name: formatter.Name()}

	files, err := formatter.GetStagedFiles(ctx)
	if err != nil {
		result.Error = err
		resultCh <- result
		return
	}

	if len(files) == 0 {
		resultCh <- result
		return
	}

	// Format each file
	var formatErrors []string
	for _, file := range files {
		if err := formatter.Format(ctx, file); err != nil {
			formatErrors = append(formatErrors, fmt.Sprintf("  %s: %v", file, err))
		} else {
			result.FilesFormatted++
		}
	}

	if len(formatErrors) > 0 {
		result.Error = fmt.Errorf("formatting errors:\n%s", strings.Join(formatErrors, "\n"))
	}

	resultCh <- result
}

func main() {
	ctx := context.Background()
	formatters := []Formatter{
		&GoFmtFormatter{},
		&SqruffFormatter{},
	}

	// Channel to collect results
	resultCh := make(chan FormatterResult, len(formatters))

	// Run all formatters concurrently
	var wg sync.WaitGroup
	for _, formatter := range formatters {
		wg.Add(1)
		go func(f Formatter) {
			defer wg.Done()
			runFormatter(ctx, f, resultCh)
		}(formatter)
	}

	// Wait for all formatters to complete
	wg.Wait()
	close(resultCh)

	// Collect and process results
	var hasError bool
	var totalFormatted int

	for result := range resultCh {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "%s✗ %s failed:%s\n%v\n", colorRed, result.Name, colorReset, result.Error)
			hasError = true
		} else if result.FilesFormatted > 0 {
			fmt.Printf("%s✓ %s:%s formatted %d file(s)\n", colorGreen, result.Name, colorReset, result.FilesFormatted)
			totalFormatted += result.FilesFormatted
		}
	}

	if hasError {
		os.Exit(1)
	}

	if totalFormatted > 0 {
		fmt.Printf("%s✓ Pre-commit formatting complete%s\n", colorGreen, colorReset)
	}

	os.Exit(0)
}
