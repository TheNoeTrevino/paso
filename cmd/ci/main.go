package main

import (
	"os"

	"github.com/thenoetrevino/paso/internal/ci"
)

func main() {
	runner := ci.NewRunner()
	exitCode := runner.Run()
	os.Exit(exitCode)
}
