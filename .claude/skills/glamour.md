# Glamour - Markdown Rendering Expert

Expert in Glamour, the stylesheet-based markdown renderer for the terminal.

## When to Use This Skill

Use this skill when:
- Rendering markdown documents in the terminal
- Displaying formatted documentation in CLI tools
- Creating help pages with rich formatting
- Showing README files in terminal applications
- Building markdown-based TUIs
- Converting markdown to ANSI-styled terminal output

## Core Concepts

Glamour renders markdown to ANSI-styled terminal output using customizable stylesheets. It automatically detects the terminal's color capabilities and background color.

## Basic Usage

```go
import "github.com/charmbracelet/glamour"

markdown := `# Hello World

This is **bold** and this is *italic*.

- Item 1
- Item 2
- Item 3
`

// Simple render with auto-detected style
out, err := glamour.Render(markdown, "dark")
fmt.Print(out)
```

## Quick Rendering

```go
// Auto-detect dark/light theme
out, err := glamour.Render(markdown, "auto")

// Specific themes
out, err := glamour.Render(markdown, "dark")
out, err := glamour.Render(markdown, "light")
out, err := glamour.Render(markdown, "pink")
out, err := glamour.Render(markdown, "notty")  // No styling

// Custom style file
out, err := glamour.Render(markdown, "/path/to/style.json")
```

## Custom Renderer

```go
r, err := glamour.NewTermRenderer(
    // Detect background and use appropriate theme
    glamour.WithAutoStyle(),

    // Or specify a style
    glamour.WithStylePath("dark"),

    // Word wrap at specific width
    glamour.WithWordWrap(80),

    // Preserve newlines
    glamour.WithPreservedNewLines(),

    // Emoji support
    glamour.WithEmoji(),
)

out, err := r.Render(markdown)
fmt.Print(out)
```

## Available Styles

Built-in style gallery:

```go
"auto"       // Auto-detect dark/light
"dark"       // Default dark theme
"light"      // Default light theme
"pink"       // Pink theme
"notty"      // No styling (plain text)
"ascii"      // ASCII-only (no Unicode)
"dracula"    // Dracula theme
"tokyo-night" // Tokyo Night theme
```

## Renderer Options

### Auto Style Detection

```go
glamour.WithAutoStyle()  // Detect terminal background and choose theme
```

### Word Wrapping

```go
glamour.WithWordWrap(80)           // Wrap at 80 columns
glamour.WithWordWrap(0)            // No wrapping
```

### Preserve Newlines

```go
glamour.WithPreservedNewLines()  // Keep original line breaks
```

### Base URL for Links

```go
glamour.WithBaseURL("https://github.com/user/repo")
```

### Environment Config

```go
// Use GLAMOUR_STYLE environment variable
glamour.WithEnvironmentConfig()

// User can set:
// export GLAMOUR_STYLE=pink
// export GLAMOUR_STYLE=/path/to/custom.json
```

## Custom Styles

Create custom JSON stylesheets:

```json
{
  "document": {
    "block_prefix": "\n",
    "block_suffix": "\n",
    "color": "252",
    "margin": 2
  },
  "heading": {
    "block_suffix": "\n",
    "color": "39",
    "bold": true
  },
  "h1": {
    "prefix": " ",
    "suffix": " ",
    "color": "228",
    "background_color": "63",
    "bold": true
  },
  "code_block": {
    "color": "244",
    "margin": 2,
    "chroma": {
      "text": {"color": "#C4C4C4"},
      "error": {"color": "#F1F1F1", "background_color": "#F05B5B"},
      "keyword": {"color": "#00E9E9"},
      "string": {"color": "#B9F18B"}
    }
  },
  "list": {
    "level_indent": 2
  },
  "link": {
    "color": "30",
    "underline": true
  }
}
```

## Integration with Bubble Tea

### Simple Markdown View

```go
type model struct {
    content  string
    renderer *glamour.TermRenderer
}

func initialModel() model {
    r, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )

    markdown := "# Title\n\nSome **content**"
    rendered, _ := r.Render(markdown)

    return model{
        content:  rendered,
        renderer: r,
    }
}

func (m model) View() string {
    return m.content
}
```

### Dynamic Markdown Rendering

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Re-render with new width
        r, _ := glamour.NewTermRenderer(
            glamour.WithAutoStyle(),
            glamour.WithWordWrap(msg.Width - 4),
        )
        m.content, _ = r.Render(m.markdown)
        return m, nil
    }
    return m, nil
}
```

### Scrollable Markdown Viewer

```go
import "github.com/charmbracelet/bubbles/viewport"

type model struct {
    markdown string
    viewport viewport.Model
    ready    bool
}

