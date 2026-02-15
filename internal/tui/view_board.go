package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/helpers"
	"github.com/thenoetrevino/paso/internal/tui/notifications"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// buildFilterBarProps constructs the FilterBarProps with resolved display names
// for the currently active filter IDs.
func (m Model) buildFilterBarProps() components.FilterBarProps {
	props := components.FilterBarProps{
		Filter:  m.UI.Filter,
		Focused: m.UIState.Mode == state.FilterBarMode,
		Width:   m.UIState.Width(),
	}

	if m.UI.Filter.PriorityID != nil {
		for _, p := range renderers.GetPriorityOptions() {
			if p.ID == *m.UI.Filter.PriorityID {
				props.PriorityName = p.Description
				break
			}
		}
	}

	if m.UI.Filter.TypeID != nil {
		for _, t := range renderers.GetTypeOptions() {
			if t.ID == *m.UI.Filter.TypeID {
				props.TypeName = t.Description
				break
			}
		}
	}

	if m.UI.Filter.AssigneeID != nil {
		assigneeID := *m.UI.Filter.AssigneeID
		if assigneeID == -1 {
			props.AssigneeName = "Unassigned"
		} else {
			// Resolve assignee name directly from service to avoid dependency on picker state
			ctx, cancel := m.DBContext()
			defer cancel()
			assignee, err := m.App.AssigneeService.GetByID(ctx, assigneeID)
			if err == nil && assignee != nil {
				props.AssigneeName = assignee.Name
			}
			// If error or not found, leave AssigneeName empty (filter ID is still active)
		}
	}

	if len(m.UI.Filter.LabelIDs) > 0 {
		labelMap := make(map[int]string)
		for _, l := range m.AppState.Labels() {
			labelMap[l.ID] = l.Name
		}
		for _, id := range m.UI.Filter.LabelIDs {
			if name, ok := labelMap[id]; ok {
				props.LabelNames = append(props.LabelNames, name)
			}
		}
	}

	return props
}

// getInlineNotification returns the inline notification content for the tab bar
// Returns empty string if no notifications
func (m Model) getInlineNotification() string {
	if !m.UI.Notification.HasAny() {
		return ""
	}
	// Get the first (most recent) notification
	allNotifications := m.UI.Notification.All()
	if len(allNotifications) == 0 {
		return ""
	}
	return notifications.RenderInlineFromState(allNotifications[0])
}

