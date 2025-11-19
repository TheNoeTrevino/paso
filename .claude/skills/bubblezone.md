# BubbleZone - Mouse Event Expert

Expert in BubbleZone, the library for easy mouse event handling in Bubble Tea applications.

## When to Use This Skill

Use this skill when:
- Adding mouse support to Bubble Tea applications
- Creating clickable buttons, links, or UI elements
- Building drag-and-drop interfaces
- Implementing context menus or tooltips
- Making any component mouse-interactive
- Tracking mouse position relative to components

## Core Concepts

BubbleZone wraps components with invisible zone markers that track their position on screen. This eliminates manual position calculation for mouse events.

## Why BubbleZone?

**Without BubbleZone:**
```go
// Manual position tracking - tedious!
if mouse.X >= buttonX && mouse.X <= buttonX+buttonWidth &&
   mouse.Y >= buttonY && mouse.Y <= buttonY+buttonHeight {
    // Click detected
}
```

**With BubbleZone:**
```go
// Simple and clean
if zone.Get("my-button").InBounds(mouse) {
    // Click detected
}
```

## Installation

```go
go get github.com/lrstanley/bubblezone
```

## Basic Usage

```go
import zone "github.com/lrstanley/bubblezone"

type model struct {
    zone *zone.Manager
}

func initialModel() model {
    return model{
        zone: zone.New(),
    }
}

func (m model) View() string {
    button := lipgloss.NewStyle().
        Padding(0, 2).
        Background(lipgloss.Color("62")).
        Render("Click me!")

    // Mark the button with an ID
    return m.zone.Mark("my-button", button)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        if msg.Type == tea.MouseLeft {
            // Check if click is in button bounds
            if m.zone.Get("my-button").InBounds(msg) {
                // Button was clicked!
                return m, handleButtonClick
            }
        }
    }
    return m, nil
}
```

## Manager API

### Creating a Manager

```go
// New manager (use one per model or globally)
manager := zone.New()

// Global manager (simpler for single-window apps)
// Import side effect initializes global manager
import _ "github.com/lrstanley/bubblezone"

// Use global manager
zone.Mark("id", content)
zone.Get("id")
```

### Scanning Updates

BubbleZone needs to scan the final rendered output to track positions:

```go
func (m model) View() string {
    content := m.renderContent()

    // Scan to update zone positions
    return m.zone.Scan(content)
}

// Or with global manager
func (m model) View() string {
    content := m.renderContent()
    return zone.Scan(content)
}
```

## Marking Zones

```go
// Mark any content with an ID
zone.Mark("button-1", buttonView)
zone.Mark("list-item-5", itemView)
zone.Mark("menu", menuView)

// Marks are invisible, zero-width
lipgloss.Width(zone.Mark("id", "content")) == lipgloss.Width("content")
```

## Checking Mouse Events

### InBounds

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        switch msg.Type {
        case tea.MouseLeft:
            if zone.Get("button").InBounds(msg) {
                // Left click on button
            }

        case tea.MouseRight:
            if zone.Get("item").InBounds(msg) {
                // Right click on item
            }

        case tea.MouseWheelUp:
            if zone.Get("scrollable").InBounds(msg) {
                // Scroll up in area
            }
        }
    }
    return m, nil
}
```

### Getting Zone Info

```go
z := zone.Get("my-zone")

// Check bounds
if z.InBounds(mouse) { }

// Get position
x, y := z.Pos()

// Get size
width, height := z.Size()

// Get relative mouse position
relX, relY := z.RawPos(mouse)
```

## Enabling Mouse Support

```go
// Enable in Bubble Tea program
p := tea.NewProgram(
    initialModel(),
    tea.WithMouseCellMotion(),  // Track all motion
    tea.WithMouseAllMotion(),   // Track outside window too
)
```

## Complete Example: Clickable List

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    zone "github.com/lrstanley/bubblezone"
)

type model struct {
    zone    *zone.Manager
    items   []string
    clicked string
}

func initialModel() model {
    return model{
        zone:  zone.New(),
        items: []string{"Apple", "Banana", "Cherry", "Date"},
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }

    case tea.MouseMsg:
        if msg.Type == tea.MouseLeft {
            // Check each item
            for i, item := range m.items {
                if m.zone.Get(fmt.Sprintf("item-%d", i)).InBounds(msg) {
                    m.clicked = item
                }
            }
        }
    }

    return m, nil
}

func (m model) View() string {
    var items []string

    for i, item := range m.items {
        style := lipgloss.NewStyle().Padding(0, 2)

        if item == m.clicked {
            style = style.Background(lipgloss.Color("62"))
        }

        // Mark each item
        rendered := m.zone.Mark(
            fmt.Sprintf("item-%d", i),
            style.Render(item),
        )
        items = append(items, rendered)
    }

    view := lipgloss.JoinVertical(lipgloss.Left, items...)

    if m.clicked != "" {
        view += "\n\nClicked: " + m.clicked
    }

    // Scan to update positions
    return m.zone.Scan(view)
}

func main() {
    p := tea.NewProgram(
        initialModel(),
        tea.WithMouseCellMotion(),
    )
    p.Run()
}
```

