package main

import (
	"context"
	"os"

	"github.com/thenoetrevino/paso/internal/ci"
)

func main() {
	runner := ci.NewRunner()
	exitCode := runner.Run(context.Background())
	os.Exit(exitCode)
}
