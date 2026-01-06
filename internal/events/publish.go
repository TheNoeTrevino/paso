package events

import (
	"log/slog"
	"time"
)

// PublishWithRetry attempts to publish an event with retry logic.
// It makes up to maxRetries attempts with exponential backoff.
// Returns the error from the final attempt if all retries fail.
func PublishWithRetry(client EventPublisher, event Event, maxRetries int) error {
	if client == nil {
		return nil // Silently skip if no client (e.g., in tests or non-daemon mode)
	}

	var lastErr error
	baseDelay := 50 * time.Millisecond

	for attempt := range maxRetries {
		err := client.SendEvent(event)
		if err == nil {
			if attempt > 0 {
				slog.Debug("event published after retry",
					"attempt", attempt+1,
					"event_type", event.Type,
					"project_id", event.ProjectID)
			}
			return nil
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt < maxRetries-1 {
			delay := baseDelay * (1 << attempt)
			slog.Debug("event publish failed, retrying",
				"attempt", attempt+1,
				"max_retries", maxRetries,
				"retry_delay", delay,
				"error", err)
			time.Sleep(delay)
		}
	}

	slog.Warn("event publish failed after all retries",
		"attempts", maxRetries,
		"event_type", event.Type,
		"project_id", event.ProjectID,
		"error", lastErr)

	return lastErr
}
