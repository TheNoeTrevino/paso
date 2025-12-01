package tui

import (
	"context"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// NORMAL MODE HANDLERS
// ============================================================================

// handleNormalMode dispatches key events in NormalMode to specific handlers.
func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorState.Clear()

	switch msg.String() {
	case "q", "ctrl+c":
		return m.handleQuit()
	case "?":
		return m.handleShowHelp()
	case "a":
		return m.handleAddTask()
	case "e":
		return m.handleEditTask()
	case "d":
		return m.handleDeleteTask()
	case " ":
		return m.handleViewTask()
	case "C":
		return m.handleCreateColumn()
	case "R":
		return m.handleRenameColumn()
	case "X":
		return m.handleDeleteColumn()
	case "]":
		return m.handleScrollRight()
	case "[":
		return m.handleScrollLeft()
	case "h", "left":
		return m.handleNavigateLeft()
	case "l", "right":
		return m.handleNavigateRight()
	case "j", "down":
		return m.handleNavigateDown()
	case "k", "up":
		return m.handleNavigateUp()
	case ">", "L":
		return m.handleMoveTaskRight()
	case "<", "H":
		return m.handleMoveTaskLeft()
	case "{":
		return m.handlePrevProject()
	case "}":
		return m.handleNextProject()
	case "ctrl+p":
		return m.handleCreateProject()
	}

	return m, nil
}

// handleQuit exits the application.
func (m Model) handleQuit() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

// handleShowHelp shows the help screen.
func (m Model) handleShowHelp() (tea.Model, tea.Cmd) {
	m.uiState.SetMode(state.HelpMode)
	return m, nil
}

// handleNavigateLeft moves selection to the previous column.
func (m Model) handleNavigateLeft() (tea.Model, tea.Cmd) {
	if m.uiState.SelectedColumn() > 0 {
		m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() - 1)
		m.uiState.SetSelectedTask(0)
		m.uiState.EnsureSelectionVisible(m.uiState.SelectedColumn())
	}
	return m, nil
}

// handleNavigateRight moves selection to the next column.
func (m Model) handleNavigateRight() (tea.Model, tea.Cmd) {
	if m.uiState.SelectedColumn() < len(m.appState.Columns())-1 {
		m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
		m.uiState.SetSelectedTask(0)
		m.uiState.EnsureSelectionVisible(m.uiState.SelectedColumn())
	}
	return m, nil
}

// handleNavigateUp moves selection to the previous task.
func (m Model) handleNavigateUp() (tea.Model, tea.Cmd) {
	if m.uiState.SelectedTask() > 0 {
		m.uiState.SetSelectedTask(m.uiState.SelectedTask() - 1)
	}
	return m, nil
}

// handleNavigateDown moves selection to the next task.
func (m Model) handleNavigateDown() (tea.Model, tea.Cmd) {
	currentTasks := m.getCurrentTasks()
	if len(currentTasks) > 0 && m.uiState.SelectedTask() < len(currentTasks)-1 {
		m.uiState.SetSelectedTask(m.uiState.SelectedTask() + 1)
	}
	return m, nil
}

// handleScrollRight scrolls the viewport right.
func (m Model) handleScrollRight() (tea.Model, tea.Cmd) {
	if m.uiState.ViewportOffset()+m.uiState.ViewportSize() < len(m.appState.Columns()) {
		m.uiState.SetViewportOffset(m.uiState.ViewportOffset() + 1)
		if m.uiState.SelectedColumn() < m.uiState.ViewportOffset() {
			m.uiState.SetSelectedColumn(m.uiState.ViewportOffset())
			m.uiState.SetSelectedTask(0)
		}
	}
	return m, nil
}

// handleScrollLeft scrolls the viewport left.
func (m Model) handleScrollLeft() (tea.Model, tea.Cmd) {
	if m.uiState.ViewportOffset() > 0 {
		m.uiState.SetViewportOffset(m.uiState.ViewportOffset() - 1)
		if m.uiState.SelectedColumn() >= m.uiState.ViewportOffset()+m.uiState.ViewportSize() {
			m.uiState.SetSelectedColumn(m.uiState.ViewportOffset() + m.uiState.ViewportSize() - 1)
			m.uiState.SetSelectedTask(0)
		}
	}
	return m, nil
}

