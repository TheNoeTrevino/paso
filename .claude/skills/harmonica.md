# Harmonica - Spring Animation Expert

Expert in Harmonica, the spring-based physics animation library for smooth, natural motion.

## When to Use This Skill

Use this skill when:
- Adding smooth animations to terminal UIs
- Creating physics-based transitions between states
- Animating values (positions, sizes, opacity, etc.)
- Building responsive, delightful user interactions
- Implementing smooth scrolling or sliding effects
- Creating bouncy, natural-feeling animations

## Core Concepts

Harmonica implements damped spring physics to create natural-looking animations. Instead of linear tweening, springs accelerate and decelerate naturally, with optional bounce/overshoot.

## Basic Usage

```go
import "github.com/charmbracelet/harmonica"

// Create a spring with FPS, angular frequency, and damping
spring := harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.5)

// Animate a value from current to target
var position float64 = 0.0
var velocity float64 = 0.0
target := 100.0

// On each frame
position, velocity = spring.Update(position, velocity, target)
```

## Spring Parameters

### FPS (Frames Per Second)

```go
// Set the time step for updates
harmonica.FPS(60)   // 60 updates per second
harmonica.FPS(30)   // 30 updates per second
```

### Angular Frequency (ω)

Controls animation speed:

```go
harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.5)  // Medium speed
harmonica.NewSpring(harmonica.FPS(60), 12.0, 0.5) // Fast
harmonica.NewSpring(harmonica.FPS(60), 3.0, 0.5)  // Slow
```

Higher values = faster animations

### Damping Ratio (ζ)

Controls springiness:

```go
// Under-damped (< 1.0) - Bouncy, overshoots
harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.3)  // Very bouncy
harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.5)  // Medium bounce
harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.7)  // Slight bounce

// Critically-damped (= 1.0) - No overshoot, fastest to settle
harmonica.NewSpring(harmonica.FPS(60), 6.0, 1.0)  // Perfect damping

// Over-damped (> 1.0) - Slow, no overshoot
harmonica.NewSpring(harmonica.FPS(60), 6.0, 1.5)  // Sluggish
```

## Integration with Bubble Tea

### Basic Animation Loop

```go
type model struct {
    spring   harmonica.Spring
    position float64
    velocity float64
    target   float64
}

func initialModel() model {
    return model{
        spring:   harmonica.NewSpring(harmonica.FPS(60), 10.0, 0.8),
        position: 0.0,
        velocity: 0.0,
        target:   0.0,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "right":
            m.target = 100.0
            return m, tick()
        }

    case tickMsg:
        // Animate toward target
        m.position, m.velocity = m.spring.Update(m.position, m.velocity, m.target)

        // Continue animation if still moving
        if !isSettled(m.position, m.target, m.velocity) {
            return m, tick()
        }
    }

    return m, nil
}

type tickMsg time.Time

func tick() tea.Cmd {
    return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func isSettled(pos, target, vel float64) bool {
    return math.Abs(pos-target) < 0.01 && math.Abs(vel) < 0.01
}

func (m model) View() string {
    // Use animated position to render
    spaces := int(m.position)
    return strings.Repeat(" ", spaces) + "█"
}
```

### Smooth Scrolling Viewport

```go
type model struct {
    spring       harmonica.Spring
    scrollOffset float64  // Animated value
    scrollVel    float64  // Velocity
    targetOffset float64  // Target value
    items        []string
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "down":
            m.targetOffset = min(m.targetOffset+1, float64(len(m.items)-10))
            return m, tick()
        case "up":
            m.targetOffset = max(m.targetOffset-1, 0)
            return m, tick()
        }

    case tickMsg:
        m.scrollOffset, m.scrollVel = m.spring.Update(
            m.scrollOffset,
            m.scrollVel,
            m.targetOffset,
        )

        if !isSettled(m.scrollOffset, m.targetOffset, m.scrollVel) {
            return m, tick()
        }
    }

    return m, nil
}

func (m model) View() string {
    // Render visible items based on animated offset
    start := int(m.scrollOffset)
    end := min(start+10, len(m.items))

    var s string
    for i := start; i < end; i++ {
        s += m.items[i] + "\n"
    }
    return s
}
```

### Sliding Panels

```go
type model struct {
    spring   harmonica.Spring
    panelX   float64  // Current X position
    panelVel float64
    visible  bool
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "tab" {
            m.visible = !m.visible
            if m.visible {
                m.targetX = 0.0  // Slide in
            } else {
                m.targetX = -50.0  // Slide out
            }
            return m, tick()
        }

    case tickMsg:
        m.panelX, m.panelVel = m.spring.Update(m.panelX, m.panelVel, m.targetX)

        if !isSettled(m.panelX, m.targetX, m.panelVel) {
            return m, tick()
        }
    }

    return m, nil
}
```

### Progress Bar Animation

```go
type model struct {
    spring   harmonica.Spring
    progress float64  // 0.0 to 1.0
    velocity float64
    target   float64
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case progressUpdateMsg:
        m.target = msg.percent
        return m, tick()

    case tickMsg:
        m.progress, m.velocity = m.spring.Update(m.progress, m.velocity, m.target)

        if !isSettled(m.progress, m.target, m.velocity) {
            return m, tick()
        }
    }

    return m, nil
}

func (m model) View() string {
    width := 40
    filled := int(m.progress * float64(width))
    bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
    return fmt.Sprintf("[%s] %.0f%%", bar, m.progress*100)
}
```

