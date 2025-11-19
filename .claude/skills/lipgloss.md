# Lip Gloss - Terminal Styling Expert

Expert in Lip Gloss, the CSS-like styling library for terminal layouts and formatting.

## When to Use This Skill

Use this skill when:
- Styling terminal output with colors, borders, padding, and margins
- Creating layouts with horizontal and vertical composition
- Building responsive terminal UIs that adapt to screen size
- Designing tables, lists, and structured terminal content
- Applying consistent theming across a TUI application

## Core Concepts

Lip Gloss provides a declarative, CSS-like API for styling terminal text. Styles are immutable and composable.

## Basic Styling

```go
import "github.com/charmbracelet/lipgloss"

// Create a style
var style = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    PaddingTop(2).
    PaddingLeft(4).
    Width(22)

// Apply the style
fmt.Println(style.Render("Hello, World!"))
```

## Color System

### Color Formats

```go
// ANSI 16 colors (4-bit)
lipgloss.Color("5")   // Magenta
lipgloss.Color("12")  // Light blue

// ANSI 256 colors (8-bit)
lipgloss.Color("86")  // Aqua
lipgloss.Color("201") // Hot pink

// True Color (24-bit)
lipgloss.Color("#0000FF")  // Blue
lipgloss.Color("#FF5733")  // Orange-red
```

### Adaptive Colors

```go
// Different colors for light/dark backgrounds
lipgloss.AdaptiveColor{
    Light: "236",  // For light terminals
    Dark: "248",   // For dark terminals
}
```

### Complete Colors

```go
// Specify exact values for each profile (no auto-degradation)
lipgloss.CompleteColor{
    TrueColor: "#0000FF",
    ANSI256:   "86",
    ANSI:      "5",
}

// Complete adaptive colors
lipgloss.CompleteAdaptiveColor{
    Light: lipgloss.CompleteColor{TrueColor: "#d7ffae", ANSI256: "193", ANSI: "11"},
    Dark:  lipgloss.CompleteColor{TrueColor: "#d75fee", ANSI256: "163", ANSI: "5"},
}
```

## Text Formatting

```go
var style = lipgloss.NewStyle().
    Bold(true).
    Italic(true).
    Underline(true).
    Strikethrough(true).
    Blink(true).
    Faint(true).
    Reverse(true)  // Swap foreground/background
```

## Block-Level Styling

### Padding

```go
// Individual sides
style.PaddingTop(2).
    PaddingRight(4).
    PaddingBottom(2).
    PaddingLeft(4)

// Shorthand (like CSS)
style.Padding(2)           // All sides
style.Padding(2, 4)        // Vertical, horizontal
style.Padding(1, 4, 2)     // Top, horizontal, bottom
style.Padding(2, 4, 3, 1)  // Top, right, bottom, left (clockwise)
```

### Margins

```go
// Individual sides
style.MarginTop(2).
    MarginRight(4).
    MarginBottom(2).
    MarginLeft(4)

// Shorthand (same as padding)
style.Margin(2)
style.Margin(2, 4)
style.Margin(1, 4, 2)
style.Margin(2, 4, 3, 1)
```

### Width and Height

```go
style.Width(80).Height(24)

// Max constraints
style.MaxWidth(100).MaxHeight(50)

// Inline rendering (ignores margins, padding, borders)
style.Inline(true).Render("compact")
```

## Borders

### Built-in Border Styles

```go
lipgloss.NormalBorder()
lipgloss.RoundedBorder()
lipgloss.ThickBorder()
lipgloss.DoubleBorder()
lipgloss.HiddenBorder()
lipgloss.BlockBorder()
lipgloss.OuterHalfBlockBorder()
lipgloss.InnerHalfBlockBorder()
```

### Applying Borders

```go
// All sides
style.Border(lipgloss.RoundedBorder())

// Specific sides
style.Border(lipgloss.NormalBorder()).
    BorderTop(true).
    BorderLeft(true).
    BorderRight(false).
    BorderBottom(false)

// Shorthand
style.Border(lipgloss.ThickBorder(), true, false)  // Top/bottom only
style.Border(lipgloss.DoubleBorder(), true, false, false, true)  // Top and left
```

### Border Colors

```go
style.BorderForeground(lipgloss.Color("62")).
    BorderBackground(lipgloss.Color("235"))
```

### Custom Borders

