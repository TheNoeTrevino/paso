# Log - Terminal Logging Expert

Expert in Charm Log, a minimal and colorful logging library designed for terminal applications.

## When to Use This Skill

Use this skill when:
- Adding structured logging to TUI applications
- Debugging Bubble Tea programs
- Creating colorful, human-readable logs
- Logging to files or external services
- Building CLI tools that need logging
- Implementing log levels (debug, info, warn, error)

## Core Concepts

Charm Log provides leveled, structured, and colorful logging optimized for terminal output. It's built on top of Lip Gloss for styling.

## Basic Usage

```go
import "github.com/charmbracelet/log"

// Package-level logger
log.Debug("Cookie 🍪")  // Won't print (default level is Info)
log.Info("Hello World!")
log.Warn("Warning message")
log.Error("Error occurred", "err", err)
log.Fatal("Fatal error")  // Exits with os.Exit(1)

// Print without level
log.Print("No level prefix")
```

## Creating Loggers

```go
// Create new logger
logger := log.New(os.Stderr)

// With options
logger := log.NewWithOptions(os.Stderr, log.Options{
    ReportCaller:    true,
    ReportTimestamp: true,
    TimeFormat:      time.Kitchen,
    Level:           log.DebugLevel,
    Prefix:          "myapp",
})
```

## Log Levels

```go
log.DebugLevel  // Detailed debugging info
log.InfoLevel   // General information
log.WarnLevel   // Warning messages
log.ErrorLevel  // Error messages
log.FatalLevel  // Fatal errors (calls os.Exit(1))

// Set level
log.SetLevel(log.DebugLevel)
logger.SetLevel(log.WarnLevel)
```

## Structured Logging

```go
// Key-value pairs
log.Info("User login", "username", "alice", "ip", "192.168.1.1")

// Multiple pairs
log.Error("Database error",
    "operation", "INSERT",
    "table", "users",
    "error", err,
)

// Any types
ingredients := []string{"flour", "butter", "sugar"}
log.Debug("Baking", "ingredients", ingredients)
// DEBUG Baking ingredients="[flour butter sugar]"
```

## Formatters

```go
// Text (default, colored)
logger.SetFormatter(log.TextFormatter)

// JSON (structured, machine-readable)
logger.SetFormatter(log.JSONFormatter)

// Logfmt (key=value format)
logger.SetFormatter(log.LogfmtFormatter)
```

### JSON Output

```go
logger := log.New(os.Stderr)
logger.SetFormatter(log.JSONFormatter)
logger.Info("User created", "id", 123, "email", "user@example.com")
// {"level":"info","msg":"User created","id":123,"email":"user@example.com"}
```

## Styling

### Custom Styles

```go
import "github.com/charmbracelet/lipgloss"

styles := log.DefaultStyles()

// Customize error level
styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
    SetString("ERROR!!").
    Padding(0, 1).
    Background(lipgloss.Color("204")).
    Foreground(lipgloss.Color("0"))

// Custom key styles
styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
styles.Values["err"] = lipgloss.NewStyle().Bold(true)

logger.SetStyles(styles)
logger.Error("Whoops!", "err", "kitchen on fire")
```

## Logger Options

```go
logger.SetReportCaller(true)       // Show file:line
logger.SetReportTimestamp(true)    // Show timestamp
logger.SetTimeFormat(time.RFC3339) // Custom time format
logger.SetPrefix("myapp")          // Add prefix
logger.SetOutput(writer)           // Change output
```

## Sub-loggers

Create loggers with specific context:

```go
logger := log.NewWithOptions(os.Stderr, log.Options{
    Prefix: "server",
})

// Sub-logger with additional fields
requestLogger := logger.With("method", "GET", "path", "/api/users")
requestLogger.Info("Request received")
// INFO server: Request received method=GET path=/api/users

batchLogger := logger.With("batch", 2, "items", 100)
batchLogger.Debug("Processing batch")
// DEBUG server: Processing batch batch=2 items=100
```

## Format Messages

```go
// Sprintf-style formatting
for i := 1; i <= 100; i++ {
    log.Infof("Processing item %d/100", i)
}

log.Debugf("User %s logged in from %s", username, ip)
log.Errorf("Failed to connect to %s: %v", host, err)
```

### All Format Methods

```go
log.Debugf(format, args...)
log.Infof(format, args...)
log.Warnf(format, args...)
log.Errorf(format, args...)
log.Fatalf(format, args...)  // Exits program
log.Printf(format, args...)
```

## Helper Functions

Skip caller frames for cleaner stack traces:

```go
func startServer(port int) {
    log.Helper()  // Mark as helper
    log.Info("Starting server", "port", port)
}

log.SetReportCaller(true)
startServer(8080)
// INFO <main.go:15> Starting server port=8080
// Shows caller of startServer, not inside it
```

## Integration with Bubble Tea