### Multiple Springs

Animate multiple values independently:

```go
type model struct {
    xSpring harmonica.Spring
    ySpring harmonica.Spring
    x, xVel float64
    y, yVel float64
    targetX, targetY float64
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        if msg.Type == tea.MouseLeft {
            m.targetX = float64(msg.X)
            m.targetY = float64(msg.Y)
            return m, tick()
        }

    case tickMsg:
        m.x, m.xVel = m.xSpring.Update(m.x, m.xVel, m.targetX)
        m.y, m.yVel = m.ySpring.Update(m.y, m.yVel, m.targetY)

        if !isSettled(m.x, m.targetX, m.xVel) || !isSettled(m.y, m.targetY, m.yVel) {
            return m, tick()
        }
    }

    return m, nil
}
```

## Animation Presets

### Snappy (Fast, No Bounce)

```go
harmonica.NewSpring(harmonica.FPS(60), 12.0, 1.0)
```

### Bouncy

```go
harmonica.NewSpring(harmonica.FPS(60), 8.0, 0.6)
```

### Smooth and Professional

```go
harmonica.NewSpring(harmonica.FPS(60), 10.0, 0.9)
```

### Slow and Gentle

```go
harmonica.NewSpring(harmonica.FPS(60), 4.0, 1.0)
```

### Playful

```go
harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.4)
```

## Common Patterns

### Toggle Animation

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == " " {
            // Toggle between 0 and 100
            if m.target == 0 {
                m.target = 100
            } else {
                m.target = 0
            }
            return m, tick()
        }
    }
    return m, nil
}
```

### Interpolated Colors

```go
// Animate between two colors
func interpolateColor(from, to lipgloss.Color, t float64) lipgloss.Color {
    // Parse colors as RGB
    r1, g1, b1 := parseRGB(from)
    r2, g2, b2 := parseRGB(to)

    // Interpolate
    r := int(float64(r1) + (float64(r2-r1) * t))
    g := int(float64(g1) + (float64(g2-g1) * t))
    b := int(float64(b1) + (float64(b2-b1) * t))

    return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// In View:
color := interpolateColor(startColor, endColor, m.progress)
style := lipgloss.NewStyle().Foreground(color)
```

### Staggered Animations

Animate multiple items with delays:

```go
type item struct {
    spring   harmonica.Spring
    position float64
    velocity float64
    target   float64
    delay    int  // Frames to wait
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tickMsg:
        animating := false

        for i := range m.items {
            if m.items[i].delay > 0 {
                m.items[i].delay--
                animating = true
                continue
            }

            m.items[i].position, m.items[i].velocity = m.items[i].spring.Update(
                m.items[i].position,
                m.items[i].velocity,
                m.items[i].target,
            )

            if !isSettled(m.items[i].position, m.items[i].target, m.items[i].velocity) {
                animating = true
            }
        }

        if animating {
            return m, tick()
        }
    }
    return m, nil
}
```

## Performance Considerations

1. **Stop When Settled**: Check if animation is complete to avoid unnecessary updates
2. **Single Tick Command**: Share one ticker across all springs
3. **Threshold Values**: Use reasonable thresholds for "settled" detection
4. **Fixed Timestep**: Use consistent FPS for predictable behavior
5. **Batch Updates**: Update all springs in a single tick handler

## Best Practices

1. **Choose Appropriate Damping**:
   - Use 0.6-0.8 for UI elements (slight bounce)
   - Use 1.0 for smooth, professional feel
   - Avoid < 0.3 (too bouncy) or > 1.5 (too slow)

2. **Match FPS to Display**: Use 60 FPS for smooth terminal animations

3. **Test on Different Terminals**: Some terminals render slower than others

4. **Provide Instant Option**: Allow users to disable animations

5. **Don't Overuse**: Animate only important transitions

## Debugging

```go
// Log spring values
log.Printf("pos: %.2f, vel: %.2f, target: %.2f", position, velocity, target)

// Visualize velocity
func (m model) View() string {
    velocityBar := strings.Repeat("=", int(math.Abs(m.velocity)))
    return fmt.Sprintf("Position: %.2f\nVelocity: %s\n", m.position, velocityBar)
}
```

## Use Cases

1. **Smooth Scrolling**: Lists, viewports, pagers
2. **Panel Transitions**: Sliding drawers, modals
3. **Progress Indicators**: Loading bars, downloads
4. **Value Changes**: Numbers, percentages, scores
5. **Layout Transitions**: Resizing, repositioning
6. **Focus Indicators**: Highlighting selected items
7. **Menu Animations**: Expanding/collapsing menus

## Reference

- GitHub: https://github.com/charmbracelet/harmonica
- Docs: https://pkg.go.dev/github.com/charmbracelet/harmonica
- Examples: https://github.com/charmbracelet/harmonica/tree/master/examples
- Article: https://www.ryanjuckett.com/damped-springs/
