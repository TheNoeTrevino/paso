package postgres

import (
	"github.com/thenoetrevino/paso/internal/database/generated_postgres"
	"github.com/thenoetrevino/paso/internal/database/types"
)

func fromGeneratedAssignee(g generated_postgres.Assignee) types.Assignee {
	return types.Assignee{
		ID:        int64(g.ID),
		Name:      g.Name,
		CreatedAt: types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt: types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedTask(g generated_postgres.Task) types.Task {
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
		DueDate:      types.FromSQLNullTime(g.DueDate),
		Archived:     g.Archived,
	}
}

func fromGeneratedProject(g generated_postgres.Project) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.FromSQLNullString(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetAllProjectsRow(g generated_postgres.GetAllProjectsRow) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.FromSQLNullString(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetProjectByIDRow(g generated_postgres.GetProjectByIDRow) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.FromSQLNullString(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedGetProjectByGitBranchRow(g generated_postgres.GetProjectByGitBranchRow) types.Project {
	return types.Project{
		ID:          g.ID,
		Name:        g.Name,
		Description: types.FromSQLNullString(g.Description),
		GitBranch:   types.FromSQLNullString(g.GitBranch),
		CreatedAt:   types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:   types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedColumn(g generated_postgres.Column) types.Column {
	return types.Column{
		ID:                   g.ID,
		Name:                 g.Name,
		PrevID:               types.FromSQLNullInt64(g.PrevID),
		NextID:               types.FromSQLNullInt64(g.NextID),
		ProjectID:            g.ProjectID,
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedLabel(g generated_postgres.Label) types.Label {
	return types.Label{
		ID:        g.ID,
		Name:      g.Name,
		Color:     g.Color,
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedTaskComment(g generated_postgres.TaskComment) types.TaskComment {
	return types.TaskComment{
		ID:        g.ID,
		TaskID:    g.TaskID,
		Content:   g.Content,
		Author:    g.Author,
		CreatedAt: types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt: types.FromSQLNullTime(g.UpdatedAt),
	}
}

func fromGeneratedPriority(g generated_postgres.Priority) types.Priority {
	return types.Priority{
		ID:          g.ID,
		Description: g.Description,
		Color:       g.Color,
	}
}

func fromGeneratedRelationType(g generated_postgres.RelationType) types.RelationType {
	return types.RelationType{
		ID:         g.ID,
		PToCLabel:  g.PToCLabel,
		CToPLabel:  g.CToPLabel,
		Color:      g.Color,
		IsBlocking: g.IsBlocking,
	}
}

func fromGeneratedType(g generated_postgres.Type) types.Type {
	return types.Type{
		ID:          g.ID,
		Description: g.Description,
	}
}

func fromGeneratedGetChildTasksRow(g generated_postgres.GetChildTasksRow) types.GetChildTasksRow {
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

func fromGeneratedGetInProgressTaskDetailsRow(g generated_postgres.GetInProgressTaskDetailsRow) types.GetInProgressTaskDetailsRow {
	return types.GetInProgressTaskDetailsRow{
		ID:                  g.ID,
		TicketNumber:        types.FromSQLNullInt64(g.TicketNumber),
		Title:               g.Title,
		Description:         types.FromSQLNullString(g.Description),
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		CreatedAt:           types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:           types.FromSQLNullTime(g.UpdatedAt),
		Estimate:            types.FromSQLNullString(g.Estimate),
		DueDate:             types.FromSQLNullTime(g.DueDate),
		ColumnName:          g.ColumnName,
		ProjectName:         g.ProjectName,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.FromSQLNullInt32(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetInProgressTasksByProjectRow(g generated_postgres.GetInProgressTasksByProjectRow) types.GetInProgressTasksByProjectRow {
	return types.GetInProgressTasksByProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Description:  types.FromSQLNullString(g.Description),
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
	}
}

func fromGeneratedGetParentTasksRow(g generated_postgres.GetParentTasksRow) types.GetParentTasksRow {
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

func fromGeneratedGetReadyTaskSummariesByProjectRow(g generated_postgres.GetReadyTaskSummariesByProjectRow) types.GetReadyTaskSummariesByProjectRow {
	return types.GetReadyTaskSummariesByProjectRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.FromSQLNullString(g.Estimate),
		DueDate:             types.FromSQLNullTime(g.DueDate),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.FromSQLNullInt32(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTaskRow(g generated_postgres.GetTaskRow) types.GetTaskRow {
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

func fromGeneratedGetTaskAboveRow(g generated_postgres.GetTaskAboveRow) types.GetTaskAboveRow {
	return types.GetTaskAboveRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskBelowRow(g generated_postgres.GetTaskBelowRow) types.GetTaskBelowRow {
	return types.GetTaskBelowRow{
		ID:       g.ID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskDetailRow(g generated_postgres.GetTaskDetailRow) types.GetTaskDetailRow {
	return types.GetTaskDetailRow{
		ID:                  g.ID,
		Title:               g.Title,
		Description:         types.FromSQLNullString(g.Description),
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		TicketNumber:        types.FromSQLNullInt64(g.TicketNumber),
		CreatedAt:           types.FromSQLNullTime(g.CreatedAt),
		UpdatedAt:           types.FromSQLNullTime(g.UpdatedAt),
		Estimate:            types.FromSQLNullString(g.Estimate),
		DueDate:             types.FromSQLNullTime(g.DueDate),
		Archived:            g.Archived,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		ColumnName:          g.ColumnName,
		ProjectName:         g.ProjectName,
		AssigneeID:          types.FromSQLNullInt32(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTaskPositionRow(g generated_postgres.GetTaskPositionRow) types.GetTaskPositionRow {
	return types.GetTaskPositionRow{
		ColumnID: g.ColumnID,
		Position: g.Position,
	}
}

func fromGeneratedGetTaskReferencesForProjectRow(g generated_postgres.GetTaskReferencesForProjectRow) types.GetTaskReferencesForProjectRow {
	return types.GetTaskReferencesForProjectRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		Name:         g.Name,
	}
}

func fromGeneratedGetTaskRelationsForProjectRow(g generated_postgres.GetTaskRelationsForProjectRow) types.GetTaskRelationsForProjectRow {
	return types.GetTaskRelationsForProjectRow{
		ParentID:      g.ParentID,
		ChildID:       g.ChildID,
		RelationLabel: g.RelationLabel,
		RelationColor: g.RelationColor,
		IsBlocking:    g.IsBlocking,
	}
}

func fromGeneratedGetTaskSummariesByColumnRow(g generated_postgres.GetTaskSummariesByColumnRow) types.GetTaskSummariesByColumnRow {
	return types.GetTaskSummariesByColumnRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.FromSQLNullString(g.Estimate),
		DueDate:             types.FromSQLNullTime(g.DueDate),
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.FromSQLNullInt32(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
	}
}

func fromGeneratedGetTaskSummariesByProjectRow(g generated_postgres.GetTaskSummariesByProjectRow) types.GetTaskSummariesByProjectRow {
	return types.GetTaskSummariesByProjectRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.FromSQLNullString(g.Estimate),
		DueDate:             types.FromSQLNullTime(g.DueDate),
		Archived:            g.Archived,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.FromSQLNullInt32(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTaskSummariesWithFiltersRow(g generated_postgres.GetTaskSummariesWithFiltersRow) types.GetTaskSummariesWithFiltersRow {
	return types.GetTaskSummariesWithFiltersRow{
		ID:                  g.ID,
		Title:               g.Title,
		ColumnID:            g.ColumnID,
		Position:            g.Position,
		Estimate:            types.FromSQLNullString(g.Estimate),
		DueDate:             types.FromSQLNullTime(g.DueDate),
		Archived:            g.Archived,
		TypeDescription:     types.FromSQLNullString(g.TypeDescription),
		PriorityDescription: types.FromSQLNullString(g.PriorityDescription),
		PriorityColor:       types.FromSQLNullString(g.PriorityColor),
		AssigneeID:          types.FromSQLNullInt32(g.AssigneeID),
		AssigneeName:        types.FromSQLNullString(g.AssigneeName),
		LabelIds:            g.LabelIds,
		LabelNames:          g.LabelNames,
		LabelColors:         g.LabelColors,
		IsBlocked:           g.IsBlocked,
	}
}

func fromGeneratedGetTasksByColumnRow(g generated_postgres.GetTasksByColumnRow) types.GetTasksByColumnRow {
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

func fromGeneratedGetTasksForTreeRow(g generated_postgres.GetTasksForTreeRow) types.GetTasksForTreeRow {
	return types.GetTasksForTreeRow{
		ID:           g.ID,
		TicketNumber: types.FromSQLNullInt64(g.TicketNumber),
		Title:        g.Title,
		ColumnName:   g.ColumnName,
		ProjectName:  g.ProjectName,
		IsCompleted:  g.IsCompleted,
	}
}

func fromGeneratedGetColumnByIDRow(g generated_postgres.GetColumnByIDRow) types.GetColumnByIDRow {
	return types.GetColumnByIDRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.FromSQLNullInt64(g.PrevID),
		NextID:               types.FromSQLNullInt64(g.NextID),
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetColumnLinkedListInfoRow(g generated_postgres.GetColumnLinkedListInfoRow) types.GetColumnLinkedListInfoRow {
	return types.GetColumnLinkedListInfoRow{
		PrevID:    types.FromSQLNullInt64(g.PrevID),
		NextID:    types.FromSQLNullInt64(g.NextID),
		ProjectID: g.ProjectID,
	}
}

func fromGeneratedGetColumnsByProjectRow(g generated_postgres.GetColumnsByProjectRow) types.GetColumnsByProjectRow {
	return types.GetColumnsByProjectRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.FromSQLNullInt64(g.PrevID),
		NextID:               types.FromSQLNullInt64(g.NextID),
		HoldsReadyTasks:      g.HoldsReadyTasks,
		HoldsCompletedTasks:  g.HoldsCompletedTasks,
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetCompletedColumnByProjectRow(g generated_postgres.GetCompletedColumnByProjectRow) types.GetCompletedColumnByProjectRow {
	return types.GetCompletedColumnByProjectRow{
		ID:        g.ID,
		Name:      g.Name,
		ProjectID: g.ProjectID,
		PrevID:    types.FromSQLNullInt64(g.PrevID),
		NextID:    types.FromSQLNullInt64(g.NextID),
	}
}

func fromGeneratedGetInProgressColumnByProjectRow(g generated_postgres.GetInProgressColumnByProjectRow) types.GetInProgressColumnByProjectRow {
	return types.GetInProgressColumnByProjectRow{
		ID:                   g.ID,
		Name:                 g.Name,
		ProjectID:            g.ProjectID,
		PrevID:               types.FromSQLNullInt64(g.PrevID),
		NextID:               types.FromSQLNullInt64(g.NextID),
		HoldsInProgressTasks: g.HoldsInProgressTasks,
	}
}

func fromGeneratedGetReadyColumnByProjectRow(g generated_postgres.GetReadyColumnByProjectRow) types.GetReadyColumnByProjectRow {
	return types.GetReadyColumnByProjectRow{
		ID:              g.ID,
		Name:            g.Name,
		ProjectID:       g.ProjectID,
		PrevID:          types.FromSQLNullInt64(g.PrevID),
		NextID:          types.FromSQLNullInt64(g.NextID),
		HoldsReadyTasks: g.HoldsReadyTasks,
	}
}
