package commands

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
	"go.uber.org/goleak"
)

func defaultPrefetchMock() *mocks.MockTaskService {
	mock := mocks.NewMockTaskService()
	mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
		return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
	}
	return mock
}

func TestFetchTaskDetailsCmd_CacheHitScenarios(t *testing.T) {
	t.Parallel()

	t.Run("all adjacent tasks cached returns empty Details map", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{
			Current: 1,
			Above:   2,
			Below:   3,
			Left:    4,
			Right:   5,
		}
		// All IDs are cached
		cachedIDs := []int{1, 2, 3, 4, 5}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		msg := cmd()

		result, ok := msg.(TaskDetailsPrefetchedMsg)
		require.True(t, ok, "expected TaskDetailsPrefetchedMsg")

		assert.Empty(t, result.Details, "expected empty Details map when all cached")
		assert.Empty(t, result.Errors, "expected no errors")
		assert.Equal(t, 1, result.CurrentID, "CurrentID should match adjacent.Current")
		assert.Equal(t, 0, mock.CallCount("GetTaskDetail"), "no service calls should be made")
	})

	t.Run("current cached but adjacent not cached fetches adjacent", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{
			Current: 1,
			Above:   2,
			Below:   3,
			Left:    0, // no left
			Right:   0, // no right
		}
		// Only current is cached
		cachedIDs := []int{1}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		msg := cmd()

		result, ok := msg.(TaskDetailsPrefetchedMsg)
		require.True(t, ok, "expected TaskDetailsPrefetchedMsg")

		// Should have fetched tasks 2 and 3
		assert.Len(t, result.Details, 2, "expected 2 fetched tasks")
		assert.Contains(t, result.Details, 2, "expected task 2 to be fetched")
		assert.Contains(t, result.Details, 3, "expected task 3 to be fetched")
		assert.Empty(t, result.Errors)
	})

	t.Run("partial cache hit only fetches uncached IDs", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{
			Current: 1,
			Above:   2,
			Below:   3,
			Left:    4,
			Right:   5,
		}
		// Tasks 1, 2, and 4 are cached
		cachedIDs := []int{1, 2, 4}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		msg := cmd()

		result, ok := msg.(TaskDetailsPrefetchedMsg)
		require.True(t, ok, "expected TaskDetailsPrefetchedMsg")

		// Should only fetch 3 and 5
		assert.Len(t, result.Details, 2, "expected 2 fetched tasks")
		assert.Contains(t, result.Details, 3, "expected task 3 to be fetched")
		assert.Contains(t, result.Details, 5, "expected task 5 to be fetched")
		assert.NotContains(t, result.Details, 1, "task 1 was cached")
		assert.NotContains(t, result.Details, 2, "task 2 was cached")
		assert.NotContains(t, result.Details, 4, "task 4 was cached")
	})

	t.Run("zero IDs are ignored", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{
			Current: 1,
			Above:   0, // no task above
			Below:   2,
			Left:    0, // no task left
			Right:   0, // no task right
		}
		cachedIDs := []int{} // nothing cached

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		msg := cmd()

		result, ok := msg.(TaskDetailsPrefetchedMsg)
		require.True(t, ok, "expected TaskDetailsPrefetchedMsg")

		// Should only fetch 1 and 2 (zero IDs are filtered out)
		assert.Len(t, result.Details, 2, "expected 2 fetched tasks")
		assert.Contains(t, result.Details, 1)
		assert.Contains(t, result.Details, 2)
	})

	t.Run("empty adjacent returns current only", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{
			Current: 10,
			Above:   0,
			Below:   0,
			Left:    0,
			Right:   0,
		}
		cachedIDs := []int{}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		msg := cmd()

		result, ok := msg.(TaskDetailsPrefetchedMsg)
		require.True(t, ok, "expected TaskDetailsPrefetchedMsg")

		assert.Len(t, result.Details, 1, "expected only current to be fetched")
		assert.Contains(t, result.Details, 10)
	})
}

