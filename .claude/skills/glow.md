# Glow - Markdown Reader Expert

Expert in Glow, a complete terminal-based markdown reader with TUI and CLI modes.

## When to Use This Skill

Use this skill when:
- Building a markdown reader application
- Learning patterns from a production Bubble Tea app
- Understanding how to combine multiple Charm libraries
- Implementing file discovery and navigation
- Creating document viewers
- Building markdown-centric workflows

## What is Glow?

Glow is a complete, production-ready terminal markdown reader that demonstrates best practices for Bubble Tea applications. It serves as both:

1. **A tool**: Read markdown beautifully in your terminal
2. **A reference**: Learn from a real-world Bubble Tea application

## Modes of Operation

### CLI Mode (Pager)

Render markdown to terminal (one-shot):

```bash
# Read from file
glow README.md

# Read from stdin
echo "# Hello" | glow -

# Fetch from GitHub
glow github.com/charmbracelet/glow

# Fetch from URL
glow https://example.com/doc.md
```

### TUI Mode (Interactive)

Full-screen markdown browser:

```bash
# Launch TUI
glow

# Finds markdown files in:
# - Current directory
# - Subdirectories
# - Git repository (if in one)
```

## Architecture Lessons from Glow

### 1. Multi-Mode Application

```go
// Glow switches between modes
type state int

const (
    stateShowingPager state = iota
    stateShowingStash
)

type model struct {
    state       state
    pagerModel  pagerModel
    stashModel  stashModel
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.state {
    case stateShowingPager:
        return m.updatePager(msg)
    case stateShowingStash:
        return m.updateStash(msg)
    }
    return m, nil
}
```

### 2. File Discovery

```go
// Find markdown files
func findMarkdownFiles(root string) ([]string, error) {
    var files []string

    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() && isMarkdown(path) {
            files = append(files, path)
        }

        return nil
    })

    return files, err
}

func isMarkdown(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    return ext == ".md" || ext == ".markdown"
}
```

### 3. Glamour Integration

```go
type model struct {
    glamourRenderer *glamour.TermRenderer
    content         string
}

func (m *model) renderMarkdown(markdown string) error {
    rendered, err := m.glamourRenderer.Render(markdown)
    if err != nil {
        return err
    }

    m.content = rendered
    return nil
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
        m.loadFile,
        m.renderMarkdown,
    )
}
```

### 4. Viewport for Scrolling

```go
import "github.com/charmbracelet/bubbles/viewport"

type pagerModel struct {
    viewport viewport.Model
    content  string
}

func (m pagerModel) Update(msg tea.Msg) (pagerModel, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height
        m.viewport.SetContent(m.content)
    }

    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m pagerModel) View() string {
    return m.viewport.View()
}
```

### 5. List for File Browser

```go
import "github.com/charmbracelet/bubbles/list"

type fileItem struct {
    path  string
    title string
}

func (i fileItem) Title() string       { return i.title }
func (i fileItem) Description() string { return i.path }
func (i fileItem) FilterValue() string { return i.title }

type stashModel struct {
    list list.Model
}

func newStashModel(files []string) stashModel {
    items := make([]list.Item, len(files))
    for i, path := range files {
        items[i] = fileItem{
            path:  path,
            title: filepath.Base(path),
        }
    }

    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "Markdown Files"

    return stashModel{list: l}
}
```

### 6. Config File Support

```go
type Config struct {
    Style   string
    Width   int
    Mouse   bool
    Pager   bool
    ShowAll bool
}

func loadConfig() (*Config, error) {
    configPath := filepath.Join(os.UserConfigDir(), "glow", "glow.yml")

    data, err := os.ReadFile(configPath)
    if err != nil {
        return defaultConfig(), nil
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

### 7. Help System

```go
import "github.com/charmbracelet/bubbles/help"

type model struct {
    help    help.Model
    keys    keyMap
    showHelp bool
}

type keyMap struct {
    Up     key.Binding
    Down   key.Binding
    Quit   key.Binding
    Help   key.Binding
}

func (m model) View() string {
    view := m.mainContent()

    if m.showHelp {
        view += "\n\n" + m.help.View(m.keys)
    }

    return view
}
```

## Key Features Implemented

### 1. GitHub Integration

Fetch READMEs directly:

```bash
glow github.com/charmbracelet/glow
```

Implementation concept:
```go
func fetchGitHubReadme(repo string) (string, error) {
    // Parse: github.com/user/repo
    parts := strings.Split(repo, "/")
    user, repo := parts[1], parts[2]

    url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/README.md", user, repo)

    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    content, err := io.ReadAll(resp.Body)
    return string(content), err
}
```

### 2. Git Repository Detection

Find files in Git repo:

```go
func isGitRepo(path string) bool {
    gitPath := filepath.Join(path, ".git")
    _, err := os.Stat(gitPath)
    return err == nil
}

