package ci

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorReset  = "\033[0m"
)

type StepResult struct {
	Name    string
	Passed  bool
	Output  string
	Message string
}

type Runner struct {
	results []StepResult
	mu      sync.Mutex
}

func NewRunner() *Runner {
	return &Runner{
		results: make([]StepResult, 0),
	}
}

func (r *Runner) addResult(result StepResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = append(r.results, result)
}

func (r *Runner) Run() int {
	fmt.Printf("%s======================================%s\n", colorBlue, colorReset)
	fmt.Printf("%s     Running CI/CD Pipeline          %s\n", colorBlue, colorReset)
	fmt.Printf("%s======================================%s\n", colorBlue, colorReset)
	fmt.Println()

	var wg sync.WaitGroup
	steps := []func(){
		r.checkFormat,
		r.runLint,
		r.runTests,
		r.runSecurityScan,
		r.runBuild,
	}

	for _, step := range steps {
		wg.Add(1)
		go func(fn func()) {
			defer wg.Done()
			fn()
		}(step)
	}

	wg.Wait()

	return r.printSummary()
}

func (r *Runner) checkFormat() {
	cmd := exec.Command("gofmt", "-s", "-l", ".")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		r.addResult(StepResult{
			Name:    "Format Check",
			Passed:  false,
			Output:  out.String(),
			Message: "Failed to run gofmt",
		})
		return
	}

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var unformatted []string
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, "crush/") {
			unformatted = append(unformatted, line)
		}
	}

	if len(unformatted) > 0 {
		r.addResult(StepResult{
			Name:    "Format Check",
			Passed:  false,
			Output:  strings.Join(unformatted, "\n"),
			Message: "Files not formatted (run 'gofmt -s -w .')",
		})
	} else {
		r.addResult(StepResult{
			Name:    "Format Check",
			Passed:  true,
			Message: "All files properly formatted",
		})
	}
}

func (r *Runner) runLint() {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		r.addResult(StepResult{
			Name:    "Lint",
			Passed:  false,
			Message: "golangci-lint not found (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)",
		})
		return
	}

	cmd := exec.Command("golangci-lint", "run", "--timeout=5m")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		r.addResult(StepResult{
			Name:    "Lint",
			Passed:  false,
			Output:  out.String(),
			Message: "Lint failed",
		})
	} else {
		r.addResult(StepResult{
			Name:    "Lint",
			Passed:  true,
			Message: "Lint passed",
		})
	}
}

func (r *Runner) runTests() {
	cmd := exec.Command("go", "test", "-race", "-coverprofile=coverage.out", "-covermode=atomic", "./...")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		r.addResult(StepResult{
			Name:    "Test",
			Passed:  false,
			Output:  out.String(),
			Message: "Tests failed",
		})
		return
	}

	r.addResult(StepResult{
		Name:    "Test",
		Passed:  true,
		Message: "Tests passed",
	})

	// Check coverage
	r.checkCoverage()
}

func (r *Runner) checkCoverage() {
	cmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		r.addResult(StepResult{
			Name:    "Coverage Threshold",
			Passed:  false,
			Message: "Failed to read coverage",
		})
		return
	}

	lines := strings.Split(out.String(), "\n")
	var coverage float64
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				fmt.Sscanf(fields[2], "%f%%", &coverage)
			}
		}
	}

	if coverage < 40.0 {
		r.addResult(StepResult{
			Name:    "Coverage Threshold",
			Passed:  false,
			Message: fmt.Sprintf("Coverage is %.1f%% - below 40%% threshold", coverage),
		})
	} else {
		r.addResult(StepResult{
			Name:    "Coverage Threshold",
			Passed:  true,
			Message: fmt.Sprintf("Coverage is %.1f%% - meets threshold", coverage),
		})
	}
}

func (r *Runner) runSecurityScan() {
	if _, err := exec.LookPath("govulncheck"); err != nil {
		r.addResult(StepResult{
			Name:    "Security Scan",
			Passed:  false,
			Message: "govulncheck not found (install with: go install golang.org/x/vuln/cmd/govulncheck@latest)",
		})
		return
	}

	cmd := exec.Command("govulncheck", "./...")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		r.addResult(StepResult{
			Name:    "Security Scan",
			Passed:  false,
			Output:  out.String(),
			Message: "Security scan failed",
		})
	} else {
		r.addResult(StepResult{
			Name:    "Security Scan",
			Passed:  true,
			Message: "Security scan passed",
		})
	}
}

func (r *Runner) runBuild() {
	cmd := exec.Command("go", "build", "-o", "bin/paso", ".")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		r.addResult(StepResult{
			Name:    "Build",
			Passed:  false,
			Output:  out.String(),
			Message: "Build failed",
		})
		return
	}

	r.addResult(StepResult{
		Name:    "Build",
		Passed:  true,
		Message: "Build successful",
	})

	// Verify binary
	if _, err := os.Stat("bin/paso"); os.IsNotExist(err) {
		r.addResult(StepResult{
			Name:    "Build Verification",
			Passed:  false,
			Message: "Binary not created",
		})
	} else {
		r.addResult(StepResult{
			Name:    "Build Verification",
			Passed:  true,
			Message: "Binary verified",
		})
	}
}

func (r *Runner) printSummary() int {
	fmt.Println()
	fmt.Printf("%s======================================%s\n", colorBlue, colorReset)
	fmt.Printf("%s          CI/CD Summary               %s\n", colorBlue, colorReset)
	fmt.Printf("%s======================================%s\n", colorBlue, colorReset)
	fmt.Println()

	passedCount := 0
	failedCount := 0

	// Sort results: passed first, then failed
	var passed []StepResult
	var failed []StepResult

	for _, result := range r.results {
		if result.Passed {
			passed = append(passed, result)
			passedCount++
		} else {
			failed = append(failed, result)
			failedCount++
		}
	}

	for _, result := range passed {
		fmt.Printf("%s✅ PASS%s  %s\n", colorGreen, colorReset, result.Name)
	}

	for _, result := range failed {
		fmt.Printf("%s❌ FAIL%s  %s", colorRed, colorReset, result.Name)
		if result.Message != "" {
			fmt.Printf(" - %s", result.Message)
		}
		fmt.Println()
		if result.Output != "" && len(result.Output) > 0 {
			fmt.Printf("%s%s%s\n", colorYellow, result.Output, colorReset)
		}
	}

	fmt.Println()
	fmt.Printf("%s======================================%s\n", colorBlue, colorReset)

	totalSteps := passedCount + failedCount
	if failedCount == 0 {
		fmt.Printf("%s     ✅ All CI/CD steps passed!       %s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("%s     ❌ CI/CD Pipeline Failed          %s\n", colorRed, colorReset)
		fmt.Printf("%s     Failed: %d/%d steps%s\n", colorRed, failedCount, totalSteps, colorReset)
	}
	fmt.Printf("%s======================================%s\n", colorBlue, colorReset)

	if failedCount > 0 {
		return 1
	}
	return 0
}
