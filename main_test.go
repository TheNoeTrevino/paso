package main

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/logging"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	logging.Disable()
	goleak.VerifyTestMain(m)
}
