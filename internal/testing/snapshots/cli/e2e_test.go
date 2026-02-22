package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/assignee"
	"github.com/thenoetrevino/paso/internal/cli/column"
	"github.com/thenoetrevino/paso/internal/cli/label"
	"github.com/thenoetrevino/paso/internal/cli/project"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/git"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/testing/snapshots"
)

// fixedTimestamp is a deterministic timestamp used in snapshot tests to avoid
// flaky whitespace issues caused by variable-width formatted dates.
const fixedTimestamp = "2006-01-02 15:04:05"

func e2eHelper(t *testing.T) *snapshots.Helper {
	t.Helper()
	return snapshots.NewHelper(t, "testdata/e2e")
}

func cleanOutput(output string) string {
	return fixtures.StripANSI(output)
}

func TestGolden_E2E_Task(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	projectID := cli.CreateTestProject(t, db, "MyProject")
	todoCol := cli.GetColumnIDByName(t, db, projectID, "Todo")
	inProgCol := cli.GetColumnIDByName(t, db, projectID, "In Progress")
	doneCol := cli.GetColumnIDByName(t, db, projectID, "Done")
	cli.SetColumnHoldsReadyTasks(t, db, todoCol)
	cli.SetColumnHoldsInProgressTasks(t, db, inProgCol)
	cli.SetColumnHoldsCompletedTasks(t, db, doneCol)

	assigneeID := cli.CreateTestAssignee(t, db, "alice")
	_ = assigneeID

	t.Run("create", func(t *testing.T) {
		cmd := task.CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Fix login bug",
			"--project", fmt.Sprintf("%d", projectID),
			"--priority", "high",
			"--type", "task",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-create", cleanOutput(output))
	})

	t.Run("create-with-description", func(t *testing.T) {
		cmd := task.CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "Add dark mode",
			"--description", "Implement dark mode toggle in settings",
			"--project", fmt.Sprintf("%d", projectID),
			"--priority", "medium",
			"--type", "feature",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-create-with-description", cleanOutput(output))
	})

	// Seed some tasks for list/show tests
	task1 := cli.CreateTestTask(t, db, todoCol, "Write unit tests")
	task2 := cli.CreateTestTask(t, db, todoCol, "Update documentation")
	task3 := cli.CreateTestTask(t, db, inProgCol, "Refactor auth module")
	labelID := cli.CreateTestLabel(t, db, projectID, "bug", "#FF0000")
	cli.AttachLabelToTask(t, db, task1, labelID)
	cli.UpdateTaskDescription(t, db, task1, "Cover all edge cases")
	// Block task2 by task1
	cli.AddTaskSubtask(t, db, task1, task2, 2)

	// Pin timestamps on task1 so task-show snapshots are deterministic.
	cli.UpdateTaskFields(t, db, task1, map[string]any{
		"created_at": fixedTimestamp,
		"updated_at": fixedTimestamp,
	})

	t.Run("list", func(t *testing.T) {
		cmd := task.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-list", cleanOutput(output))
	})

	t.Run("show", func(t *testing.T) {
		cmd := task.ShowCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", task1),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-show", cleanOutput(output))
	})

	t.Run("update", func(t *testing.T) {
		cmd := task.UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", task1),
			"--title", "Write comprehensive unit tests",
			"--priority", "critical",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-update", cleanOutput(output))
	})

	t.Run("assign", func(t *testing.T) {
		cmd := task.AssignCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", task1),
			"alice",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-assign", cleanOutput(output))
	})

	t.Run("estimate", func(t *testing.T) {
		cmd := task.EstimateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", task1),
			"2d",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-estimate", cleanOutput(output))
	})

	t.Run("comment", func(t *testing.T) {
		cmd := task.CommentCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", task1),
			"--message", "This needs review before merging",
			"--author", "alice",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-comment", cleanOutput(output))
	})

	t.Run("ready-list", func(t *testing.T) {
		cmd := task.ReadyCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-ready-list", cleanOutput(output))
	})

	t.Run("blocked-list", func(t *testing.T) {
		cmd := task.BlockedCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-blocked-list", cleanOutput(output))
	})

	t.Run("in-progress-list", func(t *testing.T) {
		cmd := task.InProgressCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-in-progress-list", cleanOutput(output))
	})

	t.Run("move", func(t *testing.T) {
		cmd := task.MoveCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", task3),
			"Done",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-move", cleanOutput(output))
	})

	// Use a fresh task for done/to-ready/in-progress move tests
	taskForMoves := cli.CreateTestTask(t, db, todoCol, "Task for moves")

	t.Run("in-progress-move", func(t *testing.T) {
		cmd := task.InProgressCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskForMoves),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-in-progress-move", cleanOutput(output))
	})

	t.Run("to-ready-move", func(t *testing.T) {
		cmd := task.ReadyMoveCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskForMoves),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-to-ready", cleanOutput(output))
	})

	t.Run("done-move", func(t *testing.T) {
		cmd := task.DoneCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", taskForMoves),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-done", cleanOutput(output))
	})

	t.Run("link", func(t *testing.T) {
		taskA := cli.CreateTestTask(t, db, todoCol, "Parent task")
		taskB := cli.CreateTestTask(t, db, todoCol, "Child task")
		cmd := task.LinkCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--parent", fmt.Sprintf("%d", taskA),
			"--child", fmt.Sprintf("%d", taskB),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-link", cleanOutput(output))
	})

	// Create a second project for switch test (needs ready column configured)
	project2ID := cli.CreateTestProject(t, db, "OtherProject")
	proj2TodoCol := cli.GetColumnIDByName(t, db, project2ID, "Todo")
	cli.SetColumnHoldsReadyTasks(t, db, proj2TodoCol)
	switchTask := cli.CreateTestTask(t, db, todoCol, "Task to switch")

	t.Run("switch", func(t *testing.T) {
		cmd := task.SwitchCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", switchTask),
			fmt.Sprintf("%d", project2ID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-switch", cleanOutput(output))
	})

	t.Run("delete", func(t *testing.T) {
		deleteTask := cli.CreateTestTask(t, db, todoCol, "Task to delete")
		cmd := task.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", deleteTask),
			"--force",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("task-delete", cleanOutput(output))
	})
}

