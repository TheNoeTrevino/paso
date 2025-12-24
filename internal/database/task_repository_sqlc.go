package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// taskRepository handles pure data access for tasks
// No business logic, no events, no validation - just database operations
type taskRepository struct {
	queries *generated.Queries
	db      *sql.DB
}

// newTaskRepository creates a new task repository
func newTaskRepository(queries *generated.Queries, db *sql.DB) *taskRepository {
	return &taskRepository{
		queries: queries,
		db:      db,
	}
}

// ============================================================================
// CRUD OPERATIONS
// ============================================================================

// CreateTask inserts a new task with the given ticket number
func (r *taskRepository) CreateTask(ctx context.Context, title, description string, columnID, position, ticketNumber int) (*models.Task, error) {
	row, err := r.queries.CreateTask(ctx, generated.CreateTaskParams{
		Title:        title,
		Description:  sql.NullString{String: description, Valid: description != ""},
		ColumnID:     int64(columnID),
		Position:     int64(position),
		TicketNumber: sql.NullInt64{Int64: int64(ticketNumber), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return toTaskModel(row), nil
}

// GetTask retrieves a task by ID
func (r *taskRepository) GetTask(ctx context.Context, taskID int) (*models.Task, error) {
	row, err := r.queries.GetTask(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task %d: %w", taskID, err)
	}
	return &models.Task{
		ID:          int(row.ID),
		Title:       row.Title,
		Description: nullStringToString(row.Description),
		ColumnID:    int(row.ColumnID),
		Position:    int(row.Position),
		CreatedAt:   nullTimeToTime(row.CreatedAt),
		UpdatedAt:   nullTimeToTime(row.UpdatedAt),
	}, nil
}

// GetTasksByColumn retrieves all tasks for a column
func (r *taskRepository) GetTasksByColumn(ctx context.Context, columnID int) ([]*models.Task, error) {
	rows, err := r.queries.GetTasksByColumn(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for column %d: %w", columnID, err)
	}

	tasks := make([]*models.Task, len(rows))
	for i, row := range rows {
		tasks[i] = &models.Task{
			ID:          int(row.ID),
			Title:       row.Title,
			Description: nullStringToString(row.Description),
			ColumnID:    int(row.ColumnID),
			Position:    int(row.Position),
			CreatedAt:   nullTimeToTime(row.CreatedAt),
			UpdatedAt:   nullTimeToTime(row.UpdatedAt),
		}
	}
	return tasks, nil
}

// GetTaskCountByColumn returns the number of tasks in a column
func (r *taskRepository) GetTaskCountByColumn(ctx context.Context, columnID int) (int, error) {
	count, err := r.queries.GetTaskCountByColumn(ctx, int64(columnID))
	if err != nil {
		return 0, fmt.Errorf("failed to get task count for column %d: %w", columnID, err)
	}
	return int(count), nil
}

// UpdateTask updates a task's title and description
func (r *taskRepository) UpdateTask(ctx context.Context, taskID int, title, description string) error {
	err := r.queries.UpdateTask(ctx, generated.UpdateTaskParams{
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
		ID:          int64(taskID),
	})
	if err != nil {
		return fmt.Errorf("failed to update task %d: %w", taskID, err)
	}
	return nil
}

// UpdateTaskPriority updates a task's priority
func (r *taskRepository) UpdateTaskPriority(ctx context.Context, taskID, priorityID int) error {
	err := r.queries.UpdateTaskPriority(ctx, generated.UpdateTaskPriorityParams{
		PriorityID: int64(priorityID),
		ID:         int64(taskID),
	})
	if err != nil {
		return fmt.Errorf("failed to update task %d priority: %w", taskID, err)
	}
	return nil
}

// UpdateTaskType updates a task's type
func (r *taskRepository) UpdateTaskType(ctx context.Context, taskID, typeID int) error {
	err := r.queries.UpdateTaskType(ctx, generated.UpdateTaskTypeParams{
		TypeID: int64(typeID),
		ID:     int64(taskID),
	})
	if err != nil {
		return fmt.Errorf("failed to update task %d type: %w", taskID, err)
	}
	return nil
}

// DeleteTask removes a task
func (r *taskRepository) DeleteTask(ctx context.Context, taskID int) error {
	err := r.queries.DeleteTask(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to delete task %d: %w", taskID, err)
	}
	return nil
}

// ============================================================================
// TASK SUMMARIES
// ============================================================================

// GetTaskSummariesByColumn retrieves task summaries for a column
func (r *taskRepository) GetTaskSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error) {
	rows, err := r.queries.GetTaskSummariesByColumn(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task summaries for column %d: %w", columnID, err)
	}

	summaries := make([]*models.TaskSummary, len(rows))
	for i, row := range rows {
		summaries[i] = &models.TaskSummary{
			ID:                  int(row.ID),
			Title:               row.Title,
			ColumnID:            int(row.ColumnID),
			Position:            int(row.Position),
			TypeDescription:     nullStringToString(row.Description),
			PriorityDescription: nullStringToString(row.Description_2),
			PriorityColor:       nullStringToString(row.Color),
			Labels:              parseLabelsFromStrings(row.LabelIds, row.LabelNames, row.LabelColors),
		}
	}
	return summaries, nil
}

// GetTaskSummariesByProject retrieves all task summaries for a project
func (r *taskRepository) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	rows, err := r.queries.GetTaskSummariesByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task summaries for project %d: %w", projectID, err)
	}

	tasksByColumn := make(map[int][]*models.TaskSummary)
	for _, row := range rows {
		summary := &models.TaskSummary{
			ID:                  int(row.ID),
			Title:               row.Title,
			ColumnID:            int(row.ColumnID),
			Position:            int(row.Position),
			TypeDescription:     nullStringToString(row.Description),
			PriorityDescription: nullStringToString(row.Description_2),
			PriorityColor:       nullStringToString(row.Color),
			Labels:              parseLabelsFromStrings(row.LabelIds, row.LabelNames, row.LabelColors),
			IsBlocked:           row.IsBlocked > 0,
		}
		tasksByColumn[summary.ColumnID] = append(tasksByColumn[summary.ColumnID], summary)
	}
	return tasksByColumn, nil
}

