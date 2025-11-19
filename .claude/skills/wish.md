# Wish - SSH Server Expert

Expert in Wish, the library for building SSH applications and making TUIs accessible over SSH.

## When to Use This Skill

Use this skill when:
- Building SSH-accessible applications
- Creating remotely accessible TUIs
- Building Git servers over SSH
- Creating multi-user terminal applications
- Adding SSH authentication to tools
- Building SSH-based services (like Soft Serve, Wishlist)

## Core Concepts

Wish is an SSH server library with sensible defaults and middleware architecture. It makes building custom SSH apps as easy as building HTTP servers, and integrates seamlessly with Bubble Tea.

## Why Wish?

SSH offers several advantages for terminal applications:
- **Secure communication** without HTTPS certificates
- **User identification** with SSH keys
- **Access from any terminal** over network
- **Built-in encryption** and authentication
- **Works over firewalls** (port 22 is commonly open)

## Basic SSH Server

```go
import (
    "github.com/charmbracelet/wish"
    "github.com/charmbracelet/wish/bubbletea"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    server, err := wish.NewServer(
        wish.WithAddress(":2222"),
        wish.WithHostKeyPath(".ssh/host_key"),
        wish.WithMiddleware(
            bubbletea.Middleware(teaHandler),
            logging.Middleware(),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Starting SSH server on :2222")
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    return initialModel(), []tea.ProgramOption{
        tea.WithAltScreen(),
    }
}
```

## Middleware Architecture

Wish uses middleware like HTTP servers:

```go
server := wish.NewServer(
    wish.WithMiddleware(
        // Middlewares execute from first to last
        bubbletea.Middleware(teaHandler),  // Bubble Tea TUI
        logging.Middleware(),              // Connection logging
        activeterm.Middleware(),           // Require active terminal
        git.Middleware(gitHandler),        // Git server
    ),
)
```

## Bubble Tea Middleware

Run Bubble Tea apps over SSH:

```go
import "github.com/charmbracelet/wish/bubbletea"

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    // Create model for this session
    m := initialModel()

    // Optional: customize per-session
    m.username = s.User()

    // Return model and options
    return m, []tea.ProgramOption{
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
    }
}

// In server setup
wish.WithMiddleware(
    bubbletea.Middleware(teaHandler),
)
```

### Per-Session State

Each SSH connection gets its own model instance:

```go
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    // Session info
    user := s.User()
    remoteAddr := s.RemoteAddr()
    pty, _, _ := s.Pty()

    log.Printf("User %s connected from %s", user, remoteAddr)

    return model{
        session:  s,
        username: user,
        width:    pty.Window.Width,
        height:   pty.Window.Height,
    }, nil
}
```

## Git Middleware

Build Git servers:

```go
import "github.com/charmbracelet/wish/git"

func gitHandler(s ssh.Session) error {
    // Handle git operations
    repo := extractRepo(s.Command())

    // Custom authorization
    if !canAccess(s.User(), repo) {
        return errors.New("access denied")
    }

    // Serve git protocol
    return git.Handle(s, repo)
}

wish.WithMiddleware(
    git.Middleware(gitHandler),
)
```

## Logging Middleware

Log connections and commands:

```go
import "github.com/charmbracelet/wish/logging"

wish.WithMiddleware(
    logging.Middleware(),
)

// Logs:
// - Connection: user, address, terminal, dimensions
// - Command: what was invoked
// - Duration: how long session lasted
```

## Access Control Middleware

Restrict access to terminal-only connections:

```go
import "github.com/charmbracelet/wish/activeterm"

wish.WithMiddleware(
    activeterm.Middleware(),  // Reject non-terminal connections
)
```

Restrict specific commands:

```go
import "github.com/charmbracelet/wish/accesscontrol"

wish.WithMiddleware(
    accesscontrol.Middleware(
        accesscontrol.WithAllowedCommands([]string{"tui", "help"}),
    ),
)
```

## Authentication

### Public Key Authentication

```go
import "github.com/charmbracelet/wish"

server := wish.NewServer(
    wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
        // Custom auth logic
        username := ctx.User()

        // Check authorized keys
        return isAuthorized(username, key)
    }),
)
```

### Password Authentication

```go
server := wish.NewServer(
    wish.WithPasswordAuth(func(ctx ssh.Context, password string) bool {
        username := ctx.User()

        // Verify password
        return checkPassword(username, password)
    }),
)
```

### Unauthenticated (Development)

```go
server := wish.NewServer(
    // No auth - anyone can connect
)
```

## Host Key Generation

```go
import "github.com/charmbracelet/wish"

// Auto-generate if missing
wish.WithHostKeyPath(".ssh/host_key")

// Or provide explicit key
keyPair, _ := rsa.GenerateKey(rand.Reader, 2048)
wish.WithHostKey(keyPair)
```

