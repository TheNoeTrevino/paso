package tui

import (
	"fmt"
	"log/slog"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/spinner"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/layers"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderTaskFormLayer renders the task creation/edit form modal as a layer
func (m Model) renderTaskFormLayer() *lipgloss.Layer {
	if m.Forms.Form.TaskForm == nil {
		return nil
	}

	const chromeHeight = 6 // border (2) + padding (2) + title (1) + blanks (1) = 6 lines

	layerWidth := m.UIState.Width() * 8 / 10
	layerHeight := m.UIState.Height * 8 / 10

	innerHeight := layerHeight - chromeHeight

	leftColumnWidth := layerWidth * 6 / 10
	rightColumnWidth := layerWidth * 4 / 10
	topLeftHeight := innerHeight * 7 / 10
	bottomLeftHeight := innerHeight * 3 / 10
	rightColumnHeight := innerHeight

	topLeftZone := m.renderFormTitleDescriptionZone(leftColumnWidth, topLeftHeight)
	bottomLeftZone := m.renderFormCommentsPreview(leftColumnWidth, bottomLeftHeight)

	leftColumn := lipgloss.JoinVertical(lipgloss.Top, topLeftZone, bottomLeftZone)
	rightColumn := m.renderFormMetadataZone(rightColumnWidth, rightColumnHeight)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	helpHintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	var formTitle string
	if m.Forms.Form.EditingTaskID == 0 {
		formTitle = titleStyle.Render("Create New Task")
	} else {
		formTitle = titleStyle.Render("Edit Task")
	}

	titleWithHint := lipgloss.JoinHorizontal(
		lipgloss.Left,
		formTitle,
		"  ",
		helpHintStyle.Render("|"),
		"  ",
		helpHintStyle.Render("Ctrl+H: help"),
	)

	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleWithHint,
		"",
		content,
	)

	formBox := components.FormBoxStyle.
		Width(layerWidth).
		Height(layerHeight).
		Render(fullContent)

	return layers.CreateCenteredLayer(formBox, m.UIState.Width(), m.UIState.Height)
}

// renderProjectFormLayer renders the project creation/edit form modal as a layer
func (m Model) renderProjectFormLayer() *lipgloss.Layer {
	if m.Forms.Form.ProjectForm == nil {
		slog.Debug("renderProjectFormLayer called with nil form", "mode", m.UIState.Mode)
		return nil
	}

	formView := m.Forms.Form.ProjectForm.View()

	var title string
	if m.UIState.Mode == state.EditProjectFormMode {
		title = "Edit Project"
	} else {
		title = "New Project"
	}

	refreshHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Render("Press F5 to refresh git branches")

	content := title + "\n\n" + formView + "\n\n" + refreshHint

	formBox := components.ProjectFormBoxStyle.
		Width(m.UIState.Width() * 3 / 4).
		Height(m.UIState.Height / 3).
		Render(content)

	return layers.CreateCenteredLayer(formBox, m.UIState.Width(), m.UIState.Height)
}

// renderTicketFormLoadingLayer renders a loading indicator while task details are being fetched
func (m Model) renderTicketFormLoadingLayer() *lipgloss.Layer {
	spinnerIcon := spinner.Frames[m.SpinnerFrame%spinner.FrameCount()]

	loadingText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true).
		Render("Loading task details " + spinnerIcon)

	formBox := components.FormBoxStyle.
		Width(m.UIState.Width() * 8 / 10).
		Height(m.UIState.Height * 8 / 10).
		Render(loadingText)

	return layers.CreateCenteredLayer(formBox, m.UIState.Width(), m.UIState.Height)
}