// GetTaskSummariesByProjectFiltered retrieves filtered task summaries
func (r *taskRepository) GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error) {
	rows, err := r.queries.GetTaskSummariesByProjectFiltered(ctx, generated.GetTaskSummariesByProjectFilteredParams{
		ProjectID: int64(projectID),
		Title:     "%" + searchQuery + "%",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered task summaries for project %d: %w", projectID, err)
	}

	tasksByColumn := make(map[int][]*models.TaskSummary)
	for _, row := range rows {
		summary := &models.TaskSummary{
			ID:                  int(row.ID),
			Title:               row.Title,
			ColumnID:            int(row.ColumnID),
			Position:            int(row.Position),
			TypeDescription:     nullStringToString(row.Description),
			PriorityDescription: nullStringToString(row.Description_2),
			PriorityColor:       nullStringToString(row.Color),
			Labels:              parseLabelsFromStrings(row.LabelIds, row.LabelNames, row.LabelColors),
			IsBlocked:           row.IsBlocked > 0,
		}
		tasksByColumn[summary.ColumnID] = append(tasksByColumn[summary.ColumnID], summary)
	}
	return tasksByColumn, nil
}

// GetTaskDetail retrieves full task details
func (r *taskRepository) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
	row, err := r.queries.GetTaskDetail(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task detail for task %d: %w", taskID, err)
	}

	detail := &models.TaskDetail{
		ID:                  int(row.ID),
		Title:               row.Title,
		Description:         nullStringToString(row.Description),
		ColumnID:            int(row.ColumnID),
		Position:            int(row.Position),
		TicketNumber:        int(row.TicketNumber.Int64),
		CreatedAt:           nullTimeToTime(row.CreatedAt),
		UpdatedAt:           nullTimeToTime(row.UpdatedAt),
		TypeDescription:     nullStringToString(row.Description_2),
		PriorityDescription: nullStringToString(row.Description_3),
		PriorityColor:       nullStringToString(row.Color),
	}

	// Get labels
	labels, err := r.queries.GetTaskLabels(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get labels for task %d: %w", taskID, err)
	}
	detail.Labels = toLabelsFromGenerated(labels)

	// Get parent tasks
	parentTasks, err := r.GetParentTasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent tasks for task %d: %w", taskID, err)
	}
	detail.ParentTasks = parentTasks

	// Get child tasks
	childTasks, err := r.GetChildTasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child tasks for task %d: %w", taskID, err)
	}
	detail.ChildTasks = childTasks

	return detail, nil
}

