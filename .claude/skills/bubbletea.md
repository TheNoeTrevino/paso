# Bubble Tea - TUI Framework Expert

Expert in Bubble Tea, the Go framework for building terminal user interfaces based on The Elm Architecture.

## When to Use This Skill

Use this skill when:
- Building any terminal user interface (TUI) application
- Creating interactive command-line tools
- Implementing Model-View-Update (MVU) architecture
- Managing application state and event handling
- Building full-screen or inline terminal applications

## Core Concepts

### The Elm Architecture (MVU Pattern)

Bubble Tea applications follow three core components:

1. **Model** - Application state
2. **Update** - Event handler that updates state
3. **View** - Renders UI from state

### Basic Structure

```go
package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
)

// 1. MODEL - Your application state
type model struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
}

// 2. INIT - Initial command (optional I/O)
func (m model) Init() tea.Cmd {
    // Return nil for no initial command
    // Or return a Cmd for initial I/O
    return nil
}

// 3. UPDATE - Event handler
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }
        case "enter", " ":
            _, ok := m.selected[m.cursor]
            if ok {
                delete(m.selected, m.cursor)
            } else {
                m.selected[m.cursor] = struct{}{}
            }
        }
    }
    return m, nil
}

// 4. VIEW - Render UI
func (m model) View() string {
    s := "What should we do?\n\n"

    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }

        checked := " "
        if _, ok := m.selected[i]; ok {
            checked = "x"
        }

        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    }

    s += "\nPress q to quit.\n"
    return s
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
    }
}
```

## Essential Message Types

### Built-in Messages

```go
// Keyboard input
case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+c", "q":
        return m, tea.Quit
    }

// Mouse events
case tea.MouseMsg:
    switch msg.Type {
    case tea.MouseLeft:
        // Left click at msg.X, msg.Y
    case tea.MouseWheelUp:
        // Scroll up
    }

// Window resize
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height

// Custom messages (via Cmd)
case myCustomMsg:
    // Handle your custom message
```

## Commands (Cmd)

Commands are functions that return messages. Use them for I/O and async operations.

```go
// Built-in commands
tea.Quit              // Exit the program
tea.Batch(cmds...)    // Run multiple commands
tea.Sequence(cmds...) // Run commands in sequence

// Custom command example
func checkServer() tea.Msg {
    resp, err := http.Get("https://api.example.com")
    if err != nil {
        return errMsg{err}
    }
    return serverStatusMsg{resp.StatusCode}
}

// Use in Update
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "r" {
            return m, checkServer
        }
    case serverStatusMsg:
        m.status = msg.code
        return m, nil
    }
    return m, nil
}
```

## Program Options

```go
// Standard program
tea.NewProgram(model)

// With alt screen (full screen mode)
tea.NewProgram(model, tea.WithAltScreen())

// With mouse support
tea.NewProgram(model, tea.WithMouseCellMotion())
tea.NewProgram(model, tea.WithMouseAllMotion())

// Custom input/output
tea.NewProgram(model, tea.WithInput(r), tea.WithOutput(w))
```

## Advanced Patterns

### Sub-models (Component Pattern)

```go
type mainModel struct {
    textInput textinput.Model // Sub-component
    focused   bool
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    // Delegate to sub-component
    m.textInput, cmd = m.textInput.Update(msg)

    return m, cmd
}

func (m mainModel) View() string {
    return m.textInput.View() // Render sub-component
}
```

### Batch Commands

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Run multiple commands concurrently
    return m, tea.Batch(
        fetchData,
        startTimer,
        playSound,
    )
}
```

### Tick/Animation Loop

```go
type tickMsg time.Time

func tick() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case tickMsg:
        m.counter++
        return m, tick() // Continue ticking
    }
    return m, nil
}

func (m model) Init() tea.Cmd {
    return tick() // Start ticking
}
```

## Key Handling Utilities

```go
// Check specific keys
switch msg.String() {
case "ctrl+c":
    return m, tea.Quit
case "enter":
    // Handle enter
}

// Key types
msg.Type == tea.KeyRunes     // Regular character
msg.Type == tea.KeyEnter     // Enter key
msg.Type == tea.KeyBackspace // Backspace
msg.Type == tea.KeyUp        // Arrow up
```

## Best Practices

1. **Keep Model Immutable**: Update fields, don't mutate slices in place
2. **Return Commands**: Always return from Update, even if `nil`
3. **One Source of Truth**: Store state in Model, not in View
4. **Separate Concerns**: Keep rendering logic in View, state logic in Update
5. **Use Sub-models**: Break complex UIs into smaller components
6. **Handle WindowSizeMsg**: Always respond to terminal resizing
7. **Test Pure Functions**: Model and Update are pure, easy to test

## Common Patterns

### Modal Dialogs

```go
type mode int

const (
    normalMode mode = iota
    dialogMode
)

type model struct {
    mode mode
    // ... other fields
}

func (m model) View() string {
    if m.mode == dialogMode {
        return renderDialog()
    }
    return renderNormalView()
}
```

### Multi-page Apps

```go
type page int

const (
    listPage page = iota
    detailPage
    settingsPage
)

type model struct {
    currentPage page
    // ... other fields
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.currentPage {
    case listPage:
        return m.updateListPage(msg)
    case detailPage:
        return m.updateDetailPage(msg)
    }
    return m, nil
}
```

### Loading States

```go
type model struct {
    loading bool
    data    []item
    err     error
}

func (m model) Init() tea.Cmd {
    return fetchData // Start loading
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case dataMsg:
        m.loading = false
        m.data = msg.items
        return m, nil
    case errMsg:
        m.loading = false
        m.err = msg.err
        return m, nil
    }
    return m, nil
}
```

## Performance Tips

1. **Minimize View Calls**: View is called frequently, keep it fast
2. **Use String Builders**: For large views, use `strings.Builder`
3. **Cache Rendered Content**: Store static strings in Model
4. **Debounce Updates**: Don't re-render on every keystroke if unnecessary
5. **Lazy Rendering**: Only render visible portions for large lists

## Integration with Other Libraries

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"      // Styling
    "github.com/charmbracelet/bubbles/list"  // Components
)

func (m model) View() string {
    style := lipgloss.NewStyle().Bold(true)
    return style.Render(m.list.View())
}
```

## Debugging

```go
// Log to file (stdout is occupied by TUI)
if len(os.Getenv("DEBUG")) > 0 {
    f, err := tea.LogToFile("debug.log", "debug")
    if err != nil {
        fmt.Println("fatal:", err)
        os.Exit(1)
    }
    defer f.Close()
}

// Then use log package
log.Println("Debug info:", m.someValue)
```

## Reference

- GitHub: https://github.com/charmbracelet/bubbletea
- Docs: https://pkg.go.dev/github.com/charmbracelet/bubbletea
- Examples: https://github.com/charmbracelet/bubbletea/tree/main/examples
- Tutorial: https://github.com/charmbracelet/bubbletea/tree/main/tutorials