func getGitRoot() (string, error) {
    cmd := exec.Command("git", "rev-parse", "--show-toplevel")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}
```

### 3. Fuzzy Search

Filter files in TUI:

```go
// Bubbles list component includes fuzzy search
l := list.New(items, delegate, width, height)
l.SetFilteringEnabled(true)

// Users press "/" to start filtering
```

### 4. Style Selection

Theme switching:

```go
type model struct {
    style   string
    styles  []string
}

func (m model) cycleStyle() {
    currentIdx := indexOf(m.styles, m.style)
    nextIdx := (currentIdx + 1) % len(m.styles)
    m.style = m.styles[nextIdx]

    // Re-render with new style
    m.renderer, _ = glamour.NewTermRenderer(
        glamour.WithStylePath(m.style),
    )
}
```

## Patterns to Learn From

### 1. Responsive Layout

```go
func (m model) View() string {
    if m.width < 80 {
        // Compact layout for narrow terminals
        return m.compactView()
    }
    // Full layout for wide terminals
    return m.fullView()
}
```

### 2. Status Line

```go
func (m model) statusLine() string {
    leftStatus := fmt.Sprintf(" %s ", m.currentFile)
    rightStatus := fmt.Sprintf(" %d%% ", m.scrollPercent())

    statusStyle := lipgloss.NewStyle().
        Background(lipgloss.Color("236")).
        Foreground(lipgloss.Color("250")).
        Width(m.width)

    gap := m.width - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus)
    return statusStyle.Render(leftStatus + strings.Repeat(" ", gap) + rightStatus)
}
```

### 3. Error Handling

```go
type model struct {
    err error
}

func (m model) View() string {
    if m.err != nil {
        errorStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Bold(true)

        return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress q to quit", m.err))
    }

    return m.normalView()
}
```

### 4. Loading States

```go
type state int

const (
    stateLoading state = iota
    stateReady
    stateError
)

type model struct {
    state   state
    spinner spinner.Model
}

func (m model) View() string {
    switch m.state {
    case stateLoading:
        return fmt.Sprintf("\n\n   %s Loading...\n\n", m.spinner.View())
    case stateReady:
        return m.content
    case stateError:
        return m.errorView()
    }
    return ""
}
```

## Code Organization

Glow's structure:

```
glow/
├── main.go              # Entry point, CLI parsing
├── ui/
│   ├── ui.go           # Main TUI model
│   ├── stash.go        # File browser
│   ├── pager.go        # Document viewer
│   ├── style.go        # Styles
│   └── utils.go        # Helpers
├── markdown/
│   └── markdown.go     # Glamour integration
└── config/
    └── config.go       # Configuration
```

## CLI Integration

Glow seamlessly switches between CLI and TUI:

```go
func main() {
    // Parse flags
    style := flag.String("s", "auto", "style")
    width := flag.Int("w", 80, "width")

    flag.Parse()

    // TUI mode if no args
    if flag.NArg() == 0 {
        runTUI()
        return
    }

    // CLI mode with file argument
    file := flag.Arg(0)
    renderFile(file, *style, *width)
}
```

## Best Practices from Glow

1. **Graceful Degradation**: Works in any terminal
2. **Configuration**: User-customizable via config file
3. **Error Recovery**: Handles missing files, network errors gracefully
4. **Performance**: Fast rendering with caching
5. **UX Polish**: Status bars, help screens, keyboard shortcuts
6. **Accessibility**: Supports screen readers in accessible mode

## Using Glow as Inspiration for Paso

Lessons for building Paso:

1. **File Discovery**: How Glow finds markdown → How Paso finds task DB
2. **TUI/CLI Modes**: Glow's dual mode → Paso could have quick commands
3. **List Component**: File browser → Task/column browser
4. **Viewport**: Document scrolling → Column/task scrolling
5. **Status Line**: File info → Current task/column info
6. **Help System**: Keyboard shortcuts → Paso's keyboard navigation
7. **Config File**: User preferences → Paso's settings
8. **Responsive Layout**: Terminal sizing → Paso's column viewport

## Reference

- GitHub: https://github.com/charmbracelet/glow
- Install: `brew install glow` or `go install github.com/charmbracelet/glow@latest`
- Docs: README at https://github.com/charmbracelet/glow
- Try it: `glow github.com/charmbracelet/glow`