func TestGolden_E2E_Project(t *testing.T) {
	t.Parallel()
	db, app, gitMock := cli.SetupCLITestWithGit(t)
	gitMock.Info = git.GitInfo{
		IsRepo:        true,
		CurrentBranch: "test/e2e-snapshots",
		HasCommits:    true,
	}

	t.Run("create", func(t *testing.T) {
		cmd := project.CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "WebApp",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("project-create", cleanOutput(output))
	})

	t.Run("create-with-description", func(t *testing.T) {
		cmd := project.CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--title", "MobileApp",
			"--description", "React Native mobile application",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("project-create-with-description", cleanOutput(output))
	})

	// Seed some projects for list
	proj1 := cli.CreateTestProject(t, db, "Alpha")
	cli.CreateTestProject(t, db, "Beta")

	t.Run("list", func(t *testing.T) {
		cmd := project.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		require.NoError(t, err)
		e2eHelper(t).Compare("project-list", cleanOutput(output))
	})

	// Seed tasks for tree view
	todoCol := cli.GetColumnIDByName(t, db, proj1, "Todo")
	parentTask := cli.CreateTestTask(t, db, todoCol, "Epic: Authentication")
	childTask := cli.CreateTestTask(t, db, todoCol, "Implement JWT")
	cli.AddTaskSubtask(t, db, parentTask, childTask, 1)

	t.Run("tree", func(t *testing.T) {
		cmd := project.TreeCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", proj1),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("project-tree", cleanOutput(output))
	})

	t.Run("delete", func(t *testing.T) {
		delProj := cli.CreateTestProject(t, db, "ToDelete")
		cmd := project.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", delProj),
			"--force",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("project-delete", cleanOutput(output))
	})
}

