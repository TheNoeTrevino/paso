// Package appcontext provides context keys for dependency injection.
package appcontext

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

// AppKey is the context key used for dependency injection of *app.App.
// CLI commands check for this key to support both production and test modes.
const AppKey ContextKey = "testApp"
