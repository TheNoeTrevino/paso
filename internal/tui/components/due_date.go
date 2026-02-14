package components

import (
	"fmt"
	"math"
	"time"

	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// FormatRelativeDueDate returns a human-readable relative due date string
// and the appropriate theme color based on urgency.
// Returns empty string and empty color if dueDate is nil.
func FormatRelativeDueDate(dueDate *time.Time) (text string, color string) {
	if dueDate == nil {
		return "", ""
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())

	days := int(math.Round(due.Sub(today).Hours() / 24))

	switch {
	case days < 0:
		absDays := -days
		if absDays == 1 {
			return "Overdue by 1 day", theme.ErrorFg
		}
		return fmt.Sprintf("Overdue by %d days", absDays), theme.ErrorFg
	case days == 0:
		return "Due today", theme.WarningFg
	case days == 1:
		return "Due tomorrow", theme.WarningFg
	case days <= 3:
		return fmt.Sprintf("Due in %d days", days), theme.WarningFg
	default:
		return fmt.Sprintf("Due in %d days", days), theme.Highlight
	}
}
