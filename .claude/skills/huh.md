# Huh - Interactive Forms Expert

Expert in Huh, the library for building interactive forms and prompts in the terminal.

## When to Use This Skill

Use this skill when:
- Creating interactive forms for user input
- Building prompts for CLI tools
- Collecting structured data from users
- Creating wizards or multi-step input flows
- Validating user input in terminal applications
- Building accessible terminal interfaces

## Core Concepts

Huh separates forms into **groups** (pages) made of **fields** (inputs). Forms can be run standalone or integrated into Bubble Tea applications.

## Basic Form Structure

```go
import "github.com/charmbracelet/huh"

var name string
var confirmed bool

form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("What's your name?").
            Value(&name),

        huh.NewConfirm().
            Title("Are you sure?").
            Value(&confirmed),
    ),
)

err := form.Run()
```

## Field Types

### Input (Single-line Text)

```go
var name string

huh.NewInput().
    Title("What's your name?").
    Prompt("?").
    Placeholder("Enter your name...").
    Value(&name).
    Validate(func(str string) error {
        if len(str) < 3 {
            return errors.New("name must be at least 3 characters")
        }
        return nil
    })
```

### Text (Multi-line Text)

```go
var story string

huh.NewText().
    Title("Tell me a story").
    Placeholder("Once upon a time...").
    CharLimit(500).
    Value(&story).
    Validate(func(str string) error {
        if len(str) < 10 {
            return errors.New("story too short")
        }
        return nil
    })
```

### Select (Single Choice)

```go
var country string

huh.NewSelect[string]().
    Title("Pick a country").
    Options(
        huh.NewOption("United States", "US"),
        huh.NewOption("Germany", "DE"),
        huh.NewOption("Brazil", "BR"),
        huh.NewOption("Canada", "CA"),
    ).
    Value(&country)

// Generic type matches stored value type
huh.NewSelect[int]().
    Title("How much sauce?").
    Options(
        huh.NewOption("None", 0),
        huh.NewOption("A little", 1),
        huh.NewOption("A lot", 2),
    ).
    Value(&sauceLevel)
```

### MultiSelect (Multiple Choices)

```go
var toppings []string

huh.NewMultiSelect[string]().
    Title("Toppings").
    Options(
        huh.NewOption("Lettuce", "lettuce").Selected(true),
        huh.NewOption("Tomatoes", "tomatoes").Selected(true),
        huh.NewOption("Cheese", "cheese"),
        huh.NewOption("Jalapeños", "jalapeños"),
    ).
    Limit(4).  // Maximum selections
    Value(&toppings)
```

### Confirm (Yes/No)

```go
var confirmed bool

huh.NewConfirm().
    Title("Are you sure?").
    Affirmative("Yes!").
    Negative("No.").
    Value(&confirmed)
```

## Multi-page Forms

```go
form := huh.NewForm(
    // First page
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Choose burger").
            Options(/* ... */).
            Value(&burger),

        huh.NewMultiSelect[string]().
            Title("Toppings").
            Options(/* ... */).
            Value(&toppings),
    ),

    // Second page
    huh.NewGroup(
        huh.NewInput().
            Title("What's your name?").
            Value(&name),

        huh.NewText().
            Title("Special instructions").
            Value(&instructions),
    ),
)

err := form.Run()
```

## Validation

```go
huh.NewInput().
    Value(&email).
    Validate(func(str string) error {
        if !strings.Contains(str, "@") {
            return errors.New("invalid email")
        }
        return nil
    })

// Validation runs when user tries to proceed
// Form won't advance if validation fails
```

## Shorthand: Quick Prompts

```go
// Run a single field directly
var name string
huh.NewInput().
    Title("What's your name?").
    Value(&name).
    Run()  // Blocking

fmt.Printf("Hello, %s!\n", name)
```

## Dynamic Forms

Forms that change based on previous input:

```go
var country string
var state string

form := huh.NewForm(
    huh.NewGroup(
        // First select
        huh.NewSelect[string]().
            Options(huh.NewOptions("United States", "Canada", "Mexico")...).
            Value(&country).
            Title("Country"),

        // Second select with dynamic title and options
        huh.NewSelect[string]().
            Value(&state).
            TitleFunc(func() string {
                switch country {
                case "United States":
                    return "State"
                case "Canada":
                    return "Province"
                default:
                    return "Territory"
                }
            }, &country).  // Recompute when country changes
            OptionsFunc(func() []huh.Option[string] {
                states := fetchStatesForCountry(country)  // API call
                return huh.NewOptions(states...)
            }, &country),  // Recompute when country changes
    ),
)
```

### Dynamic Field Functions

```go
// TitleFunc - Dynamic title
TitleFunc(func() string { return computedTitle }, &dependency)

// OptionsFunc - Dynamic options
OptionsFunc(func() []huh.Option[T] { return computedOptions }, &dependency)

// DescriptionFunc - Dynamic description
DescriptionFunc(func() string { return computedDesc }, &dependency)
```

## Accessibility Mode

```go
accessibleMode := os.Getenv("ACCESSIBLE") != ""
form.WithAccessible(accessibleMode)

// Accessible mode uses standard prompts instead of TUI
// Better for screen readers
```

## Themes

```go
form.WithTheme(huh.ThemeCharm())
form.WithTheme(huh.ThemeDracula())
form.WithTheme(huh.ThemeCatppuccin())
form.WithTheme(huh.ThemeBase16())
form.WithTheme(huh.ThemeDefault())

// Custom theme
customTheme := huh.ThemeCharm()
customTheme.Focused.Base = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
form.WithTheme(customTheme)
```

## Integration with Bubble Tea