// viewKanbanBoard renders the main kanban board (normal mode)
func (m Model) viewKanbanBoard() string {
	// Check if list view is active
	if m.UI.ListView.IsListView() {
		return m.viewListView()
	}

	// Handle empty column list edge case
	if len(m.AppState.Columns()) == 0 {
		emptyMsg := "No columns found. Please check database initialization."
		footer := components.RenderStatusBar(components.StatusBarProps{
			Width:            m.UIState.Width(),
			ConnectionStatus: m.ConnectionState.Status(),
			DatabaseName:     m.CurrentDBName,
			Tip:              m.UI.CurrentTip,
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
	endIdx := min(m.UIState.ViewportOffset+m.UIState.ViewportSize(), len(m.AppState.Columns()))
	visibleColumns := m.AppState.Columns()[m.UIState.ViewportOffset:endIdx]

	// Calculate fixed content height using shared method
	columnHeight := m.UIState.ContentHeight()

	// Calculate dynamic column width based on available space and visible column count
	columnWidth := m.UIState.ColumnContentWidth(len(visibleColumns))

	// Render only visible columns
	var columns []string
	for i, col := range visibleColumns {
		// Calculate global index for selection check
		globalIndex := m.UIState.ViewportOffset + i

		// Safe map access with defensive check
		tasks, ok := m.AppState.Tasks()[col.ID]
		if !ok {
			tasks = []*models.TaskSummary{}
		}

		// Determine selection state for this column
		isSelected := (globalIndex == m.UIState.SelectedColumn)

		// Determine which task is selected (only for the selected column)
		selectedTaskIdx := -1
		if isSelected {
			selectedTaskIdx = m.UIState.SelectedTask
		}

		scrollOffset := m.UIState.TaskScrollOffset(col.ID)

		columns = append(columns, components.RenderColumn(col, tasks, isSelected, selectedTaskIdx, columnHeight, scrollOffset, columnWidth))
	}

	scrollIndicators := helpers.GetScrollIndicators(
		m.UIState.ViewportOffset,
		m.UIState.ViewportSize(),
		len(m.AppState.Columns()),
	)

	// Layout columns horizontally with scroll indicators
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	board := lipgloss.JoinHorizontal(lipgloss.Top, scrollIndicators.Left, " ", columnsView, " ", scrollIndicators.Right)

	// Add detail panel if screen is wide enough
	if m.UIState.ShouldShowDetailPanel() {
		currentTask := m.getCurrentTask()
		panelHeight := columnHeight
		panelWidth := m.UIState.DetailPanelWidth()

		var detailPanel string
		if currentTask == nil {
			// No task selected - show empty panel
			detailPanel = components.RenderDetailPanel(nil, panelWidth, panelHeight)
		} else if m.DetailPanelLoading {
			// Task detail is being fetched - show loading spinner
			detailPanel = components.RenderDetailPanelLoading(panelWidth, panelHeight, m.SpinnerFrame)
		} else if taskDetail, ok := m.DetailCache.Get(currentTask.ID); ok {
			// Task detail is in cache - render it
			detailPanel = components.RenderDetailPanel(taskDetail, panelWidth, panelHeight)
		} else {
			// Task not in cache yet - show loading
			detailPanel = components.RenderDetailPanelLoading(panelWidth, panelHeight, m.SpinnerFrame)
		}

		board = lipgloss.JoinHorizontal(lipgloss.Top, board, detailPanel)
	}

	// Create project tabs from actual project data
	var projectTabs []string
	for _, project := range m.AppState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	// Get inline notification for tab bar
	inlineNotification := m.getInlineNotification()
	tabBar := components.RenderTabs(projectTabs, m.AppState.SelectedProject(), m.UIState.Width(), inlineNotification)

	filterBar := components.RenderFilterBar(m.buildFilterBarProps())

	footer := components.RenderStatusBar(components.StatusBarProps{
		Width:            m.UIState.Width(),
		SearchMode:       m.UIState.Mode == state.SearchMode || m.UI.Search.IsActive,
		SearchQuery:      m.UI.Search.Query,
		ConnectionStatus: m.ConnectionState.Status(),
		DatabaseName:     m.CurrentDBName,
		Tip:              m.UI.CurrentTip,
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, filterBar, board, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")

	maxContentLines := max(m.UIState.Height-1, 1)

	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + footer

	// If no notifications, return base view directly
	if !m.UI.Notification.HasAny() {
		return baseView
	}

	// Start layer stack with base view
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(baseView),
	}

	canvas := lipgloss.NewCanvas(layers...)
	return canvas.Render()
}

// viewListView renders the list/table view of all tasks.
func (m Model) viewListView() string {
	// Build rows from all tasks across columns (with sorting applied)
	rows := m.buildListViewRows()

	// Calculate fixed content height using shared method
	listHeight := m.UIState.ContentHeight()

	// Render tab bar (same as kanban)
	var projectTabs []string
	for _, project := range m.AppState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	// Get inline notification for tab bar
	inlineNotification := m.getInlineNotification()
	tabBar := components.RenderTabs(projectTabs, m.AppState.SelectedProject(), m.UIState.Width(), inlineNotification)

	filterBar := components.RenderFilterBar(m.buildFilterBarProps())

	// Render list content with sort indicator
	listContent := renderers.RenderListView(
		rows,
		m.UI.ListView.SelectedRow(),
		m.UI.ListView.ScrollOffset(),
		m.UI.ListView.SortField(),
		m.UI.ListView.SortOrder(),
		m.UIState.Width(),
		listHeight,
	)

	statusBar := components.RenderStatusBar(components.StatusBarProps{
		Width:            m.UIState.Width(),
		SearchMode:       m.UIState.Mode == state.SearchMode || m.UI.Search.IsActive,
		SearchQuery:      m.UI.Search.Query,
		ConnectionStatus: m.ConnectionState.Status(),
		DatabaseName:     m.CurrentDBName,
		Tip:              m.UI.CurrentTip,
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, filterBar, listContent, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")
	maxContentLines := max(m.UIState.Height-1, 1)

	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + statusBar

	return baseView
}