// renderProjectFormLoadingLayer renders a loading indicator while git data is being fetched
func (m Model) renderProjectFormLoadingLayer() *lipgloss.Layer {
	var title string
	if m.Forms.Form.Git.EditingProjectID != 0 {
		title = "Edit Project"
	} else {
		title = "New Project"
	}

	spinnerIcon := spinner.Frames[m.SpinnerFrame%spinner.FrameCount()]

	loadingText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true).
		Render("Loading git branches " + spinnerIcon)

	formBox := components.ProjectFormBoxStyle.
		Width(m.UIState.Width() * 3 / 4).
		Height(m.UIState.Height / 3).
		Render(title + "\n\n" + loadingText)

	return layers.CreateCenteredLayer(formBox, m.UIState.Width(), m.UIState.Height)
}

// renderColumnFormLayer renders the column creation/rename form modal as a layer
func (m Model) renderColumnFormLayer() *lipgloss.Layer {
	if m.Forms.Form.ColumnForm == nil {
		return nil
	}

	formView := m.Forms.Form.ColumnForm.View()

	var title string
	if m.UIState.Mode == state.AddColumnFormMode {
		title = "New Column"
	} else {
		title = "Rename Column"
	}

	formBox := components.CreateInputBoxStyle.
		Width(50).
		Render(title + "\n\n" + formView)

	return layers.CreateCenteredLayer(formBox, m.UIState.Width(), m.UIState.Height)
}

// renderHelpLayer renders the keyboard shortcuts help screen as a layer
func (m Model) renderHelpLayer() *lipgloss.Layer {
	helpBox := components.HelpBoxStyle.
		Width(50).
		Render(m.generateHelpText())

	return layers.CreateCenteredLayer(helpBox, m.UIState.Width(), m.UIState.Height)
}

// renderDiscardConfirmLayer renders the discard confirmation dialog as a layer
func (m Model) renderDiscardConfirmLayer() *lipgloss.Layer {
	ctx := m.UIState.DiscardContext
	if ctx == nil {
		return nil
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("%s\n\n[y]es  [n]o", ctx.Message))

	return layers.CreateCenteredLayer(confirmBox, m.UIState.Width(), m.UIState.Height)
}

func (m Model) renderProjectBranchConfirmLayer() *lipgloss.Layer {
	ctx := m.UIState.ProjectBranchContext
	if ctx == nil {
		return nil
	}

	var message string
	if ctx.ExistingProject != nil {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		message = warningStyle.Render(fmt.Sprintf(
			"▲  Branch '%s' is currently linked to project '%s'.\nTransfer it to '%s'?",
			ctx.GitBranch,
			ctx.ExistingProject.Name,
			ctx.ProjectName,
		))
	} else {
		message = fmt.Sprintf("Associate branch '%s' with project '%s'?", ctx.GitBranch, ctx.ProjectName)
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(60).
		Render(fmt.Sprintf("%s\n\n[y]es transfer  [n]o skip branch  [esc] cancel", message))

	return layers.CreateCenteredLayer(confirmBox, m.UIState.Width(), m.UIState.Height)
}

// renderDeleteTaskConfirmLayer renders the task deletion confirmation dialog as a layer
func (m Model) renderDeleteTaskConfirmLayer() *lipgloss.Layer {
	task := m.getCurrentTask()
	if task == nil {
		return nil
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))

	return layers.CreateCenteredLayer(confirmBox, m.UIState.Width(), m.UIState.Height)
}

