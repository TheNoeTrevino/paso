package tui

// ScrollIndicators holds the left and right scroll arrow indicators
type ScrollIndicators struct {
	Left  string
	Right string
}

// GetScrollIndicators returns the appropriate scroll arrows based on viewport position
func GetScrollIndicators(viewportOffset, viewportSize, columnCount int) ScrollIndicators {
	return ScrollIndicators{
		Left:  getLeftArrow(viewportOffset),
		Right: getRightArrow(viewportOffset, viewportSize, columnCount),
	}
}

// getLeftArrow returns "◀" if there are columns to the left, otherwise space
func getLeftArrow(viewportOffset int) string {
	if viewportOffset > 0 {
		return "◀"
	}
	return " "
}

// getRightArrow returns "▶" if there are columns to the right, otherwise space
func getRightArrow(viewportOffset, viewportSize, columnCount int) string {
	if viewportOffset+viewportSize < columnCount {
		return "▶"
	}
	return " "
}
