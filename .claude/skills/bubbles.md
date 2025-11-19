# Bubbles - UI Components Expert

Expert in Bubbles, the collection of reusable TUI components for Bubble Tea applications.

## When to Use This Skill

Use this skill when:
- Building Bubble Tea applications with common UI patterns
- Creating lists, tables, or data displays
- Adding spinners, progress bars, or loading indicators
- Building text input forms or text editors
- Creating paginated views or scrollable content
- Adding help screens or keybinding displays

## Component Overview

Bubbles provides production-ready components:

- **Spinner** - Loading indicators
- **Text Input** - Single-line text fields
- **Text Area** - Multi-line text editors
- **Table** - Tabular data display
- **Progress** - Progress bars (static or animated)
- **Paginator** - Pagination logic and UI
- **Viewport** - Scrollable content
- **List** - Filterable, paginated lists with help
- **File Picker** - File system navigation
- **Timer** - Countdown timers
- **Stopwatch** - Count-up timers
- **Help** - Keybinding help display
- **Key** - Keybinding management

## Spinner

Loading indicators with various styles.

```go
import "github.com/charmbracelet/bubbles/spinner"

type model struct {
    spinner spinner.Model
}

func initialModel() model {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    return model{spinner: s}
}

func (m model) Init() tea.Cmd {
    return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return fmt.Sprintf("\n\n   %s Loading...\n\n", m.spinner.View())
}
```

### Spinner Styles

```go
spinner.Line
spinner.Dot
spinner.MiniDot
spinner.Jump
spinner.Pulse
spinner.Points
spinner.Globe
spinner.Moon
spinner.Monkey
// ... and many more
```

## Text Input

Single-line text fields with validation.

```go
import "github.com/charmbracelet/bubbles/textinput"

type model struct {
    textInput textinput.Model
}

func initialModel() model {
    ti := textinput.New()
    ti.Placeholder = "Enter your name..."
    ti.Focus()
    ti.CharLimit = 50
    ti.Width = 30

    return model{textInput: ti}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return fmt.Sprintf(
        "What's your name?\n\n%s\n\n",
        m.textInput.View(),
    )
}
```

### Text Input Options

```go
ti.Placeholder = "Type here..."
ti.CharLimit = 100
ti.Width = 50
ti.Focus()
ti.Blur()
ti.SetValue("initial value")
ti.EchoMode = textinput.EchoPassword  // Password field
ti.EchoCharacter = '•'

// Validation
ti.Validate = func(s string) error {
    if len(s) < 3 {
        return errors.New("too short")
    }
    return nil
}

// Get value
value := ti.Value()
```

## Text Area

Multi-line text editor.

```go
import "github.com/charmbracelet/bubbles/textarea"

type model struct {
    textarea textarea.Model
}

func initialModel() model {
    ta := textarea.New()
    ta.Placeholder = "Enter text..."
    ta.Focus()
    ta.CharLimit = 500

    return model{textarea: ta}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.textarea, cmd = m.textarea.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.textarea.View()
}
```

### Text Area Options

```go
ta.SetWidth(80)
ta.SetHeight(10)
ta.SetValue("Initial content")
ta.ShowLineNumbers = true
ta.CharLimit = 1000

// Get value
content := ta.Value()
```

## Table

Tabular data display with navigation.

```go
import "github.com/charmbracelet/bubbles/table"

func makeTable() table.Model {
    columns := []table.Column{
        {Title: "ID", Width: 10},
        {Title: "Name", Width: 20},
        {Title: "Email", Width: 30},
    }

    rows := []table.Row{
        {"1", "Alice", "alice@example.com"},
        {"2", "Bob", "bob@example.com"},
        {"3", "Charlie", "charlie@example.com"},
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(10),
    )

    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240")).
        BorderBottom(true).
        Bold(false)
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(false)

    t.SetStyles(s)
    return t
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.table.View()
}
```

### Table Operations

```go
t.SetRows(newRows)
t.SetColumns(newColumns)
t.Focus()
t.Blur()

// Get selected row
selectedRow := t.SelectedRow()
cursor := t.Cursor()
```

## Progress Bar

Static or animated progress indicators.

```go
import "github.com/charmbracelet/bubbles/progress"

type model struct {
    progress progress.Model
    percent  float64
}

func initialModel() model {
    return model{
        progress: progress.New(progress.WithGradient("#FF7CCB", "#FDFF8C")),
        percent:  0.0,
    }
}

func (m model) View() string {
    return m.progress.ViewAs(m.percent)
}

// Update progress
m.percent = 0.75  // 75%
```

### Progress Options

```go
// Solid color
progress.New(progress.WithSolidFill("#7571F9"))

// Gradient
progress.New(progress.WithGradient("#FF7CCB", "#FDFF8C"))

// Width
progress.New(progress.WithWidth(80))

// Without percentage display
progress.New(progress.WithoutPercentage())
```

## Paginator

Pagination logic and UI.

```go
import "github.com/charmbracelet/bubbles/paginator"

type model struct {
    paginator paginator.Model
    items     []string
}

func initialModel() model {
    items := make([]string, 100)
    for i := range items {
        items[i] = fmt.Sprintf("Item %d", i)
    }

    p := paginator.New()
    p.Type = paginator.Dots
    p.PerPage = 10
    p.SetTotalPages(len(items) / 10)

    return model{
        paginator: p,
        items:     items,
    }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.paginator, cmd = m.paginator.Update(msg)
    return m, cmd
}

func (m model) View() string {
    start := m.paginator.Page * m.paginator.PerPage
    end := min(start+m.paginator.PerPage, len(m.items))

    var s string
    for _, item := range m.items[start:end] {
        s += item + "\n"
    }

    s += "\n" + m.paginator.View()
    return s
}
```

