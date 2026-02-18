package spinner

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const tickInterval = 100 * time.Millisecond

// Spinner animates a terminal spinner for long-running CLI operations.
// It writes to stderr by default so it never corrupts piped stdout.
type Spinner struct {
	message string
	writer  io.Writer
	done    chan struct{}
	wg      sync.WaitGroup
	stopped sync.Once
	lineLen int
}

// Option configures a Spinner.
type Option func(*Spinner)

// WithWriter overrides the default stderr output (useful for testing).
func WithWriter(w io.Writer) Option {
	return func(s *Spinner) {
		s.writer = w
	}
}

// New creates a spinner that displays the given message alongside a moon-phase animation.
func New(message string, opts ...Option) *Spinner {
	s := &Spinner{
		message: message,
		writer:  os.Stderr,
		done:    make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start begins the spinner animation in a background goroutine.
func (s *Spinner) Start() {
	s.wg.Add(1)
	go s.run()
}

// Stop halts the animation and clears the spinner line.
// Safe to call multiple times.
func (s *Spinner) Stop() {
	s.stopped.Do(func() {
		close(s.done)
		s.wg.Wait()
		s.clear()
	})
}

func (s *Spinner) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	frame := 0
	for {
		s.render(frame)
		frame = (frame + 1) % len(Frames)

		select {
		case <-s.done:
			return
		case <-ticker.C:
		}
	}
}

func (s *Spinner) render(frame int) {
	line := fmt.Sprintf("\r%s %s", Frames[frame], s.message)
	s.lineLen = len(line) - 1 // subtract \r
	_, _ = fmt.Fprint(s.writer, line)
}

func (s *Spinner) clear() {
	if s.lineLen > 0 {
		_, _ = fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", s.lineLen))
	}
}