// renderDeleteColumnConfirmLayer renders the column deletion confirmation dialog as a layer
func (m Model) renderDeleteColumnConfirmLayer() *lipgloss.Layer {
	column := m.getCurrentColumn()
	if column == nil {
		return nil
	}

	var content string
	taskCount := m.Forms.Input.DeleteColumnTaskCount
	if taskCount > 0 {
		content = fmt.Sprintf(
			"Delete column '%s'?\nThis will also delete %d task(s).\n\n[y]es  [n]o",
			column.Name,
			taskCount,
		)
	} else {
		content = fmt.Sprintf("Delete column '%s'?\n\n[y]es  [n]o", column.Name)
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(content)

	return layers.CreateCenteredLayer(confirmBox, m.UIState.Width(), m.UIState.Height)
}

// renderDeleteProjectConfirmLayer renders the project deletion confirmation dialog as a layer
func (m Model) renderDeleteProjectConfirmLayer() *lipgloss.Layer {
	project := m.AppState.GetCurrentProject()
	if project == nil {
		return nil
	}

	var content string
	taskCount := m.Forms.Input.DeleteProjectTaskCount
	if taskCount > 0 {
		content = fmt.Sprintf(
			"Delete project '%s'?\nThis will also delete %d task(s).\n\n[y]es  [f]orce delete  [n]o",
			project.Name,
			taskCount,
		)
	} else {
		content = fmt.Sprintf("Delete project '%s'?\n\n[y]es  [n]o", project.Name)
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(50).
		Render(content)

	return layers.CreateCenteredLayer(confirmBox, m.UIState.Width(), m.UIState.Height)
}

// renderDeleteLabelConfirmLayer renders the label deletion confirmation dialog as a layer
func (m Model) renderDeleteLabelConfirmLayer() *lipgloss.Layer {
	labelName := m.Pickers.Label.DeleteLabelName
	if labelName == "" {
		return nil
	}

	var content string
	taskCount := m.Pickers.Label.DeleteLabelTaskCount
	if taskCount > 0 {
		content = fmt.Sprintf(
			"There are %d task(s) with this label.\n\nAre you sure you want to delete '%s'?\nAssociations will be removed.\n\n[y]es  [n]o",
			taskCount,
			labelName,
		)
	} else {
		content = fmt.Sprintf("Delete label '%s'?\n\n[y]es  [n]o", labelName)
	}

	confirmBox := components.DeleteConfirmBoxStyle.
		Width(55).
		Render(content)

	return layers.CreateCenteredLayer(confirmBox, m.UIState.Width(), m.UIState.Height)
}

// generateHelpText creates help text based on current key mappings
func (m Model) generateHelpText() string {
	km := m.Config.KeyMappings
	return fmt.Sprintf(`PASO - Keyboard Shortcuts

TASKS
  %s     Add new task
  %s     Edit selected task
  %s     Delete selected task
  %s     Edit task details
  %s     Edit labels
  %s     Edit priority
  %s     Edit assignee
  %s     Edit estimate
  %s     Edit due date
  %s     Edit type
  %s     Edit parent task
  %s     Edit child task
  %s     Move task to another project

KANBAN
  %s     Create new column (after current)
  %s     Rename current column
  %s     Delete current column
  %s     Move task to previous column
  %s     Move task to next column
  %s     Move task up in column
  %s     Move task down in column

NAVIGATION
  %s     Move left
  %s     Move right
  %s     Move up
  %s     Move down
  %s     Scroll viewport left
  %s     Scroll viewport right

PROJECTS
  %s     Switch to previous project
  %s     Switch to next project
  %s     Create new project
  %s     Edit current project
  %s     Delete current project

VIEW
  %s     Toggle between kanban and list view
  %s     Search tasks
  %s     Open filter bar

FORMS
  %s     Save form
  %s     Open comments view
  %s     Refresh git data
  %s     Edit labels
  %s     Edit parent task
  %s     Edit child task
  %s     Edit priority
  %s     Edit type
  %s     Edit assignee
  %s     Edit estimate
  %s     Edit due date
  %s     Show form help

OTHER
  %s     Show this help
  %s     Connect to remote database
  %s     Suspend (resume with fg)
  %s     Quit

Press any key to close`,
		km.Tasks.AddTask,
		km.Tasks.EditTask,
		km.Tasks.DeleteTask,
		km.Tasks.ViewTask,
		km.Tasks.EditLabels,
		km.Tasks.EditPriority,
		km.Tasks.EditAssignee,
		km.Tasks.EditEstimate,
		km.Tasks.EditDueDate,
		km.Tasks.EditType,
		km.Tasks.EditParentTask,
		km.Tasks.EditChildTask,
		km.Tasks.MoveTaskToProject,
		km.Kanban.CreateColumn,
		km.Kanban.RenameColumn,
		km.Kanban.DeleteColumn,
		km.Kanban.MoveTaskLeft,
		km.Kanban.MoveTaskRight,
		km.Kanban.MoveTaskUp,
		km.Kanban.MoveTaskDown,
		km.Navigation.MoveLeft,
		km.Navigation.MoveRight,
		km.Navigation.MoveUp,
		km.Navigation.MoveDown,
		km.Navigation.ScrollViewportLeft,
		km.Navigation.ScrollViewportRight,
		km.Navigation.PrevProject,
		km.Navigation.NextProject,
		km.Projects.CreateProject,
		km.Projects.EditProject,
		km.Projects.DeleteProject,
		km.General.ToggleView,
		km.General.Search,
		km.General.FilterBar,
		km.Forms.SaveForm,
		km.Forms.OpenCommentsView,
		km.Forms.RefreshGitData,
		km.Forms.EditLabels,
		km.Forms.EditParentTask,
		km.Forms.EditChildTask,
		km.Forms.EditPriority,
		km.Forms.EditType,
		km.Forms.EditAssignee,
		km.Forms.EditEstimate,
		km.Forms.EditDueDate,
		km.Forms.ShowHelp,
		km.General.ShowHelp,
		km.General.ConnectRemote,
		km.General.Suspend,
		km.General.Quit,
	)
}

// renderCommentsViewLayer renders the comments view modal as a full-screen layer
func (m Model) renderCommentsViewLayer() *lipgloss.Layer {
	layerWidth := m.UIState.Width() * 8 / 10
	layerHeight := m.UIState.Height * 8 / 10

	content := m.renderCommentsViewContent(layerWidth, layerHeight)

	commentsBox := components.HelpBoxStyle.
		Width(layerWidth).
		Height(layerHeight).
		Render(content)

	return layers.CreateCenteredLayer(commentsBox, m.UIState.Width(), m.UIState.Height)
}

// renderCommentFormLayer renders the comment creation/edit form modal as a layer
func (m Model) renderCommentFormLayer() *lipgloss.Layer {
	if m.Forms.Form.CommentForm == nil {
		return nil
	}

	formView := m.Forms.Form.CommentForm.View()

	var title string
	if m.Forms.Form.EditingCommentID == 0 {
		title = "New Comment"
	} else {
		title = "Edit Comment"
	}

	formBox := components.CreateInputBoxStyle.
		Width(m.UIState.Width() * 3 / 4).
		Height(m.UIState.Height * 2 / 3).
		Render(title + "\n\n" + formView)

	return layers.CreateCenteredLayer(formBox, m.UIState.Width(), m.UIState.Height)
}

// renderTaskFormHelpLayer renders the task form keyboard shortcuts help screen as a layer
func (m Model) renderTaskFormHelpLayer() *lipgloss.Layer {
	km := m.Config.KeyMappings
	helpContent := fmt.Sprintf(`TASK FORM - Keyboard Shortcuts

FORM NAVIGATION
  Tab             Navigate between form fields
  Shift+Tab       Navigate backwards
  %-15s Save task and close form
  Esc             Close form (will prompt if unsaved)

TEXT EDITING (Title/Description)
  Shift+Enter     New line
  Alt+Enter       New line
  Ctrl+J          New line
  Enter           Next field

COMMENTS SECTION
  Ctrl+↓          Focus comments section
  Down            Auto-focus comments (when not focused)
  ↑↓              Scroll comments (when focused)
  Mouse wheel     Scroll comments (when focused)
  Tab/Shift+Tab   Return to form fields

QUICK ACTIONS
  %-15s Create new comment
  %-15s Manage labels
  %-15s Select parent tasks
  %-15s Select child tasks
  %-15s Change priority
  %-15s Change task type
  %-15s Change assignee
  %-15s Change estimate
  %-15s Set due date

HELP
  %-15s Toggle this help menu
  Esc             Close help menu

Press %s or Esc to close`,
		km.Forms.SaveForm,
		km.Forms.OpenCommentsView,
		km.Forms.EditLabels,
		km.Forms.EditParentTask,
		km.Forms.EditChildTask,
		km.Forms.EditPriority,
		km.Forms.EditType,
		km.Forms.EditAssignee,
		km.Forms.EditEstimate,
		km.Forms.EditDueDate,
		km.Forms.ShowHelp,
		km.Forms.ShowHelp,
	)

	helpBox := components.HelpBoxStyle.
		Width(m.UIState.Width() * 3 / 8).
		Render(helpContent)

	return layers.CreateCenteredLayer(helpBox, m.UIState.Width(), m.UIState.Height)
}

type pickerDimensionStrategy interface {
	Calculate(screenWidth, screenHeight int) (width, height int)
}

// dynamicPickerDimensions calculates dimensions based on item count, filter presence, and screen size.
// Suitable for pickers with variable content that should adapt to available space.
type dynamicPickerDimensions struct {
	itemCount int
	hasFilter bool
	minWidth  int
	maxWidth  int
}

func (d dynamicPickerDimensions) Calculate(screenWidth, screenHeight int) (width, height int) {
	return layers.CalculatePickerDimensions(
		d.itemCount,
		d.hasFilter,
		screenWidth,
		screenHeight,
		d.minWidth,
		d.maxWidth,
	)
}

// fixedPickerDimensions uses predetermined width and height values.
// Suitable for pickers with small, fixed item counts where consistent sizing is preferred.
type fixedPickerDimensions struct {
	width  int
	height int
}

func (f fixedPickerDimensions) Calculate(screenWidth, screenHeight int) (width, height int) {
	return f.width, f.height
}

// statusPickerDimensions calculates dimensions with fixed width but dynamic height based on item count.
// Specifically designed for the status/column picker where width is constant but height varies with columns.
type statusPickerDimensions struct {
	itemCount int
}

func (s statusPickerDimensions) Calculate(screenWidth, screenHeight int) (width, height int) {
	return layers.PickerStatusWidth, s.itemCount + layers.PickerStatusChromeHeight
}

// pickerLayerConfig holds configuration for creating a picker layer
type pickerLayerConfig struct {
	dimensionStrategy pickerDimensionStrategy
	contentRenderer   func(width, height int) string
	boxStyle          lipgloss.Style
}

// createPickerLayer creates a centered picker layer using the provided configuration
func (m Model) createPickerLayer(config pickerLayerConfig) *lipgloss.Layer {
	pickerWidth, pickerHeight := config.dimensionStrategy.Calculate(
		m.UIState.Width(),
		m.UIState.Height,
	)

	pickerContent := config.contentRenderer(pickerWidth, pickerHeight)

	pickerBox := config.boxStyle.
		Width(pickerWidth).
		Height(pickerHeight).
		Render(pickerContent)

	return layers.CreateCenteredLayer(pickerBox, m.UIState.Width(), m.UIState.Height)
}

// renderLabelPickerLayer renders the label picker modal as a layer
func (m Model) renderLabelPickerLayer() *lipgloss.Layer {
	// Name input step (first step of label creation)
	if m.Pickers.Label.NameInputMode {
		return m.createPickerLayer(pickerLayerConfig{
			dimensionStrategy: dynamicPickerDimensions{
				itemCount: 3, // Small dialog for name input
				hasFilter: false,
				minWidth:  layers.PickerDefaultMinWidth,
				maxWidth:  layers.PickerDefaultMaxWidth,
			},
			contentRenderer: func(width, height int) string {
				return renderers.RenderLabelNameInput(
					m.Pickers.Label.NameBuffer,
					width-layers.PickerBorderPaddingWidth,
				)
			},
			boxStyle: components.LabelPickerCreateBoxStyle,
		})
	}

	// Color picker step (second step of label creation)
	if m.Pickers.Label.CreateMode {
		return m.createPickerLayer(pickerLayerConfig{
			dimensionStrategy: dynamicPickerDimensions{
				itemCount: layers.PickerColorDefaultItemCount,
				hasFilter: false,
				minWidth:  layers.PickerDefaultMinWidth,
				maxWidth:  layers.PickerDefaultMaxWidth,
			},
			contentRenderer: func(width, height int) string {
				return renderers.RenderLabelColorPicker(
					renderers.GetDefaultLabelColors(),
					m.Pickers.Label.ColorIdx,
					m.Forms.Form.FormLabelName,
					width-layers.PickerBorderPaddingWidth,
				)
			},
			boxStyle: components.LabelPickerCreateBoxStyle,
		})
	}

	filteredItems := m.getFilteredLabelPickerItems()
	hasFilter := m.Pickers.Label.Filter != ""

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: dynamicPickerDimensions{
			itemCount: len(filteredItems) + 1,
			hasFilter: hasFilter,
			minWidth:  layers.PickerDefaultMinWidth,
			maxWidth:  layers.PickerDefaultMaxWidth,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderLabelPicker(
				filteredItems,
				m.Pickers.Label.Cursor,
				m.Pickers.Label.Filter,
				true,
				width-layers.PickerBorderPaddingWidth,
				height-layers.PickerBorderPaddingHeight,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderParentPickerLayer renders the parent task picker modal as a layer
func (m Model) renderParentPickerLayer() *lipgloss.Layer {
	filteredItems := m.Pickers.Parent.GetFilteredItems()
	hasFilter := m.Pickers.Parent.Filter != ""
	isParentPicker := true

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: dynamicPickerDimensions{
			itemCount: len(filteredItems),
			hasFilter: hasFilter,
			minWidth:  layers.PickerTaskMinWidth,
			maxWidth:  layers.PickerTaskMaxWidth,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderTaskPicker(
				filteredItems,
				m.Pickers.Parent.Cursor,
				m.Pickers.Parent.Filter,
				"Parent Issues",
				width-layers.PickerBorderPaddingWidth,
				height-layers.PickerBorderPaddingHeight,
				isParentPicker,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderChildPickerLayer renders the child task picker modal as a layer
func (m Model) renderChildPickerLayer() *lipgloss.Layer {
	filteredItems := m.Pickers.Child.GetFilteredItems()
	hasFilter := m.Pickers.Child.Filter != ""
	isParentPicker := false

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: dynamicPickerDimensions{
			itemCount: len(filteredItems),
			hasFilter: hasFilter,
			minWidth:  layers.PickerTaskMinWidth,
			maxWidth:  layers.PickerTaskMaxWidth,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderTaskPicker(
				filteredItems,
				m.Pickers.Child.Cursor,
				m.Pickers.Child.Filter,
				"Child Issues",
				width-layers.PickerBorderPaddingWidth,
				height-layers.PickerBorderPaddingHeight,
				isParentPicker,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderPriorityPickerLayer renders the priority picker modal as a layer
func (m Model) renderPriorityPickerLayer() *lipgloss.Layer {
	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  layers.PickerPriorityWidth,
			height: layers.PickerPriorityHeight,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderPriorityPicker(
				renderers.GetPriorityOptions(),
				m.Pickers.Priority.SelectedPriorityID(),
				m.Pickers.Priority.Cursor(),
				width-layers.PickerBorderPaddingWidth,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderAssigneePickerLayer renders the assignee picker modal as a layer
func (m Model) renderAssigneePickerLayer() *lipgloss.Layer {
	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  layers.PickerAssigneeWidth,
			height: layers.PickerAssigneeHeight,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderAssigneePicker(
				m.Pickers.Assignee.Assignees(),
				m.Pickers.Assignee.SelectedID(),
				m.Pickers.Assignee.Cursor(),
				true, // showClearOpt
				width-layers.PickerBorderPaddingWidth,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderProjectPickerLayer renders the project picker modal as a layer
func (m Model) renderProjectPickerLayer() *lipgloss.Layer {
	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  layers.PickerProjectWidth,
			height: layers.PickerProjectHeight,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderProjectPicker(
				m.Pickers.Project.Projects(),
				m.Pickers.Project.Cursor(),
				width-layers.PickerBorderPaddingWidth,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderEstimateInputLayer renders the estimate input overlay as a layer
func (m Model) renderEstimateInputLayer() *lipgloss.Layer {
	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  layers.EstimateInputWidth,
			height: layers.EstimateInputHeight,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderEstimateInput(
				m.Pickers.Estimate.Buffer(),
				m.Pickers.Estimate.Error(),
				width-layers.PickerBorderPaddingWidth,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderDatePickerLayer renders the date picker as a centered modal overlay
func (m Model) renderDatePickerLayer() *lipgloss.Layer {
	// Calculate dynamic height based on current month
	contentHeight := renderers.DatePickerContentHeight(m.Pickers.DatePicker.CurrentMonth, m.Pickers.DatePicker.CurrentYear)
	totalHeight := contentHeight + layers.PickerBorderPaddingHeight

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  layers.DatePickerWidth,
			height: totalHeight,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderDatePicker(
				m.Pickers.DatePicker,
				width-layers.PickerBorderPaddingWidth,
				height-layers.PickerBorderPaddingHeight,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderTypePickerLayer renders the type picker modal as a layer
func (m Model) renderTypePickerLayer() *lipgloss.Layer {
	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  layers.PickerTypeWidth,
			height: layers.PickerTypeHeight,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderTypePicker(
				renderers.GetTypeOptions(),
				m.Pickers.Type.SelectedTypeID(),
				m.Pickers.Type.Cursor(),
				width-layers.PickerBorderPaddingWidth,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderRelationTypePickerLayer renders the relation type picker modal as a layer
func (m Model) renderRelationTypePickerLayer() *lipgloss.Layer {
	pickerType := "parent"
	if m.Pickers.RelationType.ReturnMode == state.ChildPickerMode {
		pickerType = "child"
	}

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: fixedPickerDimensions{
			width:  layers.PickerRelationTypeWidth,
			height: layers.PickerRelationTypeHeight,
		},
		contentRenderer: func(width, height int) string {
			return renderers.RenderRelationTypePicker(
				renderers.GetRelationTypeOptions(),
				m.Pickers.RelationType.SelectedRelationTypeID(),
				m.Pickers.RelationType.Cursor(),
				width-layers.PickerBorderPaddingWidth,
				pickerType,
			)
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderStatusPickerLayer renders the status/column selection picker modal as a layer
func (m Model) renderStatusPickerLayer() *lipgloss.Layer {
	columns := m.Pickers.Status.Columns()
	cursor := m.Pickers.Status.Cursor()

	return m.createPickerLayer(pickerLayerConfig{
		dimensionStrategy: statusPickerDimensions{
			itemCount: len(columns),
		},
		contentRenderer: func(width, height int) string {
			var items []string
			for i, col := range columns {
				prefix := "  "
				if i == cursor {
					prefix = "> "
				}
				items = append(items, prefix+col.Name)
			}
			return "Select Status:\n\n" + lipgloss.JoinVertical(lipgloss.Left, items...) + "\n\n" + components.PickerFooterConfirm
		},
		boxStyle: components.LabelPickerBoxStyle,
	})
}

// renderConnectingSpinnerLayer renders the connecting spinner as a centered overlay
func (m Model) renderConnectingSpinnerLayer() *lipgloss.Layer {
	if !m.DatabasePicker.IsConnecting() {
		return nil
	}

	content := components.RenderConnectingSpinner(
		m.DatabasePicker.ConnectingDBName,
		m.DatabasePicker.SpinnerFrame,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Highlight)).
		Padding(1, 2)

	styledContent := boxStyle.Render(content)

	return layers.CreateCenteredLayer(styledContent, m.UIState.Width(), m.UIState.Height)
}