### File Logging

Since Bubble Tea uses stdout, log to a file:

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/log"
)

func main() {
    // Set up file logging
    if len(os.Getenv("DEBUG")) > 0 {
        f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            fmt.Println("Failed to open log file:", err)
            os.Exit(1)
        }
        defer f.Close()

        log.SetOutput(f)
        log.SetLevel(log.DebugLevel)
    }

    // Run Bubble Tea app
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        log.Fatal("App error", "err", err)
    }
}

// In your model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    log.Debug("Update called", "msg", msg)
    // ...
    return m, nil
}
```

### Bubble Tea's Built-in Logging

```go
// Bubble Tea's logging (alternative)
if len(os.Getenv("DEBUG")) > 0 {
    f, err := tea.LogToFile("debug.log", "debug")
    if err != nil {
        fmt.Println("fatal:", err)
        os.Exit(1)
    }
    defer f.Close()
}

// Then use standard log package or Charm log
import stdlog "log"
stdlog.Println("Debug message")
```

### Structured Debug Info

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    log.Debug("State transition",
        "from", m.state,
        "msg", fmt.Sprintf("%T", msg),
        "cursor", m.cursor,
    )

    switch msg := msg.(type) {
    case tea.KeyMsg:
        log.Info("Key pressed",
            "key", msg.String(),
            "mode", m.mode,
        )
    }

    return m, nil
}
```

## Slog Handler

Use as an `log/slog` handler:

```go
import "log/slog"

handler := log.New(os.Stderr)
logger := slog.New(handler)
logger.Error("Something went wrong")
```

## Standard Log Adapter

Wrap for libraries that expect `*log.Logger`:

```go
import "net/http"

logger := log.NewWithOptions(os.Stderr, log.Options{
    Prefix: "http",
})

stdlog := logger.StandardLog(log.StandardLogOptions{
    ForceLevel: log.ErrorLevel,
})

server := &http.Server{
    Addr:     ":8080",
    Handler:  handler,
    ErrorLog: stdlog,
}

stdlog.Printf("Failed to handle request, %s", err)
// ERROR http: Failed to handle request, connection timeout
```

## Context Integration

```go
import "context"

ctx := context.Background()

// Store logger in context
ctx = log.WithContext(ctx, logger)

// Retrieve from context
logger := log.FromContext(ctx)
logger.Info("Using context logger")
```

## Common Patterns

### Error Logging with Stack Trace

```go
func processTask(id int) error {
    log.Debug("Starting task", "id", id)

    if err := doWork(); err != nil {
        log.Error("Task failed",
            "id", id,
            "error", err,
            "stack", string(debug.Stack()),
        )
        return err
    }

    log.Info("Task completed", "id", id)
    return nil
}
```

### Performance Timing

```go
func slowOperation() {
    start := time.Now()
    defer func() {
        log.Debug("Operation completed",
            "duration", time.Since(start),
        )
    }()

    // ... do work
}
```

### Conditional Logging

```go
if log.GetLevel() <= log.DebugLevel {
    // Expensive debug info
    details := computeExpensiveDetails()
    log.Debug("Detailed state", "details", details)
}
```

### Request Logging Middleware

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        log.Info("Request started",
            "method", r.Method,
            "path", r.URL.Path,
            "remote", r.RemoteAddr,
        )

        next.ServeHTTP(w, r)

        log.Info("Request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "duration", time.Since(start),
        )
    })
}
```

## Environment Variables

```bash
# Enable debug logging
DEBUG=1 ./myapp

# Or in code:
if os.Getenv("DEBUG") != "" {
    log.SetLevel(log.DebugLevel)
}
```

## Best Practices

1. **Log to Files in TUIs**: Don't log to stdout when using Bubble Tea
2. **Use Appropriate Levels**: Debug for verbose, Info for important events
3. **Add Context**: Include relevant key-value pairs
4. **Structured Logging**: Use key-value pairs instead of string interpolation
5. **Performance**: Avoid expensive operations in log statements
6. **Secrets**: Never log sensitive data (passwords, tokens, etc.)
7. **Consistent Keys**: Use same keys across codebase ("user_id", not "userId" and "user_id")

## Production Configuration

```go
func setupLogging() *log.Logger {
    logger := log.New(os.Stderr)

    if os.Getenv("ENV") == "production" {
        logger.SetFormatter(log.JSONFormatter)
        logger.SetLevel(log.InfoLevel)
        logger.SetReportCaller(false)
    } else {
        logger.SetFormatter(log.TextFormatter)
        logger.SetLevel(log.DebugLevel)
        logger.SetReportCaller(true)
    }

    return logger
}
```

## Reference

- GitHub: https://github.com/charmbracelet/log
- Docs: https://pkg.go.dev/github.com/charmbracelet/log
- Examples: https://github.com/charmbracelet/log/tree/main/examples
