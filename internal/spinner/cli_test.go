package spinner

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSpinner_RendersFrames(t *testing.T) {
	var buf bytes.Buffer
	sp := New("loading...", WithWriter(&buf))

	sp.Start()
	// Let a few frames render
	time.Sleep(350 * time.Millisecond)
	sp.Stop()

	output := buf.String()

	// Should contain the message
	if !strings.Contains(output, "loading...") {
		t.Errorf("expected output to contain message, got: %q", output)
	}

	// Should contain at least one spinner frame icon
	foundFrame := false
	for _, frame := range Frames {
		if strings.Contains(output, frame) {
			foundFrame = true
			break
		}
	}
	if !foundFrame {
		t.Errorf("expected output to contain at least one spinner frame icon, got: %q", output)
	}
}

func TestSpinner_StopClearsLine(t *testing.T) {
	var buf bytes.Buffer
	sp := New("working...", WithWriter(&buf))

	sp.Start()
	time.Sleep(150 * time.Millisecond)
	sp.Stop()

	output := buf.String()

	// After Stop, the last thing written should be a blank line (spaces) followed by \r
	// This clears the spinner from the terminal
	if !strings.HasSuffix(output, "\r") {
		t.Errorf("expected output to end with carriage return after clear, got suffix: %q", output[max(0, len(output)-20):])
	}
}

func TestSpinner_StopIsIdempotent(t *testing.T) {
	var buf bytes.Buffer
	sp := New("test", WithWriter(&buf))

	sp.Start()
	time.Sleep(150 * time.Millisecond)

	// Multiple stops should not panic
	sp.Stop()
	sp.Stop()
	sp.Stop()
}

func TestSpinner_WithWriterOption(t *testing.T) {
	var buf bytes.Buffer
	sp := New("custom writer", WithWriter(&buf))

	if sp.writer != &buf {
		t.Error("expected writer to be the custom buffer")
	}
}
