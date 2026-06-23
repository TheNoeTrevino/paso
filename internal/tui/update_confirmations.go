package tui

import (
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/project"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// handleDeleteConfirm handles task or comment deletion confirmation.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Check if we're deleting a comment or task
		if m.Forms.Form.EditingCommentID != 0 {
			return m.confirmDeleteComment()
		}
		return m.confirmDeleteTask()
	case "n", "N", "esc":
		// Return to appropriate mode
		if m.Forms.Form.EditingCommentID != 0 {
			m.Forms.Form.EditingCommentID = 0
			m.UIState.Mode = state.CommentsViewMode
		} else {
			m.UIState.Mode = state.NormalMode
		}
		return m, nil
	}
	return m, nil
}

// confirmDeleteTask performs the actual task deletion.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDeleteTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task != nil {
		ctx, cancel := m.DBContext()
		defer cancel()
		err := m.App.TaskService.DeleteTask(ctx, task.ID)
		if err != nil {
			slog.Error("failed to delete task", "error", err)
			m.UIState.Mode = state.NormalMode
			return m, m.addNotification(state.LevelError, "Failed to delete task")
		} else {
			m.removeCurrentTask()
			// Invalidate the deleted task from detail cache
			if m.DetailCache != nil {
				m.DetailCache.Invalidate(task.ID)
			}
		}
	}
	m.UIState.Mode = state.NormalMode
	return m, nil
}

// confirmDeleteComment performs the actual comment deletion.
func (m Model) confirmDeleteComment() (tea.Model, tea.Cmd) {
	commentID := m.Forms.Form.EditingCommentID
	taskID := m.Forms.Comment.TaskID

	if commentID != 0 && taskID != 0 {
		ctx, cancel := m.DBContext()
		defer cancel()
		err := m.App.TaskService.DeleteComment(ctx, commentID)
		if err != nil {
			slog.Error("failed to delete comment", "error", err)
			m.Forms.Form.EditingCommentID = 0
			m.UIState.Mode = state.CommentsViewMode
			return m, m.addNotification(state.LevelError, "Failed to delete comment")
		}
		// Remove from local state
		m.Forms.Comment.DeleteSelected()
		var cmds []tea.Cmd
		cmds = append(cmds, m.addNotification(state.LevelInfo, "Comment deleted"))

		// Refresh comments in form state
		comments, err := m.App.TaskService.GetCommentsByTask(ctx, taskID)
		if err == nil {
			m.Forms.Form.FormComments = comments
		}
		// Refresh activities for comments view and task form preview
		activities, actErr := m.App.TaskService.GetTaskActivities(ctx, taskID)
		if actErr != nil {
			cmds = append(cmds, m.addNotification(state.LevelError, "Failed to refresh activity"))
		} else {
			m.Forms.Comment.SetActivities(activities)
			m.Forms.Form.FormActivities = activities
		}

		m.Forms.Form.EditingCommentID = 0
		m.UIState.Mode = state.CommentsViewMode
		return m, tea.Batch(cmds...)
	}

	m.Forms.Form.EditingCommentID = 0
	m.UIState.Mode = state.CommentsViewMode
	return m, nil
}

// handleDiscardConfirm handles discard confirmation for forms and inputs.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDiscardConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ctx := m.UIState.DiscardContext
	if ctx == nil {
		// Safety: if context is missing, return to normal mode
		m.UIState.Mode = state.NormalMode
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed discard - clear form/input and return to normal mode
		return m.confirmDiscard()

	case "n", "N", "esc":
		// User cancelled - return to source mode without clearing
		m.UIState.Mode = ctx.SourceMode
		m.UIState.DiscardContext = nil
		return m, nil
	}

	return m, nil
}