func (m model) Init() tea.Cmd {
    // Render markdown
    r, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    content, _ := r.Render(m.markdown)

    m.viewport.SetContent(content)
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        if !m.ready {
            m.viewport = viewport.New(msg.Width, msg.Height)
            m.ready = true

            // Render and set content
            r, _ := glamour.NewTermRenderer(
                glamour.WithAutoStyle(),
                glamour.WithWordWrap(msg.Width - 4),
            )
            content, _ := r.Render(m.markdown)
            m.viewport.SetContent(content)
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

## Loading Markdown Files

```go
content, err := os.ReadFile("README.md")
if err != nil {
    log.Fatal(err)
}

out, err := glamour.Render(string(content), "dark")
fmt.Print(out)
```

## Common Patterns

### README Viewer

```go
func showReadme() {
    readme, _ := os.ReadFile("README.md")

    r, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(100),
    )

    out, _ := r.Render(string(readme))
    fmt.Print(out)
}
```

### Help Command

```go
func showHelp() {
    help := `# My App Help

## Commands

- **start** - Start the server
- **stop** - Stop the server
- **status** - Check status

## Examples

` + "```bash" + `
myapp start --port 8080
myapp status
` + "```" + `
`

    out, _ := glamour.Render(help, "auto")
    fmt.Print(out)
}
```

### Documentation Browser

```go
type DocBrowser struct {
    docs     map[string]string  // topic -> markdown content
    current  string
    viewport viewport.Model
}

func (d *DocBrowser) RenderTopic(topic string) string {
    markdown := d.docs[topic]

    r, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(d.viewport.Width - 4),
    )

    content, _ := r.Render(markdown)
    return content
}
```

### Markdown Preview

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case fileChangedMsg:
        // Re-render when file changes
        content, _ := os.ReadFile(msg.path)
        rendered, _ := m.renderer.Render(string(content))
        m.viewport.SetContent(rendered)
        return m, nil
    }
    return m, nil
}
```

## Markdown Features Supported

- **Headings** (H1-H6)
- **Bold** and *italic* text
- `Code` inline and blocks
- [Links](https://example.com)
- > Block quotes
- Lists (ordered and unordered)
- Tables
- Horizontal rules
- Task lists
- Strikethrough
- Images (as links in terminal)
- Syntax highlighting in code blocks

## Syntax Highlighting

Glamour uses Chroma for syntax highlighting in code blocks:

````markdown
```go
func main() {
    fmt.Println("Hello, World!")
}
```
````

Supports 200+ languages:
- Go, Python, JavaScript, TypeScript
- Rust, C, C++, Java, Ruby
- HTML, CSS, JSON, YAML, TOML
- SQL, Bash, Markdown
- And many more...

## Performance Tips

1. **Cache Rendered Content**: Don't re-render on every frame
2. **Render Once**: Store rendered output in model
3. **Lazy Rendering**: Only render visible content for large documents
4. **Reuse Renderer**: Create renderer once, reuse for multiple documents

## Best Practices

1. **Set Word Wrap**: Match terminal width for readable output
2. **Use Auto Style**: Let Glamour detect the best theme
3. **Handle Errors**: Markdown parsing can fail, always check errors
4. **Respect Terminal Capabilities**: Use `notty` style for pipes/redirects
5. **Test Different Themes**: Preview your markdown in different styles

## Environment Variables

```bash
# User can customize rendering
export GLAMOUR_STYLE=pink
export GLAMOUR_STYLE=dark
export GLAMOUR_STYLE=/path/to/custom.json

# Then in code:
glamour.RenderWithEnvironmentConfig(markdown)
```

## Creating Custom Themes

See the [style gallery](https://github.com/charmbracelet/glamour/tree/master/styles/gallery) for examples.

Theme JSON structure:
```json
{
  "document": {...},
  "heading": {...},
  "h1": {...},
  "h2": {...},
  "paragraph": {...},
  "code_block": {...},
  "list": {...},
  "link": {...},
  "image": {...},
  "table": {...}
}
```

## Use Cases

1. **CLI Documentation**: Render help docs beautifully
2. **README Viewers**: Show project READMEs in terminal
3. **Note-taking Apps**: Display markdown notes
4. **Change Logs**: Render CHANGELOG.md with style
5. **Blog Readers**: Terminal-based blog reading
6. **Git Tools**: Show commit messages, PR descriptions

## Reference

- GitHub: https://github.com/charmbracelet/glamour
- Docs: https://pkg.go.dev/github.com/charmbracelet/glamour
- Styles: https://github.com/charmbracelet/glamour/tree/master/styles/gallery
- Style Guide: https://github.com/charmbracelet/glamour/tree/master/styles
