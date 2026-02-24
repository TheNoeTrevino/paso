package daemon

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/logging"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	logging.Disable()
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/thenoetrevino/paso/internal/events.(*Client).listenLoop"),
		goleak.IgnoreTopFunction("github.com/thenoetrevino/paso/internal/events.(*Client).reconnect"),
		goleak.IgnoreTopFunction("github.com/thenoetrevino/paso/internal/events.(*Client).eventBatcher"),
	)
}
