package daemon

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	// Ignore goroutines from the events.Client package that are spawned by
	// testutil.SetupTestClient and cleaned up via client.Close() in t.Cleanup
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/thenoetrevino/paso/internal/events.(*Client).listenLoop"),
		goleak.IgnoreTopFunction("github.com/thenoetrevino/paso/internal/events.(*Client).reconnect"),
		goleak.IgnoreTopFunction("github.com/thenoetrevino/paso/internal/events.(*Client).eventBatcher"),
	)
}