// ============================================================================
// TASK MOVEMENT
// ============================================================================

// GetTaskPosition retrieves a task's current column and position
func (r *taskRepository) GetTaskPosition(ctx context.Context, taskID int) (columnID, position int, err error) {
	row, err := r.queries.GetTaskPosition(ctx, int64(taskID))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get task %d position: %w", taskID, err)
	}
	return int(row.ColumnID), int(row.Position), nil
}

// GetNextColumnID retrieves the next column ID in the linked list
func (r *taskRepository) GetNextColumnID(ctx context.Context, columnID int) (*int, error) {
	result, err := r.queries.GetNextColumnID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get next column for %d: %w", columnID, err)
	}
	return interfaceToIntPtr(result), nil
}

// GetPrevColumnID retrieves the previous column ID in the linked list
func (r *taskRepository) GetPrevColumnID(ctx context.Context, columnID int) (*int, error) {
	result, err := r.queries.GetPrevColumnID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get prev column for %d: %w", columnID, err)
	}
	return interfaceToIntPtr(result), nil
}

// MoveTaskToColumn moves a task to a different column
func (r *taskRepository) MoveTaskToColumn(ctx context.Context, taskID, columnID, position int) error {
	err := r.queries.MoveTaskToColumn(ctx, generated.MoveTaskToColumnParams{
		ColumnID: int64(columnID),
		Position: int64(position),
		ID:       int64(taskID),
	})
	if err != nil {
		return fmt.Errorf("failed to move task %d to column %d: %w", taskID, columnID, err)
	}
	return nil
}

// SetTaskPosition updates a task's position
func (r *taskRepository) SetTaskPosition(ctx context.Context, taskID, position int) error {
	err := r.queries.SetTaskPosition(ctx, generated.SetTaskPositionParams{
		Position: int64(position),
		ID:       int64(taskID),
	})
	if err != nil {
		return fmt.Errorf("failed to set task %d position: %w", taskID, err)
	}
	return nil
}

// SetTaskPositionTemporary sets a task's position to -1 (for swapping)
func (r *taskRepository) SetTaskPositionTemporary(ctx context.Context, taskID int) error {
	err := r.queries.SetTaskPositionTemporary(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to set task %d to temporary position: %w", taskID, err)
	}
	return nil
}

// GetTaskAbove finds the task above the given position
func (r *taskRepository) GetTaskAbove(ctx context.Context, columnID, position int) (taskID, taskPosition int, err error) {
	row, err := r.queries.GetTaskAbove(ctx, generated.GetTaskAboveParams{
		ColumnID: int64(columnID),
		Position: int64(position),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, models.ErrAlreadyFirstTask
		}
		return 0, 0, fmt.Errorf("failed to get task above position %d: %w", position, err)
	}
	return int(row.ID), int(row.Position), nil
}

// GetTaskBelow finds the task below the given position
func (r *taskRepository) GetTaskBelow(ctx context.Context, columnID, position int) (taskID, taskPosition int, err error) {
	row, err := r.queries.GetTaskBelow(ctx, generated.GetTaskBelowParams{
		ColumnID: int64(columnID),
		Position: int64(position),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, models.ErrAlreadyLastTask
		}
		return 0, 0, fmt.Errorf("failed to get task below position %d: %w", position, err)
	}
	return int(row.ID), int(row.Position), nil
}

// ============================================================================
// TASK RELATIONSHIPS
// ============================================================================