func TestFetchWithRetry_RetryLogic(t *testing.T) {
	t.Parallel()

	t.Run("succeeds on first attempt", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			if taskID == 1 {
				return &models.TaskDetail{ID: 1, Title: "Test Task"}, nil
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ctx := context.Background()
		detail, err := fetchWithRetry(ctx, mock, 1, 3)

		require.NoError(t, err)
		assert.Equal(t, 1, detail.ID)
		assert.Equal(t, "Test Task", detail.Title)
		assert.Equal(t, 1, mock.CallCount("GetTaskDetail"), "should succeed on first attempt")
	})

	t.Run("succeeds after N-1 failures", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		var mu sync.Mutex
		failuresLeft := map[int]int{1: 2}
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			mu.Lock()
			if f, ok := failuresLeft[taskID]; ok && f > 0 {
				failuresLeft[taskID]--
				mu.Unlock()
				return nil, errors.New("simulated transient error")
			}
			mu.Unlock()
			if taskID == 1 {
				return &models.TaskDetail{ID: 1, Title: "Test Task"}, nil
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ctx := context.Background()
		detail, err := fetchWithRetry(ctx, mock, 1, 3)

		require.NoError(t, err)
		assert.Equal(t, 1, detail.ID)
		// Should have called 3 times (2 failures + 1 success)
		assert.Equal(t, 3, mock.CallCount("GetTaskDetail"), "should retry and eventually succeed")
	})

	t.Run("fails after max retries exhausted", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		taskErrors := map[int]error{1: errors.New("permanent error")}
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			if err, ok := taskErrors[taskID]; ok {
				return nil, err
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ctx := context.Background()
		_, err := fetchWithRetry(ctx, mock, 1, 3)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "permanent error")
		// Should have tried exactly 3 times
		assert.Equal(t, 3, mock.CallCount("GetTaskDetail"), "should try exactly maxRetries times")
	})

	t.Run("respects maxRetries parameter", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			if taskID == 1 {
				return nil, errors.New("error")
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ctx := context.Background()
		_, err := fetchWithRetry(ctx, mock, 1, 5)

		require.Error(t, err)
		assert.Equal(t, 5, mock.CallCount("GetTaskDetail"), "should try exactly 5 times with maxRetries=5")
	})

	t.Run("succeeds on last retry attempt", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		var mu sync.Mutex
		failuresLeft := map[int]int{1: 2}
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			mu.Lock()
			if f, ok := failuresLeft[taskID]; ok && f > 0 {
				failuresLeft[taskID]--
				mu.Unlock()
				return nil, errors.New("simulated transient error")
			}
			mu.Unlock()
			if taskID == 1 {
				return &models.TaskDetail{ID: 1, Title: "Success"}, nil
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ctx := context.Background()
		detail, err := fetchWithRetry(ctx, mock, 1, 3)

		require.NoError(t, err)
		assert.Equal(t, "Success", detail.Title)
		assert.Equal(t, 3, mock.CallCount("GetTaskDetail"))
	})
}

func TestFetchWithRetry_ContextTimeout(t *testing.T) {
	t.Parallel()

	t.Run("returns context error when cancelled before first attempt", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := fetchWithRetry(ctx, mock, 1, 3)

		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("returns context error when cancelled during retry backoff", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			if taskID == 1 {
				return nil, errors.New("transient")
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
		defer cancel()

		_, err := fetchWithRetry(ctx, mock, 1, 3)

		require.Error(t, err)
		// Should be cancelled during backoff wait (50ms)
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
			"expected context error, got: %v", err)
	})

	t.Run("returns context error when cancelled mid-operation", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		mock.GetTaskDetailFunc = func(ctx context.Context, taskID int) (*models.TaskDetail, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		_, err := fetchWithRetry(ctx, mock, 1, 3)

		require.Error(t, err)
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
			"expected context error, got: %v", err)
	})
}

func TestFetchTaskDetailsCmd_ContextHandling(t *testing.T) {
	t.Parallel()

	t.Run("completes successfully with valid context", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{Current: 1, Above: 2, Below: 3}
		cachedIDs := []int{}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		msg := cmd()

		result, ok := msg.(TaskDetailsPrefetchedMsg)
		require.True(t, ok)
		assert.Len(t, result.Details, 3)
		assert.Empty(t, result.Errors)
	})

	t.Run("handles service errors gracefully", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		taskErrors := map[int]error{2: errors.New("service unavailable")}
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			if err, ok := taskErrors[taskID]; ok {
				return nil, err
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		adjacent := AdjacentTasks{Current: 1, Above: 2, Below: 3}
		cachedIDs := []int{}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		msg := cmd()

		result, ok := msg.(TaskDetailsPrefetchedMsg)
		require.True(t, ok)

		// Task 1 and 3 should succeed, task 2 should have error
		assert.Contains(t, result.Details, 1)
		assert.Contains(t, result.Details, 3)
		assert.Contains(t, result.Errors, 2)
		assert.Contains(t, result.Errors[2].Error(), "service unavailable")
	})
}