// handleAddTask initiates adding a new task.
func (m Model) handleAddTask() (tea.Model, tea.Cmd) {
	if len(m.appState.Columns()) == 0 {
		m.errorState.Set("Cannot add task: No columns exist. Create a column first with 'C'")
		return m, nil
	}
	m.formState.FormTitle = ""
	m.formState.FormDescription = ""
	m.formState.FormLabelIDs = []int{}
	m.formState.FormConfirm = true
	m.formState.EditingTaskID = 0
	m.formState.TicketForm = CreateTicketForm(
		&m.formState.FormTitle,
		&m.formState.FormDescription,
		&m.formState.FormLabelIDs,
		m.appState.Labels(),
		&m.formState.FormConfirm,
	)
	m.uiState.SetMode(state.TicketFormMode)
	return m, m.formState.TicketForm.Init()
}

// handleEditTask initiates editing the selected task.
func (m Model) handleEditTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task == nil {
		m.errorState.Set("No task selected to edit")
		return m, nil
	}

	taskDetail, err := m.repo.GetTaskDetail(context.Background(), task.ID)
	if err != nil {
		log.Printf("Error loading task details: %v", err)
		m.errorState.Set("Error loading task details")
		return m, nil
	}

	m.formState.FormTitle = taskDetail.Title
	m.formState.FormDescription = taskDetail.Description
	m.formState.FormLabelIDs = make([]int, len(taskDetail.Labels))
	for i, label := range taskDetail.Labels {
		m.formState.FormLabelIDs[i] = label.ID
	}
	m.formState.FormConfirm = true
	m.formState.EditingTaskID = task.ID
	m.formState.TicketForm = CreateTicketForm(
		&m.formState.FormTitle,
		&m.formState.FormDescription,
		&m.formState.FormLabelIDs,
		m.appState.Labels(),
		&m.formState.FormConfirm,
	)
	m.uiState.SetMode(state.TicketFormMode)
	return m, m.formState.TicketForm.Init()
}

// handleDeleteTask initiates task deletion confirmation.
func (m Model) handleDeleteTask() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() == nil {
		m.errorState.Set("No task selected to delete")
		return m, nil
	}
	m.uiState.SetMode(state.DeleteConfirmMode)
	return m, nil
}

// handleViewTask shows the task detail popup.
func (m Model) handleViewTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task == nil {
		m.errorState.Set("No task selected to view")
		return m, nil
	}

	taskDetail, err := m.repo.GetTaskDetail(context.Background(), task.ID)
	if err != nil {
		log.Printf("Error loading task details: %v", err)
		m.errorState.Set("Error loading task details")
		return m, nil
	}
	m.uiState.SetViewingTask(taskDetail)
	m.uiState.SetMode(state.ViewTaskMode)
	return m, nil
}

// handleMoveTaskRight moves the task to the next column.
func (m Model) handleMoveTaskRight() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskRight()
	}
	return m, nil
}

// handleMoveTaskLeft moves the task to the previous column.
func (m Model) handleMoveTaskLeft() (tea.Model, tea.Cmd) {
	if m.getCurrentTask() != nil {
		m.moveTaskLeft()
	}
	return m, nil
}

// handleCreateColumn initiates column creation.
func (m Model) handleCreateColumn() (tea.Model, tea.Cmd) {
	m.uiState.SetMode(state.AddColumnMode)
	m.inputState.Prompt = "New column name:"
	m.inputState.Buffer = ""
	return m, nil
}

// handleRenameColumn initiates column renaming.
func (m Model) handleRenameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.errorState.Set("No column selected to rename")
		return m, nil
	}
	m.uiState.SetMode(state.EditColumnMode)
	m.inputState.Buffer = column.Name
	m.inputState.Prompt = "Rename column:"
	return m, nil
}

// handleDeleteColumn initiates column deletion confirmation.
func (m Model) handleDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column == nil {
		m.errorState.Set("No column selected to delete")
		return m, nil
	}
	taskCount, err := m.repo.GetTaskCountByColumn(context.Background(), column.ID)
	if err != nil {
		log.Printf("Error getting task count: %v", err)
		m.errorState.Set("Error getting column info")
		return m, nil
	}
	m.inputState.DeleteColumnTaskCount = taskCount
	m.uiState.SetMode(state.DeleteColumnConfirmMode)
	return m, nil
}

// handlePrevProject switches to the previous project.
func (m Model) handlePrevProject() (tea.Model, tea.Cmd) {
	if m.appState.SelectedProject() > 0 {
		m.switchToProject(m.appState.SelectedProject() - 1)
	}
	return m, nil
}

// handleNextProject switches to the next project.
func (m Model) handleNextProject() (tea.Model, tea.Cmd) {
	if m.appState.SelectedProject() < len(m.appState.Projects())-1 {
		m.switchToProject(m.appState.SelectedProject() + 1)
	}
	return m, nil
}

