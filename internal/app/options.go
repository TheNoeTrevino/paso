package app

import (
	"log/slog"

	"github.com/thenoetrevino/paso/internal/events"
)

// Option is a functional option for configuring App initialization
type Option func(*appConfig)

// appConfig holds the configuration for App initialization
type appConfig struct {
	eventClient events.EventPublisher
	logger      *slog.Logger
}

// WithEventPublisher sets the event publisher for the application
func WithEventPublisher(ec events.EventPublisher) Option {
	return func(cfg *appConfig) {
		cfg.eventClient = ec
	}
}

// WithLogger sets the logger for the application
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *appConfig) {
		cfg.logger = logger
	}
}