## Complete Example: Multi-User TUI

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/charmbracelet/ssh"
    "github.com/charmbracelet/wish"
    "github.com/charmbracelet/wish/bubbletea"
    "github.com/charmbracelet/wish/logging"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    server, err := wish.NewServer(
        wish.WithAddress(":2222"),
        wish.WithHostKeyPath(".ssh/host_key"),
        wish.WithMiddleware(
            bubbletea.Middleware(teaHandler),
            logging.Middleware(),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGTERM)

    log.Printf("Starting SSH server on :2222")
    go func() {
        if err := server.ListenAndServe(); err != nil {
            log.Fatal(err)
        }
    }()

    <-done
    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    pty, _, _ := s.Pty()

    return model{
        username: s.User(),
        width:    pty.Window.Width,
        height:   pty.Window.Height,
    }, []tea.ProgramOption{
        tea.WithAltScreen(),
        tea.WithInput(s),
        tea.WithOutput(s),
    }
}

type model struct {
    username string
    width    int
    height   int
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("Welcome, %s!\n\nPress 'q' to quit.", m.username)
}
```

## Custom Renderers for SSH

Different clients may have different color support:

```go
import "github.com/charmbracelet/lipgloss"

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    // Create renderer for this client
    renderer := lipgloss.NewRenderer(s)

    // Use adaptive colors
    style := renderer.NewStyle().
        Background(lipgloss.AdaptiveColor{Light: "63", Dark: "228"})

    return model{
        renderer: renderer,
        style:    style,
    }, []tea.ProgramOption{
        tea.WithInput(s),
        tea.WithOutput(s),
    }
}
```

## Session Information

```go
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    // User info
    username := s.User()

    // Connection info
    remoteAddr := s.RemoteAddr()

    // Terminal info
    pty, winCh, isPty := s.Pty()
    if isPty {
        term := pty.Term
        width := pty.Window.Width
        height := pty.Window.Height
    }

    // Environment
    env := s.Environ()

    // Command (if any)
    command := s.Command()

    return model{/* ... */}, nil
}
```

## Graceful Shutdown

```go
func main() {
    server, _ := wish.NewServer(/* ... */)

    // Handle signals
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGTERM)

    go server.ListenAndServe()

    <-done
    log.Println("Shutting down...")

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Shutdown error:", err)
    }
}
```

## Common Patterns

### Multi-Command Server

```go
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    cmd := s.Command()

    switch {
    case len(cmd) == 0:
        // No command - default TUI
        return mainMenuModel(), nil

    case cmd[0] == "stats":
        return statsModel(), nil

    case cmd[0] == "help":
        return helpModel(), nil

    default:
        fmt.Fprintf(s, "Unknown command: %s\n", cmd[0])
        return nil, nil
    }
}
```

### User Database Integration

```go
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    username := s.User()

    // Load user data
    user, err := db.GetUser(username)
    if err != nil {
        fmt.Fprintf(s, "User not found\n")
        return nil, nil
    }

    return model{
        user: user,
        // ... user-specific state
    }, nil
}
```

### Session Tracking

```go
var activeSessions sync.Map

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    sessionID := generateID()
    activeSessions.Store(sessionID, s)

    return model{
        sessionID: sessionID,
        onExit: func() {
            activeSessions.Delete(sessionID)
        },
    }, nil
}
```

## Security Best Practices

1. **Always Use Host Keys**: Generate and persist host keys
2. **Validate Public Keys**: Implement proper public key authentication
3. **Rate Limiting**: Prevent brute force attacks
4. **Timeout Idle Connections**: Close inactive sessions
5. **Sanitize Input**: Validate all user input
6. **Audit Logging**: Log all access and commands
7. **Least Privilege**: Only grant necessary permissions

## Development Tips

### Local Testing

```bash
# Connect to your SSH server
ssh -p 2222 localhost

# With specific user
ssh -p 2222 alice@localhost

# With command
ssh -p 2222 localhost stats
```

### Avoid known_hosts Issues

Development config in `~/.ssh/config`:

```
Host localhost
    UserKnownHostsFile /dev/null
    StrictHostKeyChecking no
```

### Debugging

```go
import "github.com/charmbracelet/log"

func main() {
    log.SetLevel(log.DebugLevel)
    log.SetOutput(os.Stderr)

    // ... server setup
}
```

## Use Cases

1. **Remote TUIs**: Access terminal apps from anywhere
2. **Git Servers**: Host Git repositories (like Soft Serve)
3. **SSH Directories**: Menu of SSH apps (like Wishlist)
4. **Admin Tools**: Remote server management
5. **Multi-User Apps**: Collaborative terminal tools
6. **Game Servers**: Terminal-based multiplayer games
7. **CI/CD Tools**: Build server interfaces

## Integration with Paso

For Paso, Wish could enable:

```go
// Remote kanban board access
ssh paso.example.com

// Multiple users sharing boards
ssh -p 2222 alice@teamserver

// Git-like operations
ssh paso.example.com sync
```

## Reference

- GitHub: https://github.com/charmbracelet/wish
- Docs: https://pkg.go.dev/github.com/charmbracelet/wish
- Examples: https://github.com/charmbracelet/wish/tree/main/examples
- Soft Serve: https://github.com/charmbracelet/soft-serve (production example)
- Wishlist: https://github.com/charmbracelet/wishlist (SSH directory)
