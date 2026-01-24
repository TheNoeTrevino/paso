// Package commands provides Bubbletea commands for the TUI application.
package commands

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/task"
)

// AdjacentTasks contains the IDs of tasks adjacent to the currently selected task.
// Used for prefetching task details to enable instant navigation in the detail panel.
type AdjacentTasks struct {
	Current int // The selected task (required)
	Above   int // Task above in same column (0 if none)
	Below   int // Task below in same column (0 if none)
	Left    int // Task in column to the left (0 if none)
	Right   int // Task in column to the right (0 if none)
}

// TaskDetailsPrefetchedMsg is sent when task details have been prefetched.
// Contains the fetched details and any errors that occurred during fetching.
type TaskDetailsPrefetchedMsg struct {
	Details   map[int]*models.TaskDetail
	CurrentID int
	Errors    map[int]error
}

// FetchTaskDetailsCmd creates a Bubbletea command that fetches task details
// for the current task and its adjacent tasks in parallel.
//
// Parameters:
//   - svc: The task service for fetching task details
//   - adjacent: The IDs of adjacent tasks to prefetch
//   - cachedIDs: IDs of tasks that are already cached (will be skipped)
//
// Returns a tea.Cmd that when executed will fetch all non-cached task details
// concurrently and return a TaskDetailsPrefetchedMsg.
func FetchTaskDetailsCmd(svc task.Service, adjacent AdjacentTasks, cachedIDs []int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Build set of cached IDs for O(1) lookup
		cached := make(map[int]bool)
		for _, id := range cachedIDs {
			cached[id] = true
		}

		// Collect task IDs to fetch, filtering out cached ones
		idsToFetch := collectIDsToFetch(adjacent, cached)

		if len(idsToFetch) == 0 {
			return TaskDetailsPrefetchedMsg{
				Details:   make(map[int]*models.TaskDetail),
				CurrentID: adjacent.Current,
				Errors:    make(map[int]error),
			}
		}

		// Fetch all tasks in parallel
		details, errors := fetchTasksInParallel(ctx, svc, idsToFetch)

		return TaskDetailsPrefetchedMsg{
			Details:   details,
			CurrentID: adjacent.Current,
			Errors:    errors,
		}
	}
}

// collectIDsToFetch returns a slice of task IDs that need to be fetched,
// excluding already cached IDs and zero values.
func collectIDsToFetch(adjacent AdjacentTasks, cached map[int]bool) []int {
	candidates := []int{
		adjacent.Current,
		adjacent.Above,
		adjacent.Below,
		adjacent.Left,
		adjacent.Right,
	}

	var ids []int
	for _, id := range candidates {
		if id > 0 && !cached[id] {
			ids = append(ids, id)
		}
	}
	return ids
}

// fetchTasksInParallel fetches multiple task details concurrently using goroutines.
// Uses sync.WaitGroup for coordination and sync.Mutex to protect result maps.
func fetchTasksInParallel(ctx context.Context, svc task.Service, ids []int) (map[int]*models.TaskDetail, map[int]error) {
	details := make(map[int]*models.TaskDetail)
	errors := make(map[int]error)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, id := range ids {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			detail, err := fetchWithRetry(ctx, svc, taskID, 3)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors[taskID] = err
			} else {
				details[taskID] = detail
			}
		}(id)
	}

	wg.Wait()
	return details, errors
}

// fetchWithRetry attempts to fetch a task detail with exponential backoff retry.
// Retries on failure with delays of 50ms, 100ms for up to maxRetries attempts.
// Respects context cancellation between retries.
func fetchWithRetry(ctx context.Context, svc task.Service, taskID int, maxRetries int) (*models.TaskDetail, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		detail, err := svc.GetTaskDetail(ctx, taskID)
		if err == nil {
			return detail, nil
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt < maxRetries-1 {
			// Exponential backoff: 50ms, 100ms
			backoff := time.Duration(50<<attempt) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return nil, fmt.Errorf("fetch task %d failed after %d attempts: %w", taskID, maxRetries, lastErr)
}
