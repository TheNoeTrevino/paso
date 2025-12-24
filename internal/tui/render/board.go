package render

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/notifications"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// getInlineNotification returns the inline notification content for the tab bar
// Returns empty string if no notifications
func (w *Wrapper) getInlineNotification() string {
	if !w.NotificationState.HasAny() {
		return ""
	}
	// Get the first (most recent) notification
	allNotifications := w.NotificationState.All()
	if len(allNotifications) == 0 {
		return ""
	}
	return notifications.RenderInlineFromState(allNotifications[0])
}

// ViewKanbanBoard renders the main kanban board (normal mode)
func (w *Wrapper) ViewKanbanBoard() string {
	// Check if list view is active
	if w.ListViewState.IsListView() {
		return w.viewListView()
	}

	// Handle empty column list edge case
	if len(w.AppState.Columns()) == 0 {
		emptyMsg := "No columns found. Please check database initialization."
		footer := components.RenderStatusBar(components.StatusBarProps{
			Width:            w.UiState.Width(),
			ConnectionStatus: w.ConnectionState.Status(),
		})
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			emptyMsg,
			"",
			footer,
		)
	}

	// Calculate visible columns based on viewport
	endIdx := min(w.UiState.ViewportOffset()+w.UiState.ViewportSize(), len(w.AppState.Columns()))
	visibleColumns := w.AppState.Columns()[w.UiState.ViewportOffset():endIdx]

	// Calculate fixed content height using shared method
	columnHeight := w.UiState.ContentHeight()

	// Render only visible columns
	var columns []string
	for i, col := range visibleColumns {
		// Calculate global index for selection check
		globalIndex := w.UiState.ViewportOffset() + i

		// Safe map access with defensive check
		tasks, ok := w.AppState.Tasks()[col.ID]
		if !ok {
			tasks = []*models.TaskSummary{}
		}

		// Determine selection state for this column
		isSelected := (globalIndex == w.UiState.SelectedColumn())

		// Determine which task is selected (only for the selected column)
		selectedTaskIdx := -1
		if isSelected {
			selectedTaskIdx = w.UiState.SelectedTask()
		}

		// Get scroll offset for this column
		scrollOffset := w.UiState.TaskScrollOffset(col.ID)

		columns = append(columns, components.RenderColumn(col, tasks, isSelected, selectedTaskIdx, columnHeight, scrollOffset))
	}

	scrollIndicators := tui.GetScrollIndicators(
		w.UiState.ViewportOffset(),
		w.UiState.ViewportSize(),
		len(w.AppState.Columns()),
	)

	// Layout columns horizontally with scroll indicators
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	board := lipgloss.JoinHorizontal(lipgloss.Top, scrollIndicators.Left, " ", columnsView, " ", scrollIndicators.Right)

	// Create project tabs from actual project data
	var projectTabs []string
	for _, project := range w.AppState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	// Get inline notification for tab bar
	inlineNotification := w.getInlineNotification()
	tabBar := components.RenderTabs(projectTabs, w.AppState.SelectedProject(), w.UiState.Width(), inlineNotification)

	footer := components.RenderStatusBar(components.StatusBarProps{
		Width:            w.UiState.Width(),
		SearchMode:       w.UiState.Mode() == state.SearchMode || w.SearchState.IsActive,
		SearchQuery:      w.SearchState.Query,
		ConnectionStatus: w.ConnectionState.Status(),
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, board, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")

	maxContentLines := max(w.UiState.Height()-1, 1)

	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + footer

	// If no notifications, return base view directly
	if !w.NotificationState.HasAny() {
		return baseView
	}

	// Start layer stack with base view
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(baseView),
	}

	// Notifications are now rendered inline with tabs, no need for floating layers

	// Combine all layers into canvas
	canvas := lipgloss.NewCanvas(layers...)
	return canvas.Render()
}

// viewListView renders the list/table view of all tasks.
func (w *Wrapper) viewListView() string {
	// Build rows from all tasks across columns (with sorting applied)
	ops := modelops.New(w.Model)
	rows := ops.BuildListViewRows()

	// Calculate fixed content height using shared method
	listHeight := w.UiState.ContentHeight()

	// Render tab bar (same as kanban)
	var projectTabs []string
	for _, project := range w.AppState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	// Get inline notification for tab bar
	inlineNotification := w.getInlineNotification()
	tabBar := components.RenderTabs(projectTabs, w.AppState.SelectedProject(), w.UiState.Width(), inlineNotification)

	// Render list content with sort indicator
	listContent := tui.RenderListView(
		rows,
		w.ListViewState.SelectedRow(),
		w.ListViewState.ScrollOffset(),
		w.ListViewState.SortField(),
		w.ListViewState.SortOrder(),
		w.UiState.Width(),
		listHeight,
	)

	// Render footer
	footer := components.RenderStatusBar(components.StatusBarProps{
		Width:            w.UiState.Width(),
		SearchMode:       w.UiState.Mode() == state.SearchMode || w.SearchState.IsActive,
		SearchQuery:      w.SearchState.Query,
		ConnectionStatus: w.ConnectionState.Status(),
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, listContent, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")
	maxContentLines := max(w.UiState.Height()-1, 1)

	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + footer

	// Notifications are now rendered inline with tabs
	return baseView
}
