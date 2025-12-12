package database

import (
	"context"
	"log"
	"testing"

	_ "modernc.org/sqlite"
)

// TestLabelPersistence tests that labels are properly saved and retrieved
func TestLabelPersistence(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create a label (projectID 1 is created by migrations)
	label, err := repo.CreateLabel(context.Background(), 1, "Bug", "#FF0000")
	if err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	if label.ID == 0 {
		t.Error("Label should have a valid ID")
	}
	if label.Name != "Bug" {
		t.Errorf("Expected label name 'Bug', got '%s'", label.Name)
	}
	if label.Color != "#FF0000" {
		t.Errorf("Expected label color '#FF0000', got '%s'", label.Color)
	}
	if label.ProjectID != 1 {
		t.Errorf("Expected label project ID 1, got %d", label.ProjectID)
	}

	// Retrieve all labels
	labels, err := repo.GetLabelsByProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(labels))
	}
	if labels[0].Name != "Bug" {
		t.Errorf("Retrieved label has wrong name: %s", labels[0].Name)
	}
}

// TestTaskLabelAssociation tests the many-to-many relationship between tasks and labels
func TestTaskLabelAssociation(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create column, task, and labels
	col, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)
	task, _ := repo.CreateTask(context.Background(), "Test Task", "Description", col.ID, 0)
	label1, _ := repo.CreateLabel(context.Background(), 1, "Bug", "#FF0000")
	label2, _ := repo.CreateLabel(context.Background(), 1, "Feature", "#00FF00")

	// Associate labels with task
	err := repo.SetTaskLabels(context.Background(), task.ID, []int{label1.ID, label2.ID})
	if err != nil {
		t.Fatalf("Failed to set task labels: %v", err)
	}

	// Retrieve labels for task
	labels, err := repo.GetLabelsForTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to get labels for task: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("Expected 2 labels, got %d", len(labels))
	}

	// Verify task summary includes labels
	summaries, err := repo.GetTaskSummariesByColumn(context.Background(), col.ID)
	if err != nil {
		t.Fatalf("Failed to get task summaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("Expected 1 summary, got %d", len(summaries))
	}
	if len(summaries[0].Labels) != 2 {
		t.Errorf("Expected summary to have 2 labels, got %d", len(summaries[0].Labels))
	}

	// Verify task detail includes labels
	detail, err := repo.GetTaskDetail(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to get task detail: %v", err)
	}
	if len(detail.Labels) != 2 {
		t.Errorf("Expected detail to have 2 labels, got %d", len(detail.Labels))
	}
}

// TestSetTaskLabelsReplaces tests that SetTaskLabels replaces existing labels
func TestSetTaskLabelsReplaces(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Create column, task, and labels
	col, _ := repo.CreateColumn(context.Background(), "Todo", 1, nil)
	task, _ := repo.CreateTask(context.Background(), "Test Task", "", col.ID, 0)
	label1, _ := repo.CreateLabel(context.Background(), 1, "Bug", "#FF0000")
	label2, _ := repo.CreateLabel(context.Background(), 1, "Feature", "#00FF00")
	label3, _ := repo.CreateLabel(context.Background(), 1, "Enhancement", "#0000FF")

	// Set initial labels
	if err := repo.SetTaskLabels(context.Background(), task.ID, []int{label1.ID, label2.ID}); err != nil {
		t.Fatalf("Failed to set initial labels: %v", err)
	}

	// Replace with different labels
	err := repo.SetTaskLabels(context.Background(), task.ID, []int{label3.ID})
	if err != nil {
		t.Fatalf("Failed to replace task labels: %v", err)
	}

	// Verify only the new label is associated
	labels, err := repo.GetLabelsForTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("Expected 1 label after replacement, got %d", len(labels))
	}
	if labels[0].ID != label3.ID {
		t.Errorf("Expected label ID %d, got %d", label3.ID, labels[0].ID)
	}
}

// TestProjectSpecificLabels tests that labels are properly scoped to projects
func TestProjectSpecificLabels(t *testing.T) {
	db := setupTestDB(t)
	defer func() { if err := db.Close(); err != nil { log.Printf("failed to close database: %v", err) } }()
	repo := NewRepository(db)

	// Project 1 is already created by migrations
	// Create a second project
	project2, err := repo.CreateProject(context.Background(), "Project 2", "Second project")
	if err != nil {
		t.Fatalf("Failed to create project 2: %v", err)
	}

	// Create labels for project 1
	label1, err := repo.CreateLabel(context.Background(), 1, "Bug", "#FF0000")
	if err != nil {
		t.Fatalf("Failed to create label for project 1: %v", err)
	}
	if label1.ProjectID != 1 {
		t.Errorf("Expected project ID 1, got %d", label1.ProjectID)
	}

	// Create labels for project 2
	label2, err := repo.CreateLabel(context.Background(), project2.ID, "Feature", "#00FF00")
	if err != nil {
		t.Fatalf("Failed to create label for project 2: %v", err)
	}
	if label2.ProjectID != project2.ID {
		t.Errorf("Expected project ID %d, got %d", project2.ID, label2.ProjectID)
	}

	// GetLabelsByProject should return only project-specific labels
	labelsP1, err := repo.GetLabelsByProject(context.Background(), 1)
	if err != nil {
		t.Fatalf("Failed to get labels for project 1: %v", err)
	}
	if len(labelsP1) != 1 {
		t.Errorf("Expected 1 label for project 1, got %d", len(labelsP1))
	}
	if labelsP1[0].Name != "Bug" {
		t.Errorf("Expected label 'Bug', got '%s'", labelsP1[0].Name)
	}

	labelsP2, err := repo.GetLabelsByProject(context.Background(), project2.ID)
	if err != nil {
		t.Fatalf("Failed to get labels for project 2: %v", err)
	}
	if len(labelsP2) != 1 {
		t.Errorf("Expected 1 label for project 2, got %d", len(labelsP2))
	}
	if labelsP2[0].Name != "Feature" {
		t.Errorf("Expected label 'Feature', got '%s'", labelsP2[0].Name)
	}
}