```go
myBorder := lipgloss.Border{
    Top:         "._.:*:",
    Bottom:      "._.:*:",
    Left:        "|*",
    Right:       "|*",
    TopLeft:     "*",
    TopRight:    "*",
    BottomLeft:  "*",
    BottomRight: "*",
}

style.Border(myBorder)
```

## Text Alignment

```go
style.Align(lipgloss.Left)
style.Align(lipgloss.Center)
style.Align(lipgloss.Right)

// Vertical alignment (for Height-constrained styles)
style.AlignVertical(lipgloss.Top)
style.AlignVertical(lipgloss.Center)
style.AlignVertical(lipgloss.Bottom)
```

## Layout Functions

### Horizontal Join

```go
// Join strings horizontally
lipgloss.JoinHorizontal(
    lipgloss.Top,     // Align to top
    columnA,
    columnB,
    columnC,
)

// Alignment positions
lipgloss.Top
lipgloss.Center
lipgloss.Bottom
lipgloss.Position(0.2)  // 20% from top
```

### Vertical Join

```go
// Join strings vertically
lipgloss.JoinVertical(
    lipgloss.Left,    // Align to left
    rowA,
    rowB,
    rowC,
)

// Alignment positions
lipgloss.Left
lipgloss.Center
lipgloss.Right
lipgloss.Position(0.5)  // 50% from left (centered)
```

### Place in Whitespace

```go
// Center content in a specific area
block := lipgloss.PlaceHorizontal(
    80,                // Width
    lipgloss.Center,   // Position
    content,
)

block := lipgloss.PlaceVertical(
    24,                // Height
    lipgloss.Bottom,   // Position
    content,
)

// Both dimensions
block := lipgloss.Place(
    80, 24,                      // Width, height
    lipgloss.Center,             // Horizontal position
    lipgloss.Center,             // Vertical position
    content,
    lipgloss.WithWhitespaceChars("."),  // Custom whitespace
    lipgloss.WithWhitespaceForeground(lipgloss.Color("240")),
)
```

## Measuring Content

```go
// Get dimensions
width := lipgloss.Width(renderedString)
height := lipgloss.Height(renderedString)

// Both at once
w, h := lipgloss.Size(renderedString)
```

## Style Composition

### Copying and Inheritance

```go
baseStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("63"))

// True copy (assignment creates new instance)
copiedStyle := baseStyle

// Copy with modifications
boldStyle := baseStyle.Bold(true)

// Inheritance (only inherit unset rules)
childStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("201")).  // This won't be inherited
    Inherit(baseStyle)                   // Only background is inherited
```

### Unsetting Rules

```go
style := lipgloss.NewStyle().
    Bold(true).
    UnsetBold().                       // Remove bold
    Background(lipgloss.Color("227")).
    UnsetBackground()                  // Remove background
```

## Tab Handling

```go
style := lipgloss.NewStyle()  // Default: tabs = 4 spaces

style.TabWidth(2)                         // Tabs = 2 spaces
style.TabWidth(0)                         // Remove tabs
style.TabWidth(lipgloss.NoTabConversion)  // Leave tabs intact
```

## Rendering Tables

```go
import "github.com/charmbracelet/lipgloss/table"

rows := [][]string{
    {"Chinese", "您好", "你好"},
    {"Japanese", "こんにちは", "やあ"},
    {"Arabic", "أهلين", "أهلا"},
}

t := table.New().
    Border(lipgloss.NormalBorder()).
    BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
    Headers("LANGUAGE", "FORMAL", "INFORMAL").
    Rows(rows...)

// Style function for alternating rows
t.StyleFunc(func(row, col int) lipgloss.Style {
    if row == table.HeaderRow {
        return headerStyle
    }
    if row%2 == 0 {
        return evenRowStyle
    }
    return oddRowStyle
})

fmt.Println(t)
```

### Table Borders

```go
// Markdown style
table.New().Border(lipgloss.MarkdownBorder())

// ASCII style
table.New().Border(lipgloss.ASCIIBorder())
```

## Rendering Lists

```go
import "github.com/charmbracelet/lipgloss/list"

l := list.New("A", "B", "C")
fmt.Println(l)
// • A
// • B
// • C

// Nested lists
l := list.New(
    "A", list.New("Apple"),
    "B", list.New("Banana", "Blueberry"),
)

// Custom enumerators
l.Enumerator(list.Arabic)    // 1. 2. 3.
l.Enumerator(list.Alphabet)  // A. B. C.
l.Enumerator(list.Roman)     // I. II. III.
l.Enumerator(list.Bullet)    // • • •
l.Enumerator(list.Tree)      // ├── └──

// Custom styles
l.EnumeratorStyle(enumStyle).
  ItemStyle(itemStyle)

// Custom enumerator function
func myEnum(items list.Items, i int) string {
    return fmt.Sprintf("[%d]", i+1)
}
l.Enumerator(myEnum)
```