// GetParentTasks retrieves parent tasks (tasks that depend on this task)
func (r *taskRepository) GetParentTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	rows, err := r.queries.GetParentTasks(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get parent tasks for task %d: %w", taskID, err)
	}

	refs := make([]*models.TaskReference, len(rows))
	for i, row := range rows {
		refs[i] = &models.TaskReference{
			ID:             int(row.ID),
			TicketNumber:   int(row.TicketNumber.Int64),
			Title:          row.Title,
			ProjectName:    row.Name,
			RelationTypeID: int(row.ID_2),
			RelationLabel:  row.PToCLabel,
			RelationColor:  row.Color,
			IsBlocking:     row.IsBlocking,
		}
	}
	return refs, nil
}

// GetChildTasks retrieves child tasks (tasks this task depends on)
func (r *taskRepository) GetChildTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	rows, err := r.queries.GetChildTasks(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get child tasks for task %d: %w", taskID, err)
	}

	refs := make([]*models.TaskReference, len(rows))
	for i, row := range rows {
		refs[i] = &models.TaskReference{
			ID:             int(row.ID),
			TicketNumber:   int(row.TicketNumber.Int64),
			Title:          row.Title,
			ProjectName:    row.Name,
			RelationTypeID: int(row.ID_2),
			RelationLabel:  row.CToPLabel,
			RelationColor:  row.Color,
			IsBlocking:     row.IsBlocking,
		}
	}
	return refs, nil
}

// GetTaskReferencesForProject retrieves all task references for a project
func (r *taskRepository) GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error) {
	rows, err := r.queries.GetTaskReferencesForProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task references for project %d: %w", projectID, err)
	}

	refs := make([]*models.TaskReference, len(rows))
	for i, row := range rows {
		refs[i] = &models.TaskReference{
			ID:           int(row.ID),
			TicketNumber: int(row.TicketNumber.Int64),
			Title:        row.Title,
			ProjectName:  row.Name,
		}
	}
	return refs, nil
}

// AddSubtask creates a parent-child relationship
func (r *taskRepository) AddSubtask(ctx context.Context, parentID, childID int) error {
	err := r.queries.AddSubtask(ctx, generated.AddSubtaskParams{
		ParentID: int64(parentID),
		ChildID:  int64(childID),
	})
	if err != nil {
		return fmt.Errorf("failed to add subtask relationship: %w", err)
	}
	return nil
}

// AddSubtaskWithRelationType creates a typed parent-child relationship
func (r *taskRepository) AddSubtaskWithRelationType(ctx context.Context, parentID, childID, relationTypeID int) error {
	err := r.queries.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
		ParentID:       int64(parentID),
		ChildID:        int64(childID),
		RelationTypeID: int64(relationTypeID),
	})
	if err != nil {
		return fmt.Errorf("failed to add subtask relationship with type: %w", err)
	}
	return nil
}

// RemoveSubtask removes a parent-child relationship
func (r *taskRepository) RemoveSubtask(ctx context.Context, parentID, childID int) error {
	err := r.queries.RemoveSubtask(ctx, generated.RemoveSubtaskParams{
		ParentID: int64(parentID),
		ChildID:  int64(childID),
	})
	if err != nil {
		return fmt.Errorf("failed to remove subtask relationship: %w", err)
	}
	return nil
}

// GetAllRelationTypes retrieves all relation types
func (r *taskRepository) GetAllRelationTypes(ctx context.Context) ([]*models.RelationType, error) {
	rows, err := r.queries.GetAllRelationTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get relation types: %w", err)
	}

	types := make([]*models.RelationType, len(rows))
	for i, row := range rows {
		types[i] = &models.RelationType{
			ID:         int(row.ID),
			PToCLabel:  row.PToCLabel,
			CToPLabel:  row.CToPLabel,
			Color:      row.Color,
			IsBlocking: row.IsBlocking,
		}
	}
	return types, nil
}

// ============================================================================
// HELPERS
// ============================================================================

// GetProjectIDFromTask gets the project ID for a task
func (r *taskRepository) GetProjectIDFromTask(ctx context.Context, taskID int) (int, error) {
	projectID, err := r.queries.GetProjectIDFromTask(ctx, int64(taskID))
	if err != nil {
		return 0, fmt.Errorf("failed to get project for task %d: %w", taskID, err)
	}
	return int(projectID), nil
}