## Advanced Patterns

### Hover Effects

```go
type model struct {
    zone    *zone.Manager
    hovered string
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        // Update hover state on mouse motion
        m.hovered = ""
        for _, id := range []string{"btn1", "btn2", "btn3"} {
            if m.zone.Get(id).InBounds(msg) {
                m.hovered = id
            }
        }
    }
    return m, nil
}

func (m model) View() string {
    for _, id := range []string{"btn1", "btn2", "btn3"} {
        style := normalStyle
        if m.hovered == id {
            style = hoverStyle
        }
        button := m.zone.Mark(id, style.Render("Button"))
        // ... render
    }
}
```

### Drag and Drop

```go
type model struct {
    zone     *zone.Manager
    dragging bool
    dragID   string
    dragX, dragY int
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        switch msg.Type {
        case tea.MouseLeft:
            // Start drag
            for i := range m.items {
                id := fmt.Sprintf("item-%d", i)
                if m.zone.Get(id).InBounds(msg) {
                    m.dragging = true
                    m.dragID = id
                    m.dragX, m.dragY = msg.X, msg.Y
                }
            }

        case tea.MouseRelease:
            // End drag
            if m.dragging {
                // Find drop target
                for i := range m.items {
                    id := fmt.Sprintf("drop-%d", i)
                    if m.zone.Get(id).InBounds(msg) {
                        // Handle drop
                    }
                }
                m.dragging = false
            }
        }
    }
    return m, nil
}
```

### Context Menus

```go
type model struct {
    zone       *zone.Manager
    contextMenu *contextMenu
}

type contextMenu struct {
    x, y int
    targetID string
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        if msg.Type == tea.MouseRight {
            // Show context menu on right-click
            for _, id := range m.getItemIDs() {
                if m.zone.Get(id).InBounds(msg) {
                    m.contextMenu = &contextMenu{
                        x: msg.X,
                        y: msg.Y,
                        targetID: id,
                    }
                }
            }
        }
    }
    return m, nil
}

func (m model) View() string {
    view := m.renderMainView()

    if m.contextMenu != nil {
        menu := m.renderContextMenu()
        // Position menu at click location
        view = overlayAt(view, menu, m.contextMenu.x, m.contextMenu.y)
    }

    return m.zone.Scan(view)
}
```

### Tooltips

```go
type model struct {
    zone    *zone.Manager
    tooltip string
    tooltipX, tooltipY int
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        m.tooltip = ""

        // Check all tooltipable zones
        for id, text := range m.tooltips {
            if m.zone.Get(id).InBounds(msg) {
                m.tooltip = text
                m.tooltipX, m.tooltipY = msg.X, msg.Y
            }
        }
    }
    return m, nil
}

func (m model) View() string {
    view := m.renderContent()

    if m.tooltip != "" {
        tooltipBox := lipgloss.NewStyle().
            Background(lipgloss.Color("240")).
            Padding(0, 1).
            Render(m.tooltip)

        view = overlayAt(view, tooltipBox, m.tooltipX, m.tooltipY+1)
    }

    return m.zone.Scan(view)
}
```

### Scrollable Areas

```go
type model struct {
    zone   *zone.Manager
    offset int
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        if m.zone.Get("scrollable").InBounds(msg) {
            switch msg.Type {
            case tea.MouseWheelUp:
                m.offset = max(0, m.offset-1)
            case tea.MouseWheelDown:
                m.offset = min(len(m.items)-10, m.offset+1)
            }
        }
    }
    return m, nil
}
```

## Best Practices

1. **One Manager**: Use a single manager per model or globally
2. **Always Scan**: Call `.Scan()` in View() to update positions
3. **Unique IDs**: Ensure zone IDs are unique
4. **Enable Mouse**: Use `tea.WithMouseCellMotion()` in program
5. **Check Bounds First**: Always check InBounds before handling clicks
6. **Dynamic IDs**: Use template-style IDs: `fmt.Sprintf("item-%d", i)`

## Performance Tips

1. **Minimal Zones**: Only mark interactive elements
2. **Efficient IDs**: Use simple ID schemes
3. **Batch Checks**: Group related zone checks together
4. **Cache Zone Refs**: Don't call `.Get()` repeatedly in tight loops

## Common Pitfalls

1. **Forgetting .Scan()**: Positions won't update without scanning
2. **Missing Mouse Enable**: Program needs mouse support enabled
3. **Duplicate IDs**: Will cause incorrect hit detection
4. **Not Handling All Mouse Types**: Remember left, right, wheel, motion

## Integration with Bubbles

Works seamlessly with Bubble components:

```go
import "github.com/charmbracelet/bubbles/list"

func (m model) View() string {
    listView := m.list.View()

    // Make list items clickable
    for i := 0; i < m.list.Len(); i++ {
        // Mark each visible item
        // (Implementation depends on list internals)
    }

    return m.zone.Scan(listView)
}
```

## Reference

- GitHub: https://github.com/lrstanley/bubblezone
- Docs: https://pkg.go.dev/github.com/lrstanley/bubblezone
