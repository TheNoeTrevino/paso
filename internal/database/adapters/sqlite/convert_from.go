package sqlite

import (
	"github.com/thenoetrevino/paso/internal/database/generated_sqlite"
	"github.com/thenoetrevino/paso/internal/database/types"
)

func fromGeneratedAssignee(g generated_sqlite.Assignee) types.Assignee {
	return types.Assignee{
		ID:        g.ID,
		Name:      g.Name,
		CreatedAt: types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt: types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedTask(g generated_sqlite.Task) types.Task {
	return types.Task{
		ID:           g.ID,
		Title:        g.Title,
		Description:  types.FromSQLNullString(g.Description),
		ColumnID:     g.ColumnID,
		Position:     g.Position,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		TypeID:       g.TypeID,
		PriorityID:   g.PriorityID,
		CreatedAt:    types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:    types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedProject(g generated_sqlite.Project) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.NullStringFromInterface(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetAllProjectsRow(g generated_sqlite.GetAllProjectsRow) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.NullStringFromInterface(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetProjectByIDRow(g generated_sqlite.GetProjectByIDRow) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.NullStringFromInterface(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetProjectByGitBranchRow(g generated_sqlite.GetProjectByGitBranchRow) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.NullStringFromInterface(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedColumn(g generated_sqlite.Column) types.Column {
	return types.Column{
		ID:                   g.ID,
		Name:                 g.Name,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		ProjectID:            g.ProjectID,
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedLabel(g generated_sqlite.Label) types.Label {
	return types.Label{
		ID:        g.ID,
		Name:      g.Name,
		Color:     g.Color,
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedTaskComment(g generated_sqlite.TaskComment) types.TaskComment {
	return types.TaskComment{
		ID:        g.ID,
		TaskID:    g.TaskID,
		Content:   g.Content,
		Author:    g.Author,
		CreatedAt: types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt: types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedPriority(g generated_sqlite.Priority) types.Priority {
	return types.Priority{
		ID:          g.ID,
		Description: g.Description,
		Color:       g.Color,
	}
}

func fromGeneratedRelationType(g generated_sqlite.RelationType) types.RelationType {
	return types.RelationType{
		ID:         g.ID,
		PToCLabel:  g.PToCLabel,
		CToPLabel:  g.CToPLabel,
		Color:      g.Color,
		IsBlocking: g.IsBlocking,
	}
}

func fromGeneratedType(g generated_sqlite.Type) types.Type {
	return types.Type{
		ID:          g.ID,
		Description: g.Description,
	}
}

func fromGeneratedGetChildTasksRow(g generated_sqlite.GetChildTasksRow) types.GetChildTasksRow {
	return types.GetChildTasksRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
		ID_2:         g.ID_2,
		CToPLabel:    g.CToPLabel,
		Color:        g.Color,
		IsBlocking:   g.IsBlocking,
	}
}

func fromGeneratedGetInProgressTaskDetailsRow(g generated_sqlite.GetInProgressTaskDetailsRow) types.GetInProgressTaskDetailsRow {
	return types.GetInProgressTaskDetailsRow{
		ID:                  g.ID,
		TicketNumber:        types.FromSQLNullInt64(g.TicketNumber),
		Title:               g.Title,
		Description:         types.FromSQLNullString(g.Description),
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		CreatedAt:           types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:           types.FromSQLNullTime(g.UpdatedAt),
		ColumnName:          g.ColumnName,
		ProjectName:         g.ProjectName,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.NullInt64FromInterface(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetInProgressTasksByProjectRow(g generated_sqlite.GetInProgressTasksByProjectRow) types.GetInProgressTasksByProjectRow {
	return types.GetInProgressTasksByProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Description:  types.FromSQLNullString(g.Description),
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
	}
}

func fromGeneratedGetParentTasksRow(g generated_sqlite.GetParentTasksRow) types.GetParentTasksRow {
	return types.GetParentTasksRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
		ID_2:         g.ID_2,
		PToCLabel:    g.PToCLabel,
		Color:        g.Color,
		IsBlocking:   g.IsBlocking,
	}
}

func fromGeneratedGetReadyTaskSummariesByProjectRow(g generated_sqlite.GetReadyTaskSummariesByProjectRow) types.GetReadyTaskSummariesByProjectRow {
	return types.GetReadyTaskSummariesByProjectRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.NullStringFromInterface(g.Estimate),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.NullInt64FromInterface(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTaskRow(g generated_sqlite.GetTaskRow) types.GetTaskRow {
	return types.GetTaskRow{
		ID:          g.ID,
		Title:       g.Title,
		Description: types.FromSQLNullString(g.Description),
		ColumnID:    g.ColumnID,
		Position:    g.Position,
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetTaskAboveRow(g generated_sqlite.GetTaskAboveRow) types.GetTaskAboveRow {
	return types.GetTaskAboveRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskBelowRow(g generated_sqlite.GetTaskBelowRow) types.GetTaskBelowRow {
	return types.GetTaskBelowRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskDetailRow(g generated_sqlite.GetTaskDetailRow) types.GetTaskDetailRow {
	return types.GetTaskDetailRow{
		ID:                  g.ID,
		Title:               g.Title,
		Description:         types.FromSQLNullString(g.Description),
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		TicketNumber:        types.FromSQLNullInt64(g.TicketNumber),
		CreatedAt:           types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:           types.FromSQLNullTime(g.UpdatedAt),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		ColumnName:          g.ColumnName,
		ProjectName:         g.ProjectName,
		AssigneeID:          types.NullInt64FromInterface(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTaskPositionRow(g generated_sqlite.GetTaskPositionRow) types.GetTaskPositionRow {
	return types.GetTaskPositionRow{
		ColumnID: g.ColumnID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskReferencesForProjectRow(g generated_sqlite.GetTaskReferencesForProjectRow) types.GetTaskReferencesForProjectRow {
	return types.GetTaskReferencesForProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
	}
}

func fromGeneratedGetTaskRelationsForProjectRow(g generated_sqlite.GetTaskRelationsForProjectRow) types.GetTaskRelationsForProjectRow {
	return types.GetTaskRelationsForProjectRow{
		ParentID:      g.ParentID,
		ChildID:       g.ChildID,
		RelationLabel: g.RelationLabel,
		RelationColor: g.RelationColor,
		IsBlocking:    g.IsBlocking,
	}
}

func fromGeneratedGetTaskSummariesByColumnRow(g generated_sqlite.GetTaskSummariesByColumnRow) types.GetTaskSummariesByColumnRow {
	return types.GetTaskSummariesByColumnRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.NullStringFromInterface(g.Estimate),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.NullInt64FromInterface(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
	}
}

func fromGeneratedGetTaskSummariesByProjectRow(g generated_sqlite.GetTaskSummariesByProjectRow) types.GetTaskSummariesByProjectRow {
	return types.GetTaskSummariesByProjectRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.NullStringFromInterface(g.Estimate),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.NullInt64FromInterface(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTaskSummariesByProjectFilteredRow(g generated_sqlite.GetTaskSummariesByProjectFilteredRow) types.GetTaskSummariesByProjectFilteredRow {
	return types.GetTaskSummariesByProjectFilteredRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.NullStringFromInterface(g.Estimate),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.NullInt64FromInterface(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked != 0,
	}
}

func fromGeneratedGetTasksByColumnRow(g generated_sqlite.GetTasksByColumnRow) types.GetTasksByColumnRow {
	return types.GetTasksByColumnRow{
		ID:          g.ID,
		Title:       g.Title,
		Description: types.FromSQLNullString(g.Description),
		ColumnID:    g.ColumnID,
		Position:    g.Position,
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetTasksForTreeRow(g generated_sqlite.GetTasksForTreeRow) types.GetTasksForTreeRow {
	return types.GetTasksForTreeRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
		IsCompleted:  g.IsCompleted,
	}
}

func fromGeneratedGetColumnByIDRow(g generated_sqlite.GetColumnByIDRow) types.GetColumnByIDRow {
	return types.GetColumnByIDRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetColumnLinkedListInfoRow(g generated_sqlite.GetColumnLinkedListInfoRow) types.GetColumnLinkedListInfoRow {
	return types.GetColumnLinkedListInfoRow{
		PrevID:    types.NullInt64FromInterface(g.PrevID),
		NextID:    types.NullInt64FromInterface(g.NextID),
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedGetColumnsByProjectRow(g generated_sqlite.GetColumnsByProjectRow) types.GetColumnsByProjectRow {
	return types.GetColumnsByProjectRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetCompletedColumnByProjectRow(g generated_sqlite.GetCompletedColumnByProjectRow) types.GetCompletedColumnByProjectRow {
	return types.GetCompletedColumnByProjectRow{
		ID:                  g.ID,
		Name:                g.Name,
		ProjectID:           g.ProjectID,
		PrevID:              types.NullInt64FromInterface(g.PrevID),
		NextID:              types.NullInt64FromInterface(g.NextID),
		HoldsCompletedTasks: g.HoldsCompletedTasks,
	}
}

func fromGeneratedGetInProgressColumnByProjectRow(g generated_sqlite.GetInProgressColumnByProjectRow) types.GetInProgressColumnByProjectRow {
	return types.GetInProgressColumnByProjectRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.NullInt64FromInterface(g.PrevID),
		NextID:               types.NullInt64FromInterface(g.NextID),
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetReadyColumnByProjectRow(g generated_sqlite.GetReadyColumnByProjectRow) types.GetReadyColumnByProjectRow {
	return types.GetReadyColumnByProjectRow{
		ID:              g.ID,
		Name:            g.Name,
		ProjectID:       g.ProjectID,
		PrevID:          types.NullInt64FromInterface(g.PrevID),
		NextID:          types.NullInt64FromInterface(g.NextID),
		HoldsReadyTasks: g.HoldsReadyTasks,
	}
}
