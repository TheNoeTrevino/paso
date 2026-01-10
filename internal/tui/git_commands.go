package tui

import (
	"context"
	"errors"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/git"
)

func loadGitDataForProjectFormCmd(ctx context.Context, cache *git.Cache, forEdit bool) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			slog.Info("git data fetch cancelled before start", "forEdit", forEdit)
			return gitInfoError{
				err:     ctx.Err(),
				forEdit: forEdit,
			}
		default:
		}

		workDir, err := os.Getwd()
		if err != nil {
			slog.Warn("failed to get working directory for cache", "error", err)
			workDir = ""
		}

		if workDir != "" {
			cachedInfo, cachedBranches, hit := cache.Get(ctx, workDir)
			if hit {
				return gitInfoFetched{
					gitInfo:     cachedInfo,
					gitBranches: cachedBranches,
					forEdit:     forEdit,
				}
			}
		}

		gitInfo := git.DetectGitInfo(ctx)

		var gitBranches []git.BranchInfo
		if gitInfo.IsRepo {
			branches, err := git.ListBranches(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					slog.Info("git branch list cancelled", "forEdit", forEdit)
				} else if errors.Is(err, context.DeadlineExceeded) {
					slog.Warn("git branch list timed out", "forEdit", forEdit)
				} else {
					slog.Warn("failed to list git branches", "error", err, "forEdit", forEdit)
				}
				return gitInfoError{
					err:     err,
					forEdit: forEdit,
				}
			}
			gitBranches = branches
		}

		if workDir != "" {
			cache.Set(workDir, gitInfo, gitBranches)
		}

		return gitInfoFetched{
			gitInfo:     gitInfo,
			gitBranches: gitBranches,
			forEdit:     forEdit,
		}
	}
}