## Rendering Trees

```go
import "github.com/charmbracelet/lipgloss/tree"

t := tree.Root(".").
    Child("src").
    Child("tests").
    Child("README.md")

// Nested trees
t := tree.Root(".").
    Child("src").
    Child(
        tree.New().
            Root("tests").
            Child("unit").
            Child("integration"),
    )

// Enumerators
t.Enumerator(tree.DefaultEnumerator)
t.Enumerator(tree.RoundedEnumerator)

// Styles
t.RootStyle(rootStyle).
  ItemStyle(itemStyle).
  EnumeratorStyle(enumStyle)
```

## Custom Renderers

```go
// Create renderer for specific output
renderer := lipgloss.NewRenderer(writer)

// Use renderer to create styles
style := renderer.NewStyle().
    Background(lipgloss.AdaptiveColor{Light: "63", Dark: "228"})

// Useful for SSH apps with different clients
func handleSSHSession(sess ssh.Session) {
    renderer := lipgloss.NewRenderer(sess)
    style := renderer.NewStyle().Foreground(lipgloss.Color("170"))
    io.WriteString(sess, style.Render("Hello!"))
}
```

## Common Patterns

### Card/Panel Component

```go
func RenderCard(title, content string) string {
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("170")).
        Padding(0, 1)

    cardStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 2).
        Width(50)

    return cardStyle.Render(
        lipgloss.JoinVertical(
            lipgloss.Left,
            titleStyle.Render(title),
            "",
            content,
        ),
    )
}
```

### Status Bar

```go
func RenderStatusBar(width int, left, right string) string {
    style := lipgloss.NewStyle().
        Background(lipgloss.Color("236")).
        Foreground(lipgloss.Color("250")).
        Width(width)

    gap := width - lipgloss.Width(left) - lipgloss.Width(right)
    return style.Render(left + strings.Repeat(" ", gap) + right)
}
```

### Two-Column Layout

```go
func TwoColumns(left, right string, width int) string {
    leftCol := lipgloss.NewStyle().
        Width(width / 2).
        Padding(1).
        Render(left)

    rightCol := lipgloss.NewStyle().
        Width(width / 2).
        Padding(1).
        Render(right)

    return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}
```

### Responsive Layout

```go
func RenderResponsive(width int, items []string) string {
    if width < 80 {
        // Stack vertically for narrow screens
        return lipgloss.JoinVertical(lipgloss.Left, items...)
    }
    // Arrange horizontally for wide screens
    return lipgloss.JoinHorizontal(lipgloss.Top, items...)
}
```

## Performance Tips

1. **Reuse Styles**: Create styles once, reuse them
2. **Cache Rendered Strings**: Store static content
3. **Measure Before Placing**: Use `lipgloss.Width/Height` for dynamic layouts
4. **Avoid Nested Renders**: Compose strings, then render once

## Best Practices

1. **Define Styles as Variables**: Keep styling separate from logic
2. **Use Adaptive Colors**: Support both light and dark terminals
3. **Respect Terminal Width**: Check available space before rendering
4. **Create Style Constants**: Define your theme in one place
5. **Test Different Terminals**: Colors render differently across terminals

## Theme Example

```go
package styles

var (
    // Colors
    Primary   = lipgloss.Color("170")
    Secondary = lipgloss.Color("62")
    Success   = lipgloss.Color("42")
    Danger    = lipgloss.Color("196")
    Muted     = lipgloss.Color("240")

    // Base styles
    Bold = lipgloss.NewStyle().Bold(true)

    // Component styles
    Title = lipgloss.NewStyle().
        Bold(true).
        Foreground(Primary).
        MarginBottom(1)

    Card = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(Secondary).
        Padding(1, 2)

    Button = lipgloss.NewStyle().
        Background(Primary).
        Foreground(lipgloss.Color("15")).
        Padding(0, 2).
        Bold(true)
)
```

## Reference

- GitHub: https://github.com/charmbracelet/lipgloss
- Docs: https://pkg.go.dev/github.com/charmbracelet/lipgloss
- Examples: https://github.com/charmbracelet/lipgloss/tree/master/examples