// confirmDiscard performs the actual discard operation based on context.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDiscard() (tea.Model, tea.Cmd) {
	ctx := m.UIState.DiscardContext
	if ctx == nil {
		m.UIState.Mode = state.NormalMode
		return m, nil
	}

	// Clear the appropriate form/input based on source mode
	switch ctx.SourceMode {
	case state.TaskFormMode:
		m.Forms.Form.ClearTaskForm()

	case state.ProjectFormMode, state.EditProjectFormMode:
		m.Forms.Form.ClearProjectForm()

	case state.AddColumnFormMode, state.EditColumnFormMode:
		m.Forms.Form.ClearColumnForm()

	case state.CommentFormMode:
		m.Forms.Form.ClearCommentForm()
		// Return to comment list instead of normal mode
		m.UIState.Mode = state.CommentsViewMode
		m.UIState.DiscardContext = nil
		return m, tea.ClearScreen
	}

	// Always return to normal mode after discard
	m.UIState.Mode = state.NormalMode
	m.UIState.DiscardContext = nil

	return m, tea.ClearScreen
}

// handleDeleteColumnConfirm handles column deletion confirmation.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDeleteColumnConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteColumn()
	case "n", "N", "esc":
		m.UIState.Mode = state.NormalMode
		return m, nil
	}
	return m, nil
}

// confirmDeleteColumn performs the actual column deletion.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		ctx, cancel := m.DBContext()
		defer cancel()
		err := m.App.ColumnService.DeleteColumn(ctx, column.ID)
		if err != nil {
			slog.Error("failed to delete column", "error", err)
			m.UIState.Mode = state.NormalMode
			return m, m.addNotification(state.LevelError, "Failed to delete column")
		} else {
			delete(m.AppState.Tasks(), column.ID)
			m.removeCurrentColumn()
		}
	}
	m.UIState.Mode = state.NormalMode
	return m, nil
}

func (m Model) handleDeleteProjectConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteProject(false)
	case "f", "F":
		if m.Forms.Input.DeleteProjectTaskCount > 0 {
			return m.confirmDeleteProject(true)
		}
		return m, nil
	case "n", "N", "esc":
		m.UIState.Mode = state.NormalMode
		return m, nil
	}
	return m, nil
}

func (m Model) confirmDeleteProject(force bool) (tea.Model, tea.Cmd) {
	currentProject := m.AppState.GetCurrentProject()
	if currentProject == nil {
		m.UIState.Mode = state.NormalMode
		return m, nil
	}

	projectName := currentProject.Name

	ctx, cancel := m.DBContext()
	defer cancel()

	err := m.App.ProjectService.DeleteProject(ctx, currentProject.ID, force)
	if err != nil {
		slog.Error("failed to delete project", "error", err)
		m.UIState.Mode = state.NormalMode
		return m, m.addNotification(state.LevelError, "Failed to delete project: "+err.Error())
	}

	m.removeCurrentProject()
	m.UIState.Mode = state.NormalMode
	return m, m.addNotification(state.LevelInfo, fmt.Sprintf("Project '%s' deleted", projectName))
}

func (m Model) handleProjectBranchConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ctx := m.UIState.ProjectBranchContext
	if ctx == nil {
		m.UIState.Mode = state.NormalMode
		m.Forms.Form.ClearProjectForm()
		return m, tea.ClearScreen
	}

	switch msg.String() {
	case "y", "Y":
		dbCtx, cancel := m.DBContext()
		defer cancel()

		emptyBranch := ""
		updateReq := project.UpdateProjectRequest{
			ID:        ctx.ExistingProject.ID,
			GitBranch: &emptyBranch,
		}

		if err := m.App.ProjectService.UpdateProject(dbCtx, updateReq); err != nil {
			slog.Error("failed to remove branch from existing project",
				"project_id", ctx.ExistingProject.ID, "error", err)
			notifCmd := m.addNotification(state.LevelError, "Failed to transfer branch association")
			m.UIState.ProjectBranchContext = nil
			m.UIState.Mode = state.NormalMode
			m.Forms.Form.ClearProjectForm()
			return m, tea.Batch(tea.ClearScreen, notifCmd)
		}

		m.createProjectWithoutDialog(ctx.ProjectName, ctx.ProjectDescription, ctx.GitBranch)
		notifCmd := m.addNotification(state.LevelInfo,
			fmt.Sprintf("Branch '%s' transferred from '%s' to '%s'",
				ctx.GitBranch, ctx.ExistingProject.Name, ctx.ProjectName))
		m.UIState.ProjectBranchContext = nil
		m.UIState.Mode = state.NormalMode
		m.Forms.Form.ClearProjectForm()
		return m, tea.Batch(tea.ClearScreen, notifCmd)

	case "n", "N":
		m.createProjectWithoutDialog(ctx.ProjectName, ctx.ProjectDescription, "")
		m.UIState.ProjectBranchContext = nil
		m.UIState.Mode = state.NormalMode
		m.Forms.Form.ClearProjectForm()
		return m, tea.ClearScreen

	case "esc":
		m.UIState.ProjectBranchContext = nil
		m.UIState.Mode = state.NormalMode
		m.Forms.Form.ClearProjectForm()
		return m, tea.ClearScreen
	}

	return m, nil
}