### Paginator Styles

```go
p.Type = paginator.Arabic    // 1, 2, 3
p.Type = paginator.Dots      // • • •
```

## Viewport

Scrollable content container.

```go
import "github.com/charmbracelet/bubbles/viewport"

type model struct {
    viewport viewport.Model
    ready    bool
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        if !m.ready {
            m.viewport = viewport.New(msg.Width, msg.Height)
            m.viewport.SetContent(largeContent)
            m.ready = true
        } else {
            m.viewport.Width = msg.Width
            m.viewport.Height = msg.Height
        }
    }

    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.viewport.View()
}
```

### Viewport Options

```go
vp.SetContent(content)
vp.Width = 80
vp.Height = 24
vp.YPosition = 0

// Mouse wheel support (enable in program)
vp.MouseWheelEnabled = true

// Scroll methods
vp.HalfViewDown()
vp.HalfViewUp()
vp.LineDown(1)
vp.LineUp(1)
vp.GotoTop()
vp.GotoBottom()

// Check position
vp.AtTop()
vp.AtBottom()
```

## List

Full-featured list with filtering, pagination, and help.

```go
import "github.com/charmbracelet/bubbles/list"

type item struct {
    title       string
    description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

func makeList() list.Model {
    items := []list.Item{
        item{title: "Task 1", description: "First task"},
        item{title: "Task 2", description: "Second task"},
        item{title: "Task 3", description: "Third task"},
    }

    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "My Tasks"
    l.SetFilteringEnabled(true)
    l.SetShowStatusBar(true)
    l.SetShowHelp(true)

    return l
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.list.View()
}
```

### List Operations

```go
l.SetItems(newItems)
l.InsertItem(0, newItem)
l.RemoveItem(index)

selected := l.SelectedItem()
index := l.Index()

l.SetSize(width, height)
l.StartSpinner()
l.StopSpinner()

// Filtering
l.SetFilteringEnabled(true)
l.ResetFilter()
```

## File Picker

File system navigation.

```go
import "github.com/charmbracelet/bubbles/filepicker"

type model struct {
    filepicker filepicker.Model
}

func initialModel() model {
    fp := filepicker.New()
    fp.CurrentDirectory, _ = os.UserHomeDir()
    fp.AllowedTypes = []string{".md", ".txt"}

    return model{filepicker: fp}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.filepicker, cmd = m.filepicker.Update(msg)

    if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
        // User selected a file
        return m, loadFile(path)
    }

    return m, cmd
}
```

## Timer

Countdown timer.

```go
import "github.com/charmbracelet/bubbles/timer"

type model struct {
    timer timer.Model
}

func initialModel() model {
    return model{
        timer: timer.New(time.Minute * 5),
    }
}

func (m model) Init() tea.Cmd {
    return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case timer.TickMsg:
        var cmd tea.Cmd
        m.timer, cmd = m.timer.Update(msg)
        return m, cmd

    case timer.TimeoutMsg:
        // Timer finished
        return m, nil
    }

    return m, nil
}

func (m model) View() string {
    return m.timer.View()
}
```

## Stopwatch

Count-up timer.

```go
import "github.com/charmbracelet/bubbles/stopwatch"

type model struct {
    stopwatch stopwatch.Model
}

func initialModel() model {
    return model{
        stopwatch: stopwatch.New(),
    }
}

func (m model) Init() tea.Cmd {
    return m.stopwatch.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "s":
            m.stopwatch.Toggle()
        case "r":
            m.stopwatch.Reset()
        }
    }

    var cmd tea.Cmd
    m.stopwatch, cmd = m.stopwatch.Update(msg)
    return m, cmd
}
```

## Key Bindings

Manage keybindings and generate help.

```go
import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
    Up   key.Binding
    Down key.Binding
    Quit key.Binding
}

var keys = keyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "move up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"),
        key.WithHelp("↓/j", "move down"),
    ),
    Quit: key.NewBinding(
        key.WithKeys("q", "ctrl+c"),
        key.WithHelp("q", "quit"),
    ),
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, keys.Up):
            // Handle up
        case key.Matches(msg, keys.Down):
            // Handle down
        case key.Matches(msg, keys.Quit):
            return m, tea.Quit
        }
    }
    return m, nil
}
```

## Help

Display keybindings to users.

```go
import "github.com/charmbracelet/bubbles/help"

type model struct {
    help help.Model
    keys keyMap
}

func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Up, k.Down, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down},
        {k.Quit},
    }
}

func (m model) View() string {
    return m.help.View(m.keys)
}
```

## Best Practices

1. **Initialize in Init()**: Start timers, spinners in Init()
2. **Handle WindowSizeMsg**: Resize viewports and lists
3. **Delegate Updates**: Always update sub-components
4. **Check Return Values**: Some components return selection info
5. **Style Consistently**: Use theme colors across components

## Reference

- GitHub: https://github.com/charmbracelet/bubbles
- Docs: https://pkg.go.dev/github.com/charmbracelet/bubbles