// handleCreateProject initiates project creation.
func (m Model) handleCreateProject() (tea.Model, tea.Cmd) {
	m.formState.FormProjectName = ""
	m.formState.FormProjectDescription = ""
	m.formState.ProjectForm = CreateProjectForm(
		&m.formState.FormProjectName,
		&m.formState.FormProjectDescription,
	)
	m.uiState.SetMode(state.ProjectFormMode)
	return m, m.formState.ProjectForm.Init()
}

// ============================================================================
// INPUT MODE HANDLERS
// ============================================================================

// handleInputMode handles text input for column creation/editing.
func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleInputConfirm()
	case "esc":
		return m.handleInputCancel()
	case "backspace", "ctrl+h":
		m.inputState.Backspace()
		return m, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			m.inputState.AppendChar(rune(key[0]))
		}
		return m, nil
	}
}

// handleInputConfirm processes the input and creates/renames column.
func (m Model) handleInputConfirm() (tea.Model, tea.Cmd) {
	if strings.TrimSpace(m.inputState.Buffer) == "" {
		m.uiState.SetMode(state.NormalMode)
		m.inputState.Clear()
		return m, nil
	}

	if m.uiState.Mode() == state.AddColumnMode {
		return m.createColumn()
	}
	return m.renameColumn()
}

// handleInputCancel cancels the input operation.
func (m Model) handleInputCancel() (tea.Model, tea.Cmd) {
	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// createColumn creates a new column with the input buffer as name.
func (m Model) createColumn() (tea.Model, tea.Cmd) {
	var afterColumnID *int
	if len(m.appState.Columns()) > 0 {
		currentCol := m.appState.Columns()[m.uiState.SelectedColumn()]
		afterColumnID = &currentCol.ID
	}

	projectID := 0
	if project := m.getCurrentProject(); project != nil {
		projectID = project.ID
	}

	column, err := m.repo.CreateColumn(context.Background(), strings.TrimSpace(m.inputState.Buffer), projectID, afterColumnID)
	if err != nil {
		log.Printf("Error creating column: %v", err)
		m.errorState.Set("Failed to create column")
	} else {
		columns, err := m.repo.GetColumnsByProject(context.Background(), projectID)
		if err != nil {
			log.Printf("Error reloading columns: %v", err)
			m.errorState.Set("Failed to reload columns")
		}
		m.appState.SetColumns(columns)
		m.appState.Tasks()[column.ID] = []*models.TaskSummary{}
		if afterColumnID != nil {
			m.uiState.SetSelectedColumn(m.uiState.SelectedColumn() + 1)
		}
	}

	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// renameColumn renames the current column with the input buffer.
func (m Model) renameColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		err := m.repo.UpdateColumnName(context.Background(), column.ID, strings.TrimSpace(m.inputState.Buffer))
		if err != nil {
			log.Printf("Error updating column: %v", err)
			m.errorState.Set("Failed to rename column")
		} else {
			column.Name = strings.TrimSpace(m.inputState.Buffer)
		}
	}

	m.uiState.SetMode(state.NormalMode)
	m.inputState.Clear()
	return m, nil
}

// ============================================================================
// CONFIRMATION HANDLERS
// ============================================================================

// handleDeleteConfirm handles task deletion confirmation.
func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteTask()
	case "n", "N", "esc":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteTask performs the actual task deletion.
func (m Model) confirmDeleteTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task != nil {
		err := m.repo.DeleteTask(context.Background(), task.ID)
		if err != nil {
			log.Printf("Error deleting task: %v", err)
			m.errorState.Set("Failed to delete task")
		} else {
			m.removeCurrentTask()
		}
	}
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}

// handleDeleteColumnConfirm handles column deletion confirmation.
func (m Model) handleDeleteColumnConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteColumn()
	case "n", "N", "esc":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteColumn performs the actual column deletion.
func (m Model) confirmDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		err := m.repo.DeleteColumn(context.Background(), column.ID)
		if err != nil {
			log.Printf("Error deleting column: %v", err)
			m.errorState.Set("Failed to delete column")
		} else {
			delete(m.appState.Tasks(), column.ID)
			m.removeCurrentColumn()
		}
	}
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}

// ============================================================================
// VIEW/HELP MODE HANDLERS
// ============================================================================

// handleHelpMode handles input in the help screen.
func (m Model) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "q", "esc", "enter", " ":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// handleViewTaskMode handles input when viewing task details.
func (m Model) handleViewTaskMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", " ", "q":
		m.uiState.SetMode(state.NormalMode)
		m.uiState.SetViewingTask(nil)
		return m, nil
	case "l":
		if m.uiState.ViewingTask() != nil {
			if m.initLabelPicker(m.uiState.ViewingTask().ID) {
				m.uiState.SetMode(state.LabelPickerMode)
			}
		}
		return m, nil
	}
	return m, nil
}
