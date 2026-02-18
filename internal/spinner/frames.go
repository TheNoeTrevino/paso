// Package spinner provides shared spinner animation frames and a CLI spinner.
package spinner

// Frames contains the moon phase icons used for spinner animations.
// Shared between TUI and CLI for visual consistency.
var Frames = []string{
	"\U000f02d9", "\U000f0ac3", "\U000f0ac4", "\U000f0ac5", "\U000f0ac6", "\U000f0ac7", "\U000f0ac8", "\U000f0ac7", "\U000f0ac6", "\U000f0ac5", "\U000f0ac4", "\U000f0ac3",
}

// FrameCount returns the total number of spinner animation frames.
func FrameCount() int {
	return len(Frames)
}
