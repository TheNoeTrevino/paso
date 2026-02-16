package types

// labels.sql.go params
type AddLabelToTaskParams struct {
	TaskID  int64
	LabelID int64
}

type CreateLabelParams struct {
	Name      string
	Color     string
	ProjectID int64
}

type InsertTaskLabelParams struct {
	TaskID  int64
	LabelID int64
}

type RemoveLabelFromTaskParams struct {
	TaskID  int64
	LabelID int64
}

type UpdateLabelParams struct {
	Name  string
	Color string
	ID    int64
}

type UpsertLabelParams struct {
	Name      string
	Color     string
	ProjectID int64
}

// columns.sql.go params
type CreateColumnParams struct {
	Name                 string
	ProjectID            int64
	PrevID               NullInt64
	NextID               NullInt64
	HoldsReadyTasks      bool
	HoldsCompletedTasks  bool
	HoldsInProgressTasks bool
}

type UpdateColumnHoldsCompletedTasksParams struct {
	HoldsCompletedTasks bool
	ID                  int64
}

type UpdateColumnHoldsInProgressTasksParams struct {
	HoldsInProgressTasks bool
	ID                   int64
}

type UpdateColumnHoldsReadyTasksParams struct {
	HoldsReadyTasks bool
	ID              int64
}

type UpdateColumnNameParams struct {
	Name string
	ID   int64
}

type UpdateColumnNextIDParams struct {
	NextID NullInt64
	ID     int64
}

type UpdateColumnPrevIDParams struct {
	PrevID NullInt64
	ID     int64
}

// projects.sql.go params
type CreateProjectRecordParams struct {
	Name        string
	Description NullString
	GitBranch   NullString
}

type UpdateProjectParams struct {
	Name        string
	Description NullString
	GitBranch   NullString
	ID          int64
}

// tasks.sql.go params
type AddSubtaskParams struct {
	ParentID int64
	ChildID  int64
}

type AddSubtaskWithRelationTypeParams struct {
	ParentID       int64
	ChildID        int64
	RelationTypeID int64
}

type CreateTaskParams struct {
	Title        string
	Description  NullString
	ColumnID     int64
	Position     int64
	TicketNumber NullInt64
	AssigneeID   NullInt64
	Estimate     NullString
	DueDate      NullTime
}

type GetTaskAboveParams struct {
	ColumnID int64
	Position int64
}

type GetTaskBelowParams struct {
	ColumnID int64
	Position int64
}

type GetTaskSummariesWithFiltersParams struct {
	ProjectID   int64
	TitleFilter NullString
	PriorityID  NullInt64
	TypeID      NullInt64
	AssigneeID  NullInt64
	LabelIdsCsv string
}

type MoveTaskToColumnParams struct {
	ColumnID int64
	Position int64
	ID       int64
}

type RemoveSubtaskParams struct {
	ParentID int64
	ChildID  int64
}

type SetTaskPositionParams struct {
	Position int64
	ID       int64
}

type UpdateTaskParams struct {
	Title       string
	Description NullString
	ID          int64
}

type UpdateTaskPriorityParams struct {
	PriorityID int64
	ID         int64
}

type UpdateTaskTypeParams struct {
	TypeID int64
	ID     int64
}

type UpdateTaskAssigneeParams struct {
	AssigneeID NullInt64
	ID         int64
}

type UpdateTaskEstimateParams struct {
	Estimate NullString
	ID       int64
}

type UpdateTaskDueDateParams struct {
	DueDate NullTime
	ID      int64
}

// comments.sql.go params
type CreateCommentParams struct {
	TaskID  int64
	Content string
	Author  string
}

type UpdateCommentParams struct {
	Content string
	ID      int64
}

// task_events.sql.go params
// CreateTaskEventParams contains the parameters for creating a task event.
type CreateTaskEventParams struct {
	TaskID  int64
	Content string
	Author  string
}