// handleDeleteLabelConfirm handles label deletion confirmation.
func (m Model) handleDeleteLabelConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteLabel()
	case "n", "N", "esc":
		// Clear delete state and return to label picker
		m.Pickers.Label.DeleteLabelID = 0
		m.Pickers.Label.DeleteLabelName = ""
		m.Pickers.Label.DeleteLabelTaskCount = 0
		m.UIState.Mode = state.LabelPickerMode
		return m, nil
	}
	return m, nil
}

// confirmDeleteLabel performs the actual label deletion.
func (m Model) confirmDeleteLabel() (tea.Model, tea.Cmd) {
	labelID := m.Pickers.Label.DeleteLabelID
	labelName := m.Pickers.Label.DeleteLabelName

	if labelID != 0 {
		ctx, cancel := m.DBContext()
		defer cancel()

		err := m.App.LabelService.DeleteLabel(ctx, labelID)
		if err != nil {
			slog.Error("failed to delete label", "error", err)
			m.Pickers.Label.DeleteLabelID = 0
			m.Pickers.Label.DeleteLabelName = ""
			m.Pickers.Label.DeleteLabelTaskCount = 0
			m.UIState.Mode = state.LabelPickerMode
			return m, m.addNotification(state.LevelError, "Failed to delete label")
		}
		// Remove from picker items
		newItems := make([]state.LabelPickerItem, 0, len(m.Pickers.Label.Items)-1)
		for _, item := range m.Pickers.Label.Items {
			if item.Label.ID != labelID {
				newItems = append(newItems, item)
			}
		}
		m.Pickers.Label.Items = newItems

		// Remove from global labels list
		currentLabels := m.AppState.Labels()
		newLabels := make([]*models.Label, 0, len(currentLabels)-1)
		for _, label := range currentLabels {
			if label.ID != labelID {
				newLabels = append(newLabels, label)
			}
		}
		m.AppState.SetLabels(newLabels)

		// Adjust cursor if needed
		if m.Pickers.Label.Cursor >= len(m.Pickers.Label.Items) && m.Pickers.Label.Cursor > 0 {
			m.Pickers.Label.Cursor--
		}

		// Reload task summaries to reflect removed label associations
		m.reloadCurrentColumnTasks()

		notifCmd := m.addNotification(state.LevelInfo, fmt.Sprintf("Label '%s' deleted", labelName))

		// Clear delete state and return to label picker
		m.Pickers.Label.DeleteLabelID = 0
		m.Pickers.Label.DeleteLabelName = ""
		m.Pickers.Label.DeleteLabelTaskCount = 0
		m.UIState.Mode = state.LabelPickerMode
		return m, notifCmd
	}

	// Clear delete state and return to label picker
	m.Pickers.Label.DeleteLabelID = 0
	m.Pickers.Label.DeleteLabelName = ""
	m.Pickers.Label.DeleteLabelTaskCount = 0
	m.UIState.Mode = state.LabelPickerMode
	return m, nil
}
