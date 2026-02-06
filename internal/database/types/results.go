package types

// GetChildTasksRow represents the result of GetChildTasks query
type GetChildTasksRow struct {
	ID           int64
	TicketNumber NullInt64
	Title        string
	Name         string
	ID_2         int64
	CToPLabel    string
	Color        string
	IsBlocking   bool
}

// GetInProgressTaskDetailsRow represents the result of GetInProgressTaskDetails query
type GetInProgressTaskDetailsRow struct {
	ID                  int64
	TicketNumber        NullInt64
	Title               string
	Description         NullString
	ColumnID            int64
	Position            int64
	CreatedAt           NullTime
	UpdatedAt           NullTime
	ColumnName          string
	ProjectName         string
	TypeDescription     NullString
	PriorityDescription NullString
	PriorityColor       NullString
	AssigneeID          NullInt64
	AssigneeName        NullString
	LabelIds            string
	LabelNames          string
	LabelColors         string
	IsBlocked           bool
}

// GetInProgressTasksByProjectRow represents the result of GetInProgressTasksByProject query
type GetInProgressTasksByProjectRow struct {
	ID           int64
	TicketNumber NullInt64
	Title        string
	Description  NullString
	ColumnName   string
	ProjectName  string
}

// GetParentTasksRow represents the result of GetParentTasks query
type GetParentTasksRow struct {
	ID           int64
	TicketNumber NullInt64
	Title        string
	Name         string
	ID_2         int64
	PToCLabel    string
	Color        string
	IsBlocking   bool
}

// GetReadyTaskSummariesByProjectRow represents the result of GetReadyTaskSummariesByProject query
type GetReadyTaskSummariesByProjectRow struct {
	ID                  int64
	Title               string
	ColumnID            int64
	Position            int64
	TypeDescription     NullString
	PriorityDescription NullString
	PriorityColor       NullString
	AssigneeID          NullInt64
	AssigneeName        NullString
	LabelIds            string
	LabelNames          string
	LabelColors         string
	IsBlocked           bool
}

// GetTaskRow represents the result of GetTask query
type GetTaskRow struct {
	ID          int64
	Title       string
	Description NullString
	ColumnID    int64
	Position    int64
	CreatedAt   NullTime
	UpdatedAt   NullTime
}

// GetTaskAboveRow represents the result of GetTaskAbove query
type GetTaskAboveRow struct {
	ID       int64
	Position int64
}

// GetTaskBelowRow represents the result of GetTaskBelow query
type GetTaskBelowRow struct {
	ID       int64
	Position int64
}

// GetTaskDetailRow represents the result of GetTaskDetail query
type GetTaskDetailRow struct {
	ID                  int64
	Title               string
	Description         NullString
	ColumnID            int64
	Position            int64
	TicketNumber        NullInt64
	CreatedAt           NullTime
	UpdatedAt           NullTime
	TypeDescription     NullString
	PriorityDescription NullString
	PriorityColor       NullString
	ColumnName          string
	ProjectName         string
	AssigneeID          NullInt64
	AssigneeName        NullString
	IsBlocked           bool
}

// GetTaskPositionRow represents the result of GetTaskPosition query
type GetTaskPositionRow struct {
	ColumnID int64
	Position int64
}

// GetTaskReferencesForProjectRow represents the result of GetTaskReferencesForProject query
type GetTaskReferencesForProjectRow struct {
	ID           int64
	TicketNumber NullInt64
	Title        string
	Name         string
}

// GetTaskRelationsForProjectRow represents the result of GetTaskRelationsForProject query
type GetTaskRelationsForProjectRow struct {
	ParentID      int64
	ChildID       int64
	RelationLabel string
	RelationColor string
	IsBlocking    bool
}

// GetTaskSummariesByColumnRow represents the result of GetTaskSummariesByColumn query
type GetTaskSummariesByColumnRow struct {
	ID                  int64
	Title               string
	ColumnID            int64
	Position            int64
	TypeDescription     NullString
	PriorityDescription NullString
	PriorityColor       NullString
	AssigneeID          NullInt64
	AssigneeName        NullString
	LabelIds            string
	LabelNames          string
	LabelColors         string
}

// GetTaskSummariesByProjectRow represents the result of GetTaskSummariesByProject query
type GetTaskSummariesByProjectRow struct {
	ID                  int64
	Title               string
	ColumnID            int64
	Position            int64
	TypeDescription     NullString
	PriorityDescription NullString
	PriorityColor       NullString
	AssigneeID          NullInt64
	AssigneeName        NullString
	LabelIds            string
	LabelNames          string
	LabelColors         string
	IsBlocked           bool
}

// GetTaskSummariesByProjectFilteredRow represents the result of GetTaskSummariesByProjectFiltered query
type GetTaskSummariesByProjectFilteredRow struct {
	ID                  int64
	Title               string
	ColumnID            int64
	Position            int64
	TypeDescription     NullString
	PriorityDescription NullString
	PriorityColor       NullString
	AssigneeID          NullInt64
	AssigneeName        NullString
	LabelIds            string
	LabelNames          string
	LabelColors         string
	IsBlocked           bool
}

// GetTasksByColumnRow represents the result of GetTasksByColumn query
type GetTasksByColumnRow struct {
	ID          int64
	Title       string
	Description NullString
	ColumnID    int64
	Position    int64
	CreatedAt   NullTime
	UpdatedAt   NullTime
}

// GetTasksForTreeRow represents the result of GetTasksForTree query
type GetTasksForTreeRow struct {
	ID           int64
	TicketNumber NullInt64
	Title        string
	ColumnName   string
	ProjectName  string
	IsCompleted  bool
}

// GetColumnByIDRow represents the result of GetColumnByID query
type GetColumnByIDRow struct {
	ID                   int64
	Name                 string
	ProjectID            int64
	PrevID               NullInt64
	NextID               NullInt64
	HoldsReadyTasks      bool
	HoldsCompletedTasks  bool
	HoldsInProgressTasks bool
}

// GetColumnLinkedListInfoRow represents the result of GetColumnLinkedListInfo query
type GetColumnLinkedListInfoRow struct {
	PrevID    NullInt64
	NextID    NullInt64
	ProjectID int64
}

// GetColumnsByProjectRow represents the result of GetColumnsByProject query
type GetColumnsByProjectRow struct {
	ID                   int64
	Name                 string
	ProjectID            int64
	PrevID               NullInt64
	NextID               NullInt64
	HoldsReadyTasks      bool
	HoldsCompletedTasks  bool
	HoldsInProgressTasks bool
}

// GetCompletedColumnByProjectRow represents the result of GetCompletedColumnByProject query
type GetCompletedColumnByProjectRow struct {
	ID                  int64
	Name                string
	ProjectID           int64
	PrevID              NullInt64
	NextID              NullInt64
	HoldsCompletedTasks bool
}

// GetInProgressColumnByProjectRow represents the result of GetInProgressColumnByProject query
type GetInProgressColumnByProjectRow struct {
	ID                   int64
	Name                 string
	ProjectID            int64
	PrevID               NullInt64
	NextID               NullInt64
	HoldsInProgressTasks bool
}

// GetReadyColumnByProjectRow represents the result of GetReadyColumnByProject query
type GetReadyColumnByProjectRow struct {
	ID              int64
	Name            string
	ProjectID       int64
	PrevID          NullInt64
	NextID          NullInt64
	HoldsReadyTasks bool
}