func TestGolden_E2E_Column(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	projectID := cli.CreateTestProject(t, db, "ColProject")

	t.Run("create", func(t *testing.T) {
		cmd := column.CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "Review",
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("column-create", cleanOutput(output))
	})

	t.Run("list", func(t *testing.T) {
		cmd := column.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("column-list", cleanOutput(output))
	})

	t.Run("update", func(t *testing.T) {
		colID := cli.GetColumnIDByName(t, db, projectID, "Todo")
		cmd := column.UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", colID),
			"--name", "Backlog",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("column-update", cleanOutput(output))
	})

	t.Run("delete", func(t *testing.T) {
		delCol := cli.CreateTestColumn(t, db, projectID, "TempColumn")
		cmd := column.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", delCol),
			"--force",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("column-delete", cleanOutput(output))
	})
}

func TestGolden_E2E_Label(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	projectID := cli.CreateTestProject(t, db, "LabelProject")
	todoCol := cli.GetColumnIDByName(t, db, projectID, "Todo")

	t.Run("create", func(t *testing.T) {
		cmd := label.CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "enhancement",
			"--color", "#00FF00",
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("label-create", cleanOutput(output))
	})

	// Seed labels for list test
	cli.CreateTestLabel(t, db, projectID, "bug", "#FF0000")
	cli.CreateTestLabel(t, db, projectID, "docs", "#0000FF")

	t.Run("list", func(t *testing.T) {
		cmd := label.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--project", fmt.Sprintf("%d", projectID),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("label-list", cleanOutput(output))
	})

	t.Run("update", func(t *testing.T) {
		updateLabel := cli.CreateTestLabel(t, db, projectID, "wip", "#999999")
		cmd := label.UpdateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", updateLabel),
			"--name", "in-review",
			"--color", "#FFA500",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("label-update", cleanOutput(output))
	})

	// Attach/detach tests
	taskID := cli.CreateTestTask(t, db, todoCol, "Labeled task")
	attachLabel := cli.CreateTestLabel(t, db, projectID, "critical", "#FF0000")

	t.Run("attach", func(t *testing.T) {
		cmd := label.AttachCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", attachLabel),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("label-attach", cleanOutput(output))
	})

	t.Run("detach", func(t *testing.T) {
		cmd := label.DetachCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--task", fmt.Sprintf("%d", taskID),
			"--label", fmt.Sprintf("%d", attachLabel),
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("label-detach", cleanOutput(output))
	})

	t.Run("delete", func(t *testing.T) {
		delLabel := cli.CreateTestLabel(t, db, projectID, "temp", "#AAAAAA")
		cmd := label.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			fmt.Sprintf("%d", delLabel),
			"--force",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("label-delete", cleanOutput(output))
	})
}

func TestGolden_E2E_Assignee(t *testing.T) {
	t.Parallel()
	db, app := cli.SetupCLITest(t)

	t.Run("create", func(t *testing.T) {
		cmd := assignee.CreateCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--name", "bob",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("assignee-create", cleanOutput(output))
	})

	// Seed assignees for list
	cli.CreateTestAssignee(t, db, "charlie")
	cli.CreateTestAssignee(t, db, "diana")

	t.Run("list", func(t *testing.T) {
		cmd := assignee.ListCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{})
		require.NoError(t, err)
		e2eHelper(t).Compare("assignee-list", cleanOutput(output))
	})

	t.Run("delete", func(t *testing.T) {
		delAssignee := cli.CreateTestAssignee(t, db, "temp-user")
		cmd := assignee.DeleteCmd()
		output, err := cli.ExecuteCLICommand(t, app, cmd, []string{
			"--id", fmt.Sprintf("%d", delAssignee),
			"--force",
		})
		require.NoError(t, err)
		e2eHelper(t).Compare("assignee-delete", cleanOutput(output))
	})
}