> **IMPORTANT: Value Binding Caveat**
>
> When using huh forms with Bubble Tea, **DO NOT rely on pointer bindings** (`.Value(&myVar)`) to read form values after submission. Bubble Tea passes models by value, so the pointers end up pointing to stale copies of your model's fields.
>
> **Always use `GetString()`, `GetBool()`, `GetInt()`, or `Get()` to read values from the form:**
>
> ```go
> // WRONG - pointer bindings are stale in Bubble Tea
> if m.form.State == huh.StateCompleted {
>     fmt.Println(m.myBoundVar)  // May have old/wrong value!
> }
>
> // CORRECT - read directly from form using keys
> if m.form.State == huh.StateCompleted {
>     title := m.form.GetString("title")
>     confirmed := m.form.GetBool("confirm")
>     // For generic types, use Get() with type assertion:
>     if ids, ok := m.form.Get("labels").([]int); ok {
>         // use ids
>     }
> }
> ```
>
> This means you must assign `.Key("fieldname")` to every field you want to retrieve later.

```go
type Model struct {
    form *huh.Form  // Huh form is a tea.Model
}

func NewModel() Model {
    return Model{
        form: huh.NewForm(
            huh.NewGroup(
                huh.NewSelect[string]().
                    Key("class").  // Key is required for retrieval!
                    Options(huh.NewOptions("Warrior", "Mage", "Rogue")...).
                    Title("Choose your class"),
            ),
        ),
    }
}

func (m Model) Init() tea.Cmd {
    return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    // Delegate to form
    form, cmd := m.form.Update(msg)
    if f, ok := form.(*huh.Form); ok {
        m.form = f
    }

    return m, cmd
}

func (m Model) View() string {
    // Check if form is completed
    if m.form.State == huh.StateCompleted {
        // Use GetString, not bound variables!
        class := m.form.GetString("class")
        return fmt.Sprintf("You selected: %s", class)
    }
    return m.form.View()
}
```

## Getting Form Values

```go
// Using field keys
form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Key("name").
            Value(&name),
        huh.NewSelect[int]().
            Key("level").
            Value(&level),
    ),
)

// Retrieve by key
name := form.GetString("name")
level := form.GetInt("level")
```

## Form State

```go
form.State == huh.StateNormal      // Form is running
form.State == huh.StateCompleted   // Form finished
form.State == huh.StateAborted     // User cancelled (Ctrl+C)
```

## Bonus: Spinner

Huh ships with a standalone spinner for background tasks:

```go
import "github.com/charmbracelet/huh/spinner"

// Action style
err := spinner.New().
    Title("Making your burger...").
    Action(makeBurger).
    Run()

// Context style
ctx, cancel := context.WithCancel(context.Background())
go makeBurger(ctx)

err := spinner.New().
    Title("Making your burger...").
    Context(ctx).
    Run()

fmt.Println("Done!")
```

## Common Patterns

### Multi-step Wizard

```go
var (
    projectName string
    framework   string
    features    []string
)

wizard := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("Project name").
            Value(&projectName),
    ),
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Choose framework").
            Options(
                huh.NewOption("React", "react"),
                huh.NewOption("Vue", "vue"),
                huh.NewOption("Svelte", "svelte"),
            ).
            Value(&framework),
    ),
    huh.NewGroup(
        huh.NewMultiSelect[string]().
            Title("Select features").
            Options(
                huh.NewOption("TypeScript", "ts"),
                huh.NewOption("Testing", "test"),
                huh.NewOption("Linting", "lint"),
            ).
            Value(&features),
    ),
)

wizard.Run()
```

### Conditional Fields

```go
var (
    useDatabase bool
    dbType      string
)

form := huh.NewForm(
    huh.NewGroup(
        huh.NewConfirm().
            Title("Use database?").
            Value(&useDatabase),
    ),
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Database type").
            Options(
                huh.NewOption("PostgreSQL", "postgres"),
                huh.NewOption("MySQL", "mysql"),
            ).
            Value(&dbType),
    ).WithHideFunc(func() bool {
        return !useDatabase  // Hide if no database
    }),
)
```

### Form with Progress

```go
func (m Model) View() string {
    if m.form.State == huh.StateCompleted {
        return "✓ Setup complete!"
    }

    // Show progress
    current := m.form.GetCurrentGroup()
    total := m.form.GetGroupCount()
    progress := fmt.Sprintf("Step %d/%d", current+1, total)

    return lipgloss.JoinVertical(
        lipgloss.Left,
        progress,
        "",
        m.form.View(),
    )
}
```

## Validation Examples

```go
// Email validation
Validate(func(s string) error {
    if !strings.Contains(s, "@") {
        return errors.New("invalid email")
    }
    return nil
})

// Length validation
Validate(func(s string) error {
    if len(s) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    return nil
})

// Async validation (e.g., check username availability)
Validate(func(s string) error {
    available, err := checkUsername(s)
    if err != nil {
        return err
    }
    if !available {
        return errors.New("username already taken")
    }
    return nil
})
```

## Best Practices

1. **Use Groups for Pagination**: Break long forms into logical pages
2. **Validate Early**: Provide validation feedback before submission
3. **Set Placeholders**: Help users understand expected input
4. **Limit Multi-selects**: Don't overwhelm users with too many options
5. **Use Accessible Mode**: Support screen readers with environment variable
6. **Provide Defaults**: Pre-select common choices
7. **Show Progress**: Indicate current step in multi-page forms

## Reference

- GitHub: https://github.com/charmbracelet/huh
- Docs: https://pkg.go.dev/github.com/charmbracelet/huh
- Examples: https://github.com/charmbracelet/huh/tree/main/examples
