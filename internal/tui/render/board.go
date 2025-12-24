package render

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/helpers"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/notifications"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// getInlineNotification returns the inline notification content for the tab bar
// Returns empty string if no notifications
func getInlineNotification(m *tui.Model) string {
	if !m.NotificationState.HasAny() {
		return ""
	}
	// Get the first (most recent) notification
	allNotifications := m.NotificationState.All()
	if len(allNotifications) == 0 {
		return ""
	}
	return notifications.RenderInlineFromState(allNotifications[0])
}

// ViewKanbanBoard renders the main kanban board (normal mode)
func ViewKanbanBoard(m *tui.Model) string {
	// Check if list view is active
	if m.ListViewState.IsListView() {
		return viewListView(m)
	}

	// Handle empty column list edge case
	if len(m.AppState.Columns()) == 0 {
		emptyMsg := "No columns found. Please check database initialization."
		footer := components.RenderStatusBar(components.StatusBarProps{
			Width:            m.UiState.Width(),
			ConnectionStatus: m.ConnectionState.Status(),
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
	endIdx := min(m.UiState.ViewportOffset()+m.UiState.ViewportSize(), len(m.AppState.Columns()))
	visibleColumns := m.AppState.Columns()[m.UiState.ViewportOffset():endIdx]

	// Calculate fixed content height using shared method
	columnHeight := m.UiState.ContentHeight()

	// Render only visible columns
	var columns []string
	for i, col := range visibleColumns {
		// Calculate global index for selection check
		globalIndex := m.UiState.ViewportOffset() + i

		// Safe map access with defensive check
		tasks, ok := m.AppState.Tasks()[col.ID]
		if !ok {
			tasks = []*models.TaskSummary{}
		}

		// Determine selection state for this column
		isSelected := (globalIndex == m.UiState.SelectedColumn())

		// Determine which task is selected (only for the selected column)
		selectedTaskIdx := -1
		if isSelected {
			selectedTaskIdx = m.UiState.SelectedTask()
		}

		// Get scroll offset for this column
		scrollOffset := m.UiState.TaskScrollOffset(col.ID)

		columns = append(columns, components.RenderColumn(col, tasks, isSelected, selectedTaskIdx, columnHeight, scrollOffset))
	}

	scrollIndicators := helpers.GetScrollIndicators(
		m.UiState.ViewportOffset(),
		m.UiState.ViewportSize(),
		len(m.AppState.Columns()),
	)

	// Layout columns horizontally with scroll indicators
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	board := lipgloss.JoinHorizontal(lipgloss.Top, scrollIndicators.Left, " ", columnsView, " ", scrollIndicators.Right)

	// Create project tabs from actual project data
	var projectTabs []string
	for _, project := range m.AppState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	// Get inline notification for tab bar
	inlineNotification := getInlineNotification(m)
	tabBar := components.RenderTabs(projectTabs, m.AppState.SelectedProject(), m.UiState.Width(), inlineNotification)

	footer := components.RenderStatusBar(components.StatusBarProps{
		Width:            m.UiState.Width(),
		SearchMode:       m.UiState.Mode() == state.SearchMode || m.SearchState.IsActive,
		SearchQuery:      m.SearchState.Query,
		ConnectionStatus: m.ConnectionState.Status(),
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, board, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")

	maxContentLines := max(m.UiState.Height()-1, 1)

	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + footer

	// If no notifications, return base view directly
	if !m.NotificationState.HasAny() {
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
func viewListView(m *tui.Model) string {
	// Build rows from all tasks across columns (with sorting applied)
	rows := modelops.BuildListViewRows(m)

	// Calculate fixed content height using shared method
	listHeight := m.UiState.ContentHeight()

	// Render tab bar (same as kanban)
	var projectTabs []string
	for _, project := range m.AppState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	// Get inline notification for tab bar
	inlineNotification := getInlineNotification(m)
	tabBar := components.RenderTabs(projectTabs, m.AppState.SelectedProject(), m.UiState.Width(), inlineNotification)

	// Render list content with sort indicator
	listContent := renderers.RenderListView(
		rows,
		m.ListViewState.SelectedRow(),
		m.ListViewState.ScrollOffset(),
		m.ListViewState.SortField(),
		m.ListViewState.SortOrder(),
		m.UiState.Width(),
		listHeight,
	)

	// Render footer
	footer := components.RenderStatusBar(components.StatusBarProps{
		Width:            m.UiState.Width(),
		SearchMode:       m.UiState.Mode() == state.SearchMode || m.SearchState.IsActive,
		SearchQuery:      m.SearchState.Query,
		ConnectionStatus: m.ConnectionState.Status(),
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, listContent, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")
	maxContentLines := max(m.UiState.Height()-1, 1)

	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + footer

	// Notifications are now rendered inline with tabs
	return baseView
}