func TestFetchTaskDetailsCmd_NoGoroutineLeaks(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("successful completion", func(t *testing.T) {
		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{Current: 1, Above: 2, Below: 3, Left: 4, Right: 5}
		cachedIDs := []int{}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		_ = cmd()
	})

	t.Run("all cached", func(t *testing.T) {
		mock := defaultPrefetchMock()
		adjacent := AdjacentTasks{Current: 1, Above: 2, Below: 3}
		cachedIDs := []int{1, 2, 3}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		_ = cmd()
	})

	t.Run("partial errors", func(t *testing.T) {
		mock := mocks.NewMockTaskService()
		taskErrors := map[int]error{2: errors.New("error"), 4: errors.New("error")}
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			if err, ok := taskErrors[taskID]; ok {
				return nil, err
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		adjacent := AdjacentTasks{Current: 1, Above: 2, Below: 3, Left: 4, Right: 5}
		cachedIDs := []int{}

		cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
		_ = cmd()
	})
}

func TestFetchTasksInParallel(t *testing.T) {
	t.Parallel()

	t.Run("fetches all tasks concurrently", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		mock.GetTaskDetailFunc = func(ctx context.Context, taskID int) (*models.TaskDetail, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ids := []int{1, 2, 3, 4, 5}
		ctx := context.Background()

		start := time.Now()
		details, errs := fetchTasksInParallel(ctx, mock, ids)
		elapsed := time.Since(start)

		assert.Len(t, details, 5)
		assert.Empty(t, errs)
		// If sequential, would take 50ms+. Parallel should be much faster.
		assert.Less(t, elapsed, 40*time.Millisecond,
			"parallel execution should be faster than sequential")
	})

	t.Run("collects errors from failed fetches", func(t *testing.T) {
		t.Parallel()

		mock := mocks.NewMockTaskService()
		taskErrors := map[int]error{2: errors.New("error 2"), 4: errors.New("error 4")}
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			if err, ok := taskErrors[taskID]; ok {
				return nil, err
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		ids := []int{1, 2, 3, 4, 5}
		ctx := context.Background()

		details, errs := fetchTasksInParallel(ctx, mock, ids)

		assert.Len(t, details, 3, "should have 3 successful fetches")
		assert.Len(t, errs, 2, "should have 2 errors")
		assert.Contains(t, errs, 2)
		assert.Contains(t, errs, 4)
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		ids := []int{}
		ctx := context.Background()

		details, errs := fetchTasksInParallel(ctx, mock, ids)

		assert.Empty(t, details)
		assert.Empty(t, errs)
		assert.Equal(t, 0, mock.CallCount("GetTaskDetail"))
	})

	t.Run("handles single task", func(t *testing.T) {
		t.Parallel()

		mock := defaultPrefetchMock()
		ids := []int{42}
		ctx := context.Background()

		details, errs := fetchTasksInParallel(ctx, mock, ids)

		assert.Len(t, details, 1)
		assert.Contains(t, details, 42)
		assert.Empty(t, errs)
	})
}

func TestCollectIDsToFetch(t *testing.T) {
	t.Parallel()

	t.Run("returns all non-zero uncached IDs", func(t *testing.T) {
		t.Parallel()

		adjacent := AdjacentTasks{
			Current: 1,
			Above:   2,
			Below:   3,
			Left:    4,
			Right:   5,
		}
		cached := map[int]bool{}

		ids := collectIDsToFetch(adjacent, cached)

		assert.Len(t, ids, 5)
		assert.Contains(t, ids, 1)
		assert.Contains(t, ids, 2)
		assert.Contains(t, ids, 3)
		assert.Contains(t, ids, 4)
		assert.Contains(t, ids, 5)
	})

	t.Run("excludes cached IDs", func(t *testing.T) {
		t.Parallel()

		adjacent := AdjacentTasks{
			Current: 1,
			Above:   2,
			Below:   3,
			Left:    4,
			Right:   5,
		}
		cached := map[int]bool{1: true, 3: true, 5: true}

		ids := collectIDsToFetch(adjacent, cached)

		assert.Len(t, ids, 2)
		assert.Contains(t, ids, 2)
		assert.Contains(t, ids, 4)
		assert.NotContains(t, ids, 1)
		assert.NotContains(t, ids, 3)
		assert.NotContains(t, ids, 5)
	})

	t.Run("excludes zero IDs", func(t *testing.T) {
		t.Parallel()

		adjacent := AdjacentTasks{
			Current: 1,
			Above:   0,
			Below:   2,
			Left:    0,
			Right:   0,
		}
		cached := map[int]bool{}

		ids := collectIDsToFetch(adjacent, cached)

		assert.Len(t, ids, 2)
		assert.Contains(t, ids, 1)
		assert.Contains(t, ids, 2)
	})

	t.Run("returns empty when all cached", func(t *testing.T) {
		t.Parallel()

		adjacent := AdjacentTasks{
			Current: 1,
			Above:   2,
			Below:   3,
			Left:    0,
			Right:   0,
		}
		cached := map[int]bool{1: true, 2: true, 3: true}

		ids := collectIDsToFetch(adjacent, cached)

		assert.Empty(t, ids)
	})
}

func TestFetchTaskDetailsCmd_ConcurrencySafety(t *testing.T) {
	t.Parallel()

	t.Run("results map is thread-safe", func(t *testing.T) {
		t.Parallel()

		var callCount atomic.Int64
		mock := mocks.NewMockTaskService()
		mock.GetTaskDetailFunc = func(_ context.Context, taskID int) (*models.TaskDetail, error) {
			n := callCount.Add(1)
			// Add small variable delay based on call count (0-4ms)
			waitTime := time.Duration(n%5) * time.Millisecond
			if waitTime > 0 {
				time.Sleep(waitTime)
			}
			return &models.TaskDetail{ID: taskID, Title: "Mock Task"}, nil
		}

		adjacent := AdjacentTasks{Current: 1, Above: 2, Below: 3, Left: 4, Right: 5}
		cachedIDs := []int{}

		// Run multiple times to catch potential race conditions
		for i := 0; i < 10; i++ {
			cmd := FetchTaskDetailsCmd(mock, adjacent, cachedIDs)
			msg := cmd()

			result, ok := msg.(TaskDetailsPrefetchedMsg)
			require.True(t, ok)
			assert.Len(t, result.Details, 5)
		}
	})
}