// GetProjectIDFromColumn gets the project ID for a column
func (r *taskRepository) GetProjectIDFromColumn(ctx context.Context, columnID int) (int, error) {
	projectID, err := r.queries.GetProjectIDFromColumn(ctx, int64(columnID))
	if err != nil {
		return 0, fmt.Errorf("failed to get project for column %d: %w", columnID, err)
	}
	return int(projectID), nil
}

// GetNextTicketNumber retrieves the next ticket number for a project
func (r *taskRepository) GetNextTicketNumber(ctx context.Context, projectID int) (int, error) {
	ticketNum, err := r.queries.GetNextTicketNumber(ctx, int64(projectID))
	if err != nil {
		return 0, fmt.Errorf("failed to get next ticket number for project %d: %w", projectID, err)
	}
	return int(ticketNum.Int64), nil
}

// IncrementTicketNumber increments the ticket counter for a project
func (r *taskRepository) IncrementTicketNumber(ctx context.Context, projectID int) error {
	err := r.queries.IncrementTicketNumber(ctx, int64(projectID))
	if err != nil {
		return fmt.Errorf("failed to increment ticket number for project %d: %w", projectID, err)
	}
	return nil
}

// GetAllPriorities retrieves all priorities
func (r *taskRepository) GetAllPriorities(ctx context.Context) ([]*models.Priority, error) {
	rows, err := r.queries.GetAllPriorities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get priorities: %w", err)
	}

	priorities := make([]*models.Priority, len(rows))
	for i, row := range rows {
		priorities[i] = &models.Priority{
			ID:          int(row.ID),
			Description: row.Description,
			Color:       row.Color,
		}
	}
	return priorities, nil
}

// GetAllTypes retrieves all types
func (r *taskRepository) GetAllTypes(ctx context.Context) ([]*models.Type, error) {
	rows, err := r.queries.GetAllTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get types: %w", err)
	}

	types := make([]*models.Type, len(rows))
	for i, row := range rows {
		types[i] = &models.Type{
			ID:          int(row.ID),
			Description: row.Description,
		}
	}
	return types, nil
}

// WithTx returns a new repository instance that uses the given transaction
func (r *taskRepository) WithTx(tx *sql.Tx) *taskRepository {
	return &taskRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

// BeginTx starts a new transaction
func (r *taskRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ============================================================================
// MODEL CONVERSION HELPERS
// ============================================================================

func toTaskModel(row generated.Task) *models.Task {
	return &models.Task{
		ID:          int(row.ID),
		Title:       row.Title,
		Description: nullStringToString(row.Description),
		ColumnID:    int(row.ColumnID),
		Position:    int(row.Position),
		TypeID:      int(row.TypeID),
		PriorityID:  int(row.PriorityID),
		CreatedAt:   nullTimeToTime(row.CreatedAt),
		UpdatedAt:   nullTimeToTime(row.UpdatedAt),
	}
}

func toLabelsFromGenerated(rows []generated.Label) []*models.Label {
	labels := make([]*models.Label, len(rows))
	for i, row := range rows {
		labels[i] = &models.Label{
			ID:        int(row.ID),
			Name:      row.Name,
			Color:     row.Color,
			ProjectID: int(row.ProjectID),
		}
	}
	return labels
}

func parseLabelsFromStrings(labelIDsStr, labelNamesStr, labelColorsStr string) []*models.Label {
	// Handle empty strings (no labels)
	if labelIDsStr == "" || labelNamesStr == "" || labelColorsStr == "" {
		return []*models.Label{}
	}

	const delimiter = "\x1F"
	ids := strings.Split(labelIDsStr, delimiter)
	names := strings.Split(labelNamesStr, delimiter)
	colors := strings.Split(labelColorsStr, delimiter)

	if len(ids) != len(names) || len(ids) != len(colors) {
		return []*models.Label{}
	}

	labels := make([]*models.Label, 0, len(ids))
	for i := range ids {
		id, err := strconv.Atoi(ids[i])
		if err != nil {
			continue
		}
		labels = append(labels, &models.Label{
			ID:    id,
			Name:  names[i],
			Color: colors[i],
		})
	}
	return labels
}
