# Paso - Terminal Kanban Board Implementation Guide

## Overview

**Paso** (Spanish for "step") is a zero-setup, terminal-based kanban board for personal task management. It brings the visual organization of tools like Jira into the terminal with a beautiful, performant interface that requires no configuration - just run `paso` and start managing your tasks.

### Purpose

Paso solves the problem of heavyweight project management tools for personal use. While Jira excels for team collaboration, it's overkill for individual developers who want to:
- Visualize task progression through workflow stages
- Quickly capture and organize work
- Stay in the terminal without context switching to web UIs
- Own their data locally without cloud dependencies

Unlike Jira, Paso is:
- **Instant**: No setup, no configuration, no server
- **Local**: All data stored in a single SQLite file
- **Fast**: Native performance, sub-millisecond response times
- **Beautiful**: Smooth animations and polished terminal UI
- **Portable**: Single binary, works anywhere

### Comparison to Jira

| Feature | Jira | Paso |
|---------|------|------|
| Setup Time | Hours (server, config, users) | 0 seconds (just run it) |
| Data Location | Cloud/Server | Local SQLite file |
| UI | Web browser | Terminal (TUI) |
| Performance | Network-dependent, can be slow | Native, instant |
| Complexity | Hundreds of features | Focused on core kanban workflow |
| Cost | $7-14/user/month | Free, open source |
| Target Audience | Teams | Individual developers |

## Tech Stack

### Core Technologies

**Go** - Primary programming language
- Fast compilation and runtime performance
- Single binary distribution (compile anywhere with modernc.org/sqlite)
- Excellent standard library for file I/O and concurrency
- Strong typing for reliable refactoring

**Bubble Tea** - TUI framework (https://github.com/charmbracelet/bubbletea)
- Elm-inspired architecture (Model-View-Update pattern)
- Handles keyboard input, terminal rendering, and state management
- Battle-tested by Charm's ecosystem of tools
- Clean separation of concerns

**Lipgloss** - Styling and layout (https://github.com/charmbracelet/lipgloss)
- CSS-like styling for terminal output
- Layout primitives (JoinHorizontal, JoinVertical)
- Border styles, padding, margins, colors
- Composable style definitions

**Harmonica** - Animation library (https://github.com/charmbracelet/harmonica)
- Spring-based physics animations
- Smooth transitions between states
- Makes UI feel polished and responsive
- 60 FPS animations in the terminal

**modernc.org/sqlite** - Pure Go SQLite implementation
- No CGo dependencies (compiles anywhere)
- Embedded database (zero setup)
- ACID transactions for data integrity
- Fast queries with indexes

### Why These Choices?

- **Go + Pure Go SQLite**: Single binary with no runtime dependencies, works on any platform
- **Bubble Tea**: Proven TUI framework with excellent developer experience
- **Lipgloss**: Makes terminal layouts manageable without manual positioning
- **Harmonica**: Differentiates Paso from basic TUIs with delightful animations
- **SQLite**: Relational data modeling without setup complexity

## Visual Design

### Kanban Board Layout

```
┌────────────────────────────────────────────────────────────────────────────┐
│                              PASO - Your Tasks                             │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐          │
│  │   📋 Todo        │  │  🔨 In Progress  │  │   ✅ Done        │          │
│  ├──────────────────┤  ├──────────────────┤  ├──────────────────┤          │
│  │                  │  │                  │  │                  │          │
│  │ ▸ Fix auth bug   │  │ ▸ Add tests      │  │   Deploy v1.0    │          │
│  │   PASO-1         │  │   PASO-3         │  │   PASO-5         │          │
│  │                  │  │                  │  │                  │          │
│  │   Refactor UI    │  │   Review PR #42  │  │   Hotfix prod    │          │
│  │   PASO-2         │  │   PASO-4         │  │   PASO-6         │          │
│  │                  │  │                  │  │                  │          │
│  │   Update deps    │  │                  │  │                  │          │
│  │   PASO-7         │  │                  │  │                  │          │
│  │                  │  │                  │  │                  │          │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘          │
│                                                                            │
├────────────────────────────────────────────────────────────────────────────┤
│ [a]dd  [e]dit  [d]elete  [←→] move  [hjkl] navigate  [q]uit                │
└────────────────────────────────────────────────────────────────────────────┘
```

### Column Viewport Scrolling

When you have more than 3 columns, the viewport slides horizontally:

```
Initial view:
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   📋 Todo    │  │ 🔨 Progress  │  │  ✅ Done     │
└──────────────┘  └──────────────┘  └──────────────┘

Press ] to scroll right:
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ 🔨 Progress  │  │  ✅ Done     │  │  📦 Archive  │
└──────────────┘  └──────────────┘  └──────────────┘

Press [ to scroll left (back to start):
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   📋 Todo    │  │ 🔨 Progress  │  │  ✅ Done     │
└──────────────┘  └──────────────┘  └──────────────┘
```

If possible, make this smooth using the Harmonica animation library to slide
columns in and out of view.

### Task Movement Flow

```
1. Select task with j/k (up/down)
┌──────────────┐
│   📋 Todo    │
├──────────────┤
│ ▸ Fix auth   │  ← Selected (cursor here)
│   PASO-1     │
│              │
│   Refactor   │
│   PASO-2     │
└──────────────┘

2. Press > to move right (or h/l to change columns, then Enter to drop)
┌──────────────┐  ┌──────────────┐
│   📋 Todo    │  │ 🔨 Progress  │
├──────────────┤  ├──────────────┤
│              │  │ ▸ Fix auth   │  ← Task moved!
│   Refactor   │  │   PASO-1     │
│   PASO-2     │  │              │
└──────────────┘  └──────────────┘
```

### Data Model

```
┌─────────────┐
│   columns   │
├─────────────┤
│ id          │
│ name        │
│ position    │───┐
└─────────────┘   │
                  │
                  │ 1:N
                  │
┌─────────────┐   │
│    tasks    │◄──┘
├─────────────┤
│ id          │
│ title       │
│ description │
│ column_id   │ (FK)
│ position    │ (order within column)
│ created_at  │
│ updated_at  │
└─────────────┘
```

## Implementation Roadmap

This guide breaks down Paso's implementation into 10 logical phases. Each phase builds on the previous and produces a working, testable application. Complete each phase fully before moving to the next.

---

## Phase 1: Project Setup & Database Foundation

**Goal**: Establish project structure, initialize SQLite database, and create data models.

### Objectives
- Set up Go module and directory structure
- Initialize SQLite database with schema
- Create data access layer (repository pattern)
- Implement basic CRUD operations for columns and tasks

### Files to Create

```
paso/
├── main.go                    # Entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependencies
├── internal/
│   ├── database/
│   │   ├── db.go             # Database initialization
│   │   ├── migrations.go     # Schema definitions
│   │   └── repository.go     # CRUD operations
│   └── models/
│       ├── column.go         # Column struct
│       └── task.go           # Task struct
└── README.md
```

### Implementation Steps

1. **Initialize Go module**
   ```bash
   go mod init github.com/yourusername/paso
   go get modernc.org/sqlite
   ```

2. **Create models** (`internal/models/`)
   - Define `Column` struct with: ID, Name, Position
   - Define `Task` struct with: ID, Title, Description, ColumnID, Position, CreatedAt, UpdatedAt
   - Add methods for serialization/deserialization if needed

3. **Database initialization** (`internal/database/db.go`)
   - Function to open/create SQLite file at `~/.paso/tasks.db`
   - Ensure directory exists (`~/.paso/`)
   - Handle connection errors gracefully

4. **Schema migrations** (`internal/database/migrations.go`)
   - Create `columns` table with auto-increment ID
   - Create `tasks` table with foreign key to columns
   - Add indexes for common queries (column_id, position)
   - Function to run migrations (CREATE TABLE IF NOT EXISTS)

5. **Repository layer** (`internal/database/repository.go`)
   - `CreateColumn(name string, position int) (*Column, error)`
   - `GetAllColumns() ([]*Column, error)`
   - `CreateTask(title, description string, columnID, position int) (*Task, error)`
   - `GetTasksByColumn(columnID int) ([]*Task, error)`
   - `UpdateTaskColumn(taskID, newColumnID, newPosition int) error`
   - `DeleteTask(taskID int) error`

6. **Seed default columns** (`internal/database/migrations.go`)
   - On first run, create three default columns: "Todo", "In Progress", "Done"
   - Check if columns exist before seeding

### Testing Phase 1

Create `main.go` with a simple test:
```go
func main() {
    db, err := database.InitDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Test: Create a task in "Todo" column
    columns, _ := database.GetAllColumns(db)
    task, _ := database.CreateTask(db, "Test task", "", columns[0].ID, 0)
    fmt.Printf("Created task: %s (ID: %d)\n", task.Title, task.ID)
    
    // Test: Retrieve all tasks
    tasks, _ := database.GetTasksByColumn(db, columns[0].ID)
    fmt.Printf("Found %d tasks in Todo column\n", len(tasks))
}
```

**Definition of Done**:
- Running `paso` creates `~/.paso/tasks.db` automatically
- Database contains `columns` and `tasks` tables
- Three default columns exist (Todo, In Progress, Done)
- Can create, read, update, delete tasks via repository functions
- No errors or panics during database operations

---

## Phase 2: Bubble Tea Application Skeleton

**Goal**: Set up the Bubble Tea framework and create a minimal working TUI that displays "Hello, Paso!".

### Objectives
- Initialize Bubble Tea application structure
- Implement Model-View-Update pattern
- Handle basic keyboard input (quit on 'q')
- Render a simple screen

### Files to Create/Modify

```
paso/
├── main.go                    # Modified: Initialize Bubble Tea
└── internal/
    └── tui/
        ├── model.go          # Application state
        ├── update.go         # Event handlers
        └── view.go           # Rendering logic
```

### Implementation Steps

1. **Install Bubble Tea**
   ```bash
   go get github.com/charmbracelet/bubbletea
   ```

2. **Create TUI model** (`internal/tui/model.go`)
   ```go
   type Model struct {
       db       *sql.DB
       columns  []*models.Column
       tasks    map[int][]*models.Task  // columnID -> tasks
       width    int
       height   int
   }
   
   func InitialModel(db *sql.DB) Model {
       // Load columns and tasks from database
       return Model{db: db, ...}
   }
   ```

3. **Implement Update** (`internal/tui/update.go`)
   ```go
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case tea.KeyMsg:
           switch msg.String() {
           case "q", "ctrl+c":
               return m, tea.Quit
           }
       case tea.WindowSizeMsg:
           m.width = msg.Width
           m.height = msg.Height
       }
       return m, nil
   }
   ```

4. **Implement View** (`internal/tui/view.go`)
   ```go
   func (m Model) View() string {
       return "Hello, Paso!\n\nPress 'q' to quit."
   }
   ```

5. **Wire up main.go**
   ```go
   func main() {
       db, err := database.InitDB()
       if err != nil {
           log.Fatal(err)
       }
       defer db.Close()

       p := tea.NewProgram(tui.InitialModel(db))
       if _, err := p.Run(); err != nil {
           log.Fatal(err)
       }
   }
   ```

### Testing Phase 2

Run `paso`:
- Should see "Hello, Paso!" in terminal
- Pressing 'q' should exit cleanly
- Terminal should restore properly after exit

**Definition of Done**:
- Bubble Tea application runs without errors
- Can quit with 'q' or Ctrl+C
- Terminal state is properly restored on exit
- Model loads columns and tasks from database (even if not displayed yet)

---

## Phase 3: Static Kanban Board Rendering

**Goal**: Render all columns and tasks in a static kanban board layout using Lipgloss.

### Objectives
- Install and configure Lipgloss
- Create styled components for columns and tasks
- Layout columns horizontally
- Display tasks within their respective columns

### Files to Create/Modify

```
paso/
└── internal/
    └── tui/
        ├── styles.go         # Lipgloss style definitions
        ├── view.go           # Modified: Render kanban board
        └── components.go     # Reusable UI components
```

### Implementation Steps

1. **Install Lipgloss**
   ```bash
   go get github.com/charmbracelet/lipgloss
   ```

2. **Define styles** (`internal/tui/styles.go`)
   ```go
   var (
       ColumnStyle = lipgloss.NewStyle().
           Border(lipgloss.RoundedBorder()).
           BorderForeground(lipgloss.Color("62")).
           Padding(1).
           Width(30)
       
       TaskStyle = lipgloss.NewStyle().
           Padding(0, 1).
           MarginBottom(1)
       
       TitleStyle = lipgloss.NewStyle().
           Bold(true).
           Foreground(lipgloss.Color("170"))
   )
   ```

3. **Create component renderers** (`internal/tui/components.go`)
   ```go
   func RenderTask(task *models.Task) string {
       return TaskStyle.Render(fmt.Sprintf("▸ %s\n  PASO-%d", task.Title, task.ID))
   }
   
   func RenderColumn(column *models.Column, tasks []*models.Task) string {
       var taskViews []string
       for _, task := range tasks {
           taskViews = append(taskViews, RenderTask(task))
       }
       
       content := TitleStyle.Render(column.Name) + "\n\n"
       content += strings.Join(taskViews, "\n")
       
       return ColumnStyle.Render(content)
   }
   ```

4. **Update View** (`internal/tui/view.go`)
   ```go
   func (m Model) View() string {
       if m.width == 0 {
           return "Loading..."
       }
       
       var columns []string
       for _, col := range m.columns {
           tasks := m.tasks[col.ID]
           columns = append(columns, RenderColumn(col, tasks))
       }
       
       board := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
       
       header := TitleStyle.Render("PASO - Your Tasks")
       footer := "Press 'q' to quit"
       
       return lipgloss.JoinVertical(
           lipgloss.Left,
           header,
           "",
           board,
           "",
           footer,
       )
   }
   ```

### Testing Phase 3

1. Add some test tasks to the database (manually via SQL or a helper script)
2. Run `paso`
3. Verify:
   - Three columns displayed side-by-side
   - Tasks appear in correct columns
   - Styling is clean and readable
   - Board adjusts to terminal size

**Definition of Done**:
- Kanban board renders with all columns visible
- Tasks display with IDs and titles
- Columns have borders and proper spacing
- Layout is clean and professional-looking
- Board displays correctly on different terminal sizes

---

## Phase 4: Navigation System

**Goal**: Implement keyboard navigation to move between columns and tasks.

### Objectives
- Track cursor position (column and task)
- Implement hjkl/arrow key navigation
- Add visual indicators for selected column/task
- Handle edge cases (empty columns, boundaries)

### Files to Create/Modify

```
paso/
└── internal/
    └── tui/
        ├── model.go          # Add: selectedColumn, selectedTask
        ├── update.go         # Add: navigation handlers
        ├── view.go           # Add: selection highlighting
        └── components.go     # Add: selected state rendering
```

### Implementation Steps

1. **Update Model** (`internal/tui/model.go`)
   ```go
   type Model struct {
       db              *sql.DB
       columns         []*models.Column
       tasks           map[int][]*models.Task
       selectedColumn  int  // Index in columns slice
       selectedTask    int  // Index in tasks[currentColumnID] slice
       width           int
       height          int
   }
   ```

2. **Add navigation logic** (`internal/tui/update.go`)
   ```go
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case tea.KeyMsg:
           switch msg.String() {
           case "q", "ctrl+c":
               return m, tea.Quit
           
           case "h", "left":
               if m.selectedColumn > 0 {
                   m.selectedColumn--
                   m.selectedTask = 0  // Reset task selection
               }
           
           case "l", "right":
               if m.selectedColumn < len(m.columns)-1 {
                   m.selectedColumn++
                   m.selectedTask = 0
               }
           
           case "j", "down":
               currentCol := m.columns[m.selectedColumn]
               tasksInCol := m.tasks[currentCol.ID]
               if m.selectedTask < len(tasksInCol)-1 {
                   m.selectedTask++
               }
           
           case "k", "up":
               if m.selectedTask > 0 {
                   m.selectedTask--
               }
           }
       }
       return m, nil
   }
   ```

3. **Add selection highlighting** (`internal/tui/components.go`)
   ```go
   func RenderTask(task *models.Task, selected bool) string {
       style := TaskStyle
       if selected {
           style = style.
               Foreground(lipgloss.Color("170")).
               Bold(true).
               BorderLeft(true).
               BorderStyle(lipgloss.ThickBorder()).
               BorderForeground(lipgloss.Color("170"))
       }
       return style.Render(fmt.Sprintf("▸ %s\n  PASO-%d", task.Title, task.ID))
   }
   
   func RenderColumn(column *models.Column, tasks []*models.Task, selected bool, selectedTaskIdx int) string {
       style := ColumnStyle
       if selected {
           style = style.BorderForeground(lipgloss.Color("170"))
       }
       
       // ... render tasks with selection state
   }
   ```

4. **Update View to pass selection state** (`internal/tui/view.go`)
   ```go
   for i, col := range m.columns {
       tasks := m.tasks[col.ID]
       selectedTaskIdx := -1
       if i == m.selectedColumn {
           selectedTaskIdx = m.selectedTask
       }
       columns = append(columns, RenderColumn(col, tasks, i == m.selectedColumn, selectedTaskIdx))
   }
   ```

5. **Add help text** (`internal/tui/view.go`)
   ```go
   footer := "[hjkl/arrows] navigate  [q] quit"
   ```

### Testing Phase 4

Run `paso` and verify:
1. Initial cursor is on first column, first task
2. 'h' and 'l' move between columns
3. 'j' and 'k' move between tasks in a column
4. Selected column/task is visually highlighted
5. Can't move beyond boundaries (first/last column, first/last task)
6. Switching columns resets task selection to first task

**Definition of Done**:
- Can navigate between all columns using h/l or left/right arrows
- Can navigate between tasks within a column using j/k or up/down arrows
- Selected column and task are clearly highlighted
- Navigation respects boundaries (no crashes at edges)
- Empty columns can be selected but show no tasks

---

## Phase 5: Task CRUD Operations

**Goal**: Enable creating, editing, and deleting tasks through keyboard commands.

### Objectives
- Add task creation dialog
- Implement task editing
- Add task deletion with confirmation
- Update database and refresh view after operations

### Files to Create/Modify

```
paso/
└── internal/
    └── tui/
        ├── model.go          # Add: mode (normal/input), inputBuffer
        ├── update.go         # Add: CRUD command handlers
        ├── view.go           # Add: input dialog rendering
        └── input.go          # New: Input handling logic
```

### Implementation Steps

1. **Extend Model with input mode** (`internal/tui/model.go`)
   ```go
   type Mode int
   const (
       NormalMode Mode = iota
       AddTaskMode
       EditTaskMode
       DeleteConfirmMode
   )
   
   type Model struct {
       // ... existing fields
       mode         Mode
       inputBuffer  string
       inputPrompt  string
   }
   ```

2. **Add task creation** (`internal/tui/update.go`)
   ```go
   case tea.KeyMsg:
       if m.mode == NormalMode {
           switch msg.String() {
           case "a":
               m.mode = AddTaskMode
               m.inputPrompt = "New task title:"
               m.inputBuffer = ""
               return m, nil
           }
       } else if m.mode == AddTaskMode {
           switch msg.String() {
           case "enter":
               // Create task in selected column
               currentCol := m.columns[m.selectedColumn]
               task, err := database.CreateTask(m.db, m.inputBuffer, "", currentCol.ID, len(m.tasks[currentCol.ID]))
               if err == nil {
                   m.tasks[currentCol.ID] = append(m.tasks[currentCol.ID], task)
               }
               m.mode = NormalMode
               m.inputBuffer = ""
               return m, nil
           
           case "esc":
               m.mode = NormalMode
               m.inputBuffer = ""
               return m, nil
           
           default:
               // Append to input buffer
               m.inputBuffer += msg.String()
           }
       }
   ```

3. **Add task editing** (`internal/tui/update.go`)
   ```go
   case "e":
       if m.mode == NormalMode && len(m.getCurrentTasks()) > 0 {
           m.mode = EditTaskMode
           task := m.getCurrentTask()
           m.inputBuffer = task.Title
           m.inputPrompt = "Edit task title:"
       }
   
   // In EditTaskMode handling:
   case "enter":
       task := m.getCurrentTask()
       database.UpdateTaskTitle(m.db, task.ID, m.inputBuffer)
       task.Title = m.inputBuffer
       m.mode = NormalMode
       m.inputBuffer = ""
   ```

4. **Add task deletion** (`internal/tui/update.go`)
   ```go
   case "d":
       if m.mode == NormalMode && len(m.getCurrentTasks()) > 0 {
           m.mode = DeleteConfirmMode
       }
   
   // In DeleteConfirmMode:
   case "y":
       task := m.getCurrentTask()
       database.DeleteTask(m.db, task.ID)
       // Remove from local state
       m.removeCurrentTask()
       m.mode = NormalMode
   
   case "n", "esc":
       m.mode = NormalMode
   ```

5. **Helper methods** (`internal/tui/model.go`)
   ```go
   func (m Model) getCurrentTask() *models.Task {
       col := m.columns[m.selectedColumn]
       tasks := m.tasks[col.ID]
       if len(tasks) == 0 {
           return nil
       }
       return tasks[m.selectedTask]
   }
   
   func (m *Model) removeCurrentTask() {
       col := m.columns[m.selectedColumn]
       tasks := m.tasks[col.ID]
       m.tasks[col.ID] = append(tasks[:m.selectedTask], tasks[m.selectedTask+1:]...)
       if m.selectedTask >= len(m.tasks[col.ID]) && m.selectedTask > 0 {
           m.selectedTask--
       }
   }
   ```

6. **Update View for input mode** (`internal/tui/view.go`)
   ```go
   func (m Model) View() string {
       // ... existing board rendering
       
       footer := "[a]dd  [e]dit  [d]elete  [hjkl] navigate  [q] quit"
       
       if m.mode == AddTaskMode || m.mode == EditTaskMode {
           inputBox := lipgloss.NewStyle().
               Border(lipgloss.RoundedBorder()).
               Padding(1).
               Width(50).
               Render(fmt.Sprintf("%s\n> %s_", m.inputPrompt, m.inputBuffer))
           
           return lipgloss.Place(
               m.width, m.height,
               lipgloss.Center, lipgloss.Center,
               inputBox,
           )
       }
       
       if m.mode == DeleteConfirmMode {
           task := m.getCurrentTask()
           confirmBox := lipgloss.NewStyle().
               Border(lipgloss.RoundedBorder()).
               Padding(1).
               Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))
           
           return lipgloss.Place(
               m.width, m.height,
               lipgloss.Center, lipgloss.Center,
               confirmBox,
           )
       }
       
       return lipgloss.JoinVertical(lipgloss.Left, header, "", board, "", footer)
   }
   ```

### Testing Phase 5

1. Press 'a', type "Test task", press Enter → Task appears in selected column
2. Select a task, press 'e', modify title, press Enter → Task updates
3. Select a task, press 'd', press 'y' → Task is deleted
4. Press 'd', press 'n' → Deletion is cancelled
5. Press Esc during input → Cancels operation

**Definition of Done**:
- Can add new tasks with 'a' key
- Can edit existing task titles with 'e' key
- Can delete tasks with 'd' key (requires confirmation)
- All operations update the database immediately
- UI refreshes to show changes
- Input can be cancelled with Esc
- Input dialogs are centered and styled

---

## Phase 6: Column Viewport Scrolling

**Goal**: Enable horizontal scrolling when there are more columns than fit on screen.

### Objectives
- Calculate how many columns fit in viewport
- Implement viewport offset tracking
- Add scroll left/right with [ and ] keys
- Show indicators when more columns exist off-screen

### Files to Create/Modify

```
paso/
└── internal/
    └── tui/
        ├── model.go          # Add: viewportOffset, viewportSize
        ├── update.go         # Add: scroll handlers
        └── view.go           # Modify: render only visible columns
```

### Implementation Steps

1. **Update Model** (`internal/tui/model.go`)
   ```go
   type Model struct {
       // ... existing fields
       viewportOffset int  // Index of leftmost visible column
       viewportSize   int  // Number of columns that fit on screen
   }
   
   func (m *Model) calculateViewportSize() {
       columnWidth := 34  // 30 width + 2 margin + 2 border
       m.viewportSize = max(1, m.width / columnWidth)
   }
   ```

2. **Add scroll handlers** (`internal/tui/update.go`)
   ```go
   case "]":
       // Scroll right
       if m.viewportOffset + m.viewportSize < len(m.columns) {
           m.viewportOffset++
           // Adjust selectedColumn if it's now off-screen
           if m.selectedColumn < m.viewportOffset {
               m.selectedColumn = m.viewportOffset
           }
       }
   
   case "[":
       // Scroll left
       if m.viewportOffset > 0 {
           m.viewportOffset--
           // Adjust selectedColumn if it's now off-screen
           if m.selectedColumn >= m.viewportOffset + m.viewportSize {
               m.selectedColumn = m.viewportOffset + m.viewportSize - 1
           }
       }
   ```

3. **Auto-scroll to keep selection visible** (`internal/tui/update.go`)
   ```go
   case "h", "left":
       if m.selectedColumn > 0 {
           m.selectedColumn--
           // Auto-scroll viewport if needed
           if m.selectedColumn < m.viewportOffset {
               m.viewportOffset = m.selectedColumn
           }
       }
   
   case "l", "right":
       if m.selectedColumn < len(m.columns)-1 {
           m.selectedColumn++
           // Auto-scroll viewport if needed
           if m.selectedColumn >= m.viewportOffset + m.viewportSize {
               m.viewportOffset = m.selectedColumn - m.viewportSize + 1
           }
       }
   ```

4. **Update View to render viewport** (`internal/tui/view.go`)
   ```go
   func (m Model) View() string {
       // ... existing setup
       
       // Calculate visible columns
       visibleColumns := m.columns[m.viewportOffset : min(m.viewportOffset+m.viewportSize, len(m.columns))]
       
       var columnViews []string
       for i, col := range visibleColumns {
           globalIndex := m.viewportOffset + i
           // ... render with correct selection state
       }
       
       // Add scroll indicators
       leftArrow := " "
       rightArrow := " "
       if m.viewportOffset > 0 {
           leftArrow = "◀"
       }
       if m.viewportOffset + m.viewportSize < len(m.columns) {
           rightArrow = "▶"
       }
       
       board := lipgloss.JoinHorizontal(lipgloss.Top, leftArrow, columnsView, rightArrow)
       
       footer := "[hjkl] navigate  [[ ]] scroll  [a]dd  [e]dit  [d]elete  [q] quit"
       // ... rest of view
   }
   ```

5. **Update on window resize** (`internal/tui/update.go`)
   ```go
   case tea.WindowSizeMsg:
       m.width = msg.Width
       m.height = msg.Height
       m.calculateViewportSize()
       // Ensure viewport offset is still valid
       if m.viewportOffset + m.viewportSize > len(m.columns) {
           m.viewportOffset = max(0, len(m.columns) - m.viewportSize)
       }
   ```

### Testing Phase 6

1. Add more than 3 columns to database (e.g., "Backlog", "Review", "Archive")
2. Run `paso` with narrow terminal width
3. Verify:
   - Only visible columns render
   - `]` scrolls right to show hidden columns
   - `[` scrolls left
   - Arrows (◀ ▶) indicate more columns exist
   - Navigation with h/l auto-scrolls to keep selection visible
4. Resize terminal and verify viewport adjusts

**Definition of Done**:
- Viewport correctly calculates how many columns fit
- `[` and `]` keys scroll the viewport horizontally
- Arrow indicators show when columns exist off-screen
- Selected column stays visible (auto-scrolls if needed)
- Viewport adjusts correctly on terminal resize
- Works correctly with 1, 3, 5+ columns

---

## Phase 7: Task Movement Between Columns

**Goal**: Enable moving tasks between columns with keyboard shortcuts.

### Objectives
- Implement task movement to adjacent columns
- Update task position in database
- Add visual feedback during movement
- Handle edge cases (moving from first/last column)

### Files to Create/Modify

```
paso/
└── internal/
    └── tui/
        ├── update.go         # Add: task movement handlers
        └── database/
            └── repository.go  # Add: MoveTask function
```

### Implementation Steps

1. **Add MoveTask to repository** (`internal/database/repository.go`)
   ```go
   func MoveTask(db *sql.DB, taskID, newColumnID, newPosition int) error {
       tx, err := db.Begin()
       if err != nil {
           return err
       }
       defer tx.Rollback()
       
       // Update task's column
       _, err = tx.Exec(
           "UPDATE tasks SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
           newColumnID, newPosition, taskID,
       )
       if err != nil {
           return err
       }
       
       return tx.Commit()
   }
   ```

2. **Add movement handlers** (`internal/tui/update.go`)
   ```go
   case ">", "L":  // Shift+L or >
       if m.mode == NormalMode && len(m.getCurrentTasks()) > 0 {
           m.moveTaskRight()
       }
   
   case "<", "H":  // Shift+H or <
       if m.mode == NormalMode && len(m.getCurrentTasks()) > 0 {
           m.moveTaskLeft()
       }
   ```

3. **Implement movement logic** (`internal/tui/model.go`)
   ```go
   func (m *Model) moveTaskRight() {
       // Can't move right from last column
       if m.selectedColumn >= len(m.columns)-1 {
           return
       }
       
       task := m.getCurrentTask()
       if task == nil {
           return
       }
       
       // Remove from current column
       currentCol := m.columns[m.selectedColumn]
       m.tasks[currentCol.ID] = append(
           m.tasks[currentCol.ID][:m.selectedTask],
           m.tasks[currentCol.ID][m.selectedTask+1:]...,
       )
       
       // Add to next column
       nextCol := m.columns[m.selectedColumn+1]
       newPosition := len(m.tasks[nextCol.ID])
       task.ColumnID = nextCol.ID
       task.Position = newPosition
       m.tasks[nextCol.ID] = append(m.tasks[nextCol.ID], task)
       
       // Update database
       database.MoveTask(m.db, task.ID, nextCol.ID, newPosition)
       
       // Move selection to follow task
       m.selectedColumn++
       m.selectedTask = newPosition
       
       // Ensure new position is visible
       if m.selectedColumn >= m.viewportOffset + m.viewportSize {
           m.viewportOffset++
       }
   }
   
   func (m *Model) moveTaskLeft() {
       // Can't move left from first column
       if m.selectedColumn <= 0 {
           return
       }
       
       task := m.getCurrentTask()
       if task == nil {
           return
       }
       
       // Remove from current column
       currentCol := m.columns[m.selectedColumn]
       m.tasks[currentCol.ID] = append(
           m.tasks[currentCol.ID][:m.selectedTask],
           m.tasks[currentCol.ID][m.selectedTask+1:]...,
       )
       
       // Add to previous column
       prevCol := m.columns[m.selectedColumn-1]
       newPosition := len(m.tasks[prevCol.ID])
       task.ColumnID = prevCol.ID
       task.Position = newPosition
       m.tasks[prevCol.ID] = append(m.tasks[prevCol.ID], task)
       
       // Update database
       database.MoveTask(m.db, task.ID, prevCol.ID, newPosition)
       
       // Move selection to follow task
       m.selectedColumn--
       m.selectedTask = newPosition
       
       // Ensure new position is visible
       if m.selectedColumn < m.viewportOffset {
           m.viewportOffset--
       }
   }
   ```

4. **Update help text** (`internal/tui/view.go`)
   ```go
   footer := "[hjkl] navigate  [<>] move task  [a]dd  [e]dit  [d]elete  [q] quit"
   ```

### Testing Phase 7

1. Create a task in "Todo" column
2. Press `>` to move it to "In Progress" → Task moves and selection follows
3. Press `>` again to move to "Done"
4. Press `<` to move back to "In Progress"
5. Try moving from first column with `<` → Should do nothing
6. Try moving from last column with `>` → Should do nothing
7. Restart app and verify task is in correct column (database persistence)

**Definition of Done**:
- `>` moves task to the right column
- `<` moves task to the left column
- Selection follows the moved task
- Viewport scrolls if needed to keep moved task visible
- Can't move beyond first/last column (no errors)
- Database is updated immediately
- Task positions are correct after movement

---

## Phase 8: Smooth Animations with Harmonica

**Goal**: Add spring-based animations to make the UI feel polished and responsive.

### Objectives
- Install Harmonica
- Animate viewport scrolling
- Animate task movement between columns
- Add subtle selection animations

### Files to Create/Modify

```
paso/
└── internal/
    └── tui/
        ├── model.go          # Add: animation springs
        ├── update.go         # Add: frame tick handling
        └── view.go           # Use animated values for rendering
```

### Implementation Steps

1. **Install Harmonica**
   ```bash
   go get github.com/charmbracelet/harmonica
   ```

2. **Add animation state** (`internal/tui/model.go`)
   ```go
   import "github.com/charmbracelet/harmonica"
   
   type Model struct {
       // ... existing fields
       viewportOffsetSpring harmonica.Spring
       columnOffsetX        float64  // Animated X position for rendering
   }
   
   func InitialModel(db *sql.DB) Model {
       m := Model{
           db: db,
           // ... load data
       }
       
       // Initialize spring (FPS, stiffness, damping)
       m.viewportOffsetSpring = harmonica.NewSpring(harmonica.FPS(60), 10.0, 1.0)
       
       return m
   }
   ```

3. **Update on scroll** (`internal/tui/update.go`)
   ```go
   case "]":
       if m.viewportOffset + m.viewportSize < len(m.columns) {
           m.viewportOffset++
           // Set spring target to new offset
           m.viewportOffsetSpring.SetTarget(float64(m.viewportOffset))
           return m, m.viewportOffsetSpring.Tick()
       }
   
   case "[":
       if m.viewportOffset > 0 {
           m.viewportOffset--
           m.viewportOffsetSpring.SetTarget(float64(m.viewportOffset))
           return m, m.viewportOffsetSpring.Tick()
       }
   ```

4. **Handle animation frames** (`internal/tui/update.go`)
   ```go
   case harmonica.FrameMsg:
       var cmd tea.Cmd
       m.viewportOffsetSpring, cmd = m.viewportOffsetSpring.Update(msg)
       m.columnOffsetX = m.viewportOffsetSpring.Value()
       
       // Continue ticking if still moving
       if m.viewportOffsetSpring.Moving() {
           return m, tea.Batch(cmd, m.viewportOffsetSpring.Tick())
       }
       return m, cmd
   ```

5. **Use animated offset in rendering** (`internal/tui/view.go`)
   ```go
   func (m Model) View() string {
       // Calculate visible columns based on animated offset
       // This creates smooth sliding effect
       
       columnWidth := 34
       startX := int(-m.columnOffsetX * float64(columnWidth))
       
       // Render columns with calculated offset
       // (Implementation depends on your terminal rendering approach)
       // For simple version, can just use the integer viewport offset
       // and let the spring smooth the transitions
       
       // ... rest of rendering
   }
   ```

6. **Optional: Animate task movement** (`internal/tui/model.go`)
   ```go
   type MovingTask struct {
       task      *models.Task
       startCol  int
       endCol    int
       progress  harmonica.Spring
   }
   
   // Add to Model
   movingTask *MovingTask
   
   // When moving task:
   m.movingTask = &MovingTask{
       task:     task,
       startCol: m.selectedColumn,
       endCol:   m.selectedColumn + 1,
   }
   m.movingTask.progress = harmonica.NewSpring(harmonica.FPS(60), 15.0, 0.8)
   m.movingTask.progress.SetTarget(1.0)
   
   // Render moving task between columns with interpolated position
   ```

7. **Spring tuning options** (`internal/tui/model.go`)
   ```go
   // Fast and snappy
   harmonica.NewSpring(harmonica.FPS(60), 12.0, 1.0)
   
   // Bouncy and playful
   harmonica.NewSpring(harmonica.FPS(60), 8.0, 0.6)
   
   // Smooth and professional
   harmonica.NewSpring(harmonica.FPS(60), 10.0, 0.9)
   ```

### Testing Phase 8

1. Add 5+ columns to test scrolling
2. Press `]` → Columns should smoothly slide left (not snap)
3. Press `[` → Columns should smoothly slide right
4. Move a task with `>` → Should see smooth transition (if implemented)
5. Animations should feel natural, not sluggish or too fast

**Definition of Done**:
- Viewport scrolling animates smoothly with spring physics
- No jarring jumps when scrolling columns
- Animations run at 60 FPS without performance issues
- Can still navigate quickly (animations don't block input)
- Spring feels natural (good balance of speed and bounce)
- Optional: Task movement animates between columns

---

## Phase 9: Data Persistence & Reloading

**Goal**: Ensure all state is properly saved and reloaded across sessions.

### Objectives
- Auto-save on every change
- Reload state on startup
- Handle column reordering persistence
- Add task metadata (created_at, updated_at)

### Files to Create/Modify

```
paso/
└── internal/
    └── database/
        ├── repository.go     # Add: UpdateTaskPosition, etc.
        └── migrations.go     # Ensure proper indexes
```

### Implementation Steps

1. **Add position update function** (`internal/database/repository.go`)
   ```go
   func UpdateTaskPosition(db *sql.DB, taskID, columnID, position int) error {
       _, err := db.Exec(
           `UPDATE tasks 
            SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP 
            WHERE id = ?`,
           columnID, position, taskID,
       )
       return err
   }
   
   func ReorderTasksInColumn(db *sql.DB, columnID int) error {
       // Fix position values to be sequential (0, 1, 2, ...)
       rows, err := db.Query(
           `SELECT id FROM tasks WHERE column_id = ? ORDER BY position`,
           columnID,
       )
       if err != nil {
           return err
       }
       defer rows.Close()
       
       var taskIDs []int
       for rows.Next() {
           var id int
           rows.Scan(&id)
           taskIDs = append(taskIDs, id)
       }
       
       for i, id := range taskIDs {
           db.Exec("UPDATE tasks SET position = ? WHERE id = ?", i, id)
       }
       
       return nil
   }
   ```

2. **Ensure proper loading order** (`internal/database/repository.go`)
   ```go
   func GetAllTasks(db *sql.DB) (map[int][]*models.Task, error) {
       rows, err := db.Query(
           `SELECT id, title, description, column_id, position, created_at, updated_at
            FROM tasks
            ORDER BY column_id, position`,
       )
       if err != nil {
           return nil, err
       }
       defer rows.Close()
       
       tasks := make(map[int][]*models.Task)
       for rows.Next() {
           task := &models.Task{}
           err := rows.Scan(
               &task.ID, &task.Title, &task.Description,
               &task.ColumnID, &task.Position,
               &task.CreatedAt, &task.UpdatedAt,
           )
           if err != nil {
               continue
           }
           tasks[task.ColumnID] = append(tasks[task.ColumnID], task)
       }
       
       return tasks, nil
   }
   ```

3. **Add indexes for performance** (`internal/database/migrations.go`)
   ```go
   CREATE INDEX IF NOT EXISTS idx_tasks_column 
   ON tasks(column_id, position);
   
   CREATE INDEX IF NOT EXISTS idx_tasks_updated 
   ON tasks(updated_at DESC);
   ```

4. **Verify saves after every operation** (`internal/tui/update.go`)
   ```go
   // After task creation
   task, err := database.CreateTask(...)
   if err != nil {
       // Handle error (could show error message to user)
       return m, nil
   }
   
   // After task movement
   err := database.MoveTask(...)
   if err != nil {
       // Handle error
       return m, nil
   }
   
   // After task edit
   err := database.UpdateTaskTitle(...)
   if err != nil {
       // Handle error
       return m, nil
   }
   ```

5. **Add graceful error handling** (`internal/tui/model.go`)
   ```go
   type Model struct {
       // ... existing fields
       errorMessage string
       errorTimeout int  // Frames to show error
   }
   
   func (m *Model) setError(err error) {
       if err != nil {
           m.errorMessage = err.Error()
           m.errorTimeout = 180  // Show for 3 seconds at 60fps
       }
   }
   
   // In Update, decrement errorTimeout each frame
   // In View, show error if errorTimeout > 0
   ```

6. **Test data integrity** (create test script)
   ```go
   // test_persistence.go
   func main() {
       // Create tasks
       // Move tasks around
       // Close app
       // Reopen and verify state matches
   }
   ```

### Testing Phase 9

1. Create several tasks across different columns
2. Move tasks around
3. Edit task titles
4. Close `paso` with Ctrl+C
5. Reopen `paso` → All tasks should be exactly where you left them
6. Check database file directly with `sqlite3 ~/.paso/tasks.db`
7. Verify timestamps are updating correctly

**Definition of Done**:
- All task operations immediately save to database
- Restarting the app loads exact previous state
- Task positions are maintained correctly
- Timestamps (created_at, updated_at) are accurate
- Database queries are efficient (proper indexes)
- No data loss on unexpected exit (database is transactional)

---

## Phase 10: Polish & User Experience

**Goal**: Add final touches to make Paso production-ready and delightful to use.

### Objectives
- Add color themes
- Show task counts in column headers
- Add timestamps/metadata display
- Implement better error messages
- Add confirmation for destructive actions
- Create proper help screen

### Files to Create/Modify

```
paso/
└── internal/
    └── tui/
        ├── styles.go         # Add: color themes, more styles
        ├── help.go           # New: Help screen
        ├── view.go           # Add: metadata display
        └── components.go     # Polish: better formatting
```

### Implementation Steps

1. **Add color theme** (`internal/tui/styles.go`)
   ```go
   var (
       // Colors
       PrimaryColor   = lipgloss.Color("170")  // Purple
       SecondaryColor = lipgloss.Color("62")   // Blue
       SuccessColor   = lipgloss.Color("42")   // Green
       DangerColor    = lipgloss.Color("196")  // Red
       MutedColor     = lipgloss.Color("240")  // Gray
       
       // Column styles by type
       TodoColumnStyle = ColumnStyle.Copy().
           BorderForeground(lipgloss.Color("33"))  // Blue
       
       ProgressColumnStyle = ColumnStyle.Copy().
           BorderForeground(lipgloss.Color("214"))  // Orange
       
       DoneColumnStyle = ColumnStyle.Copy().
           BorderForeground(lipgloss.Color("42"))  // Green
   )
   ```

2. **Enhance column headers** (`internal/tui/components.go`)
   ```go
   func RenderColumn(column *models.Column, tasks []*models.Task, selected bool, selectedTaskIdx int) string {
       // Column icon based on name
       icon := "📋"
       if strings.Contains(strings.ToLower(column.Name), "progress") {
           icon = "🔨"
       } else if strings.Contains(strings.ToLower(column.Name), "done") {
           icon = "✅"
       }
       
       // Header with task count
       header := fmt.Sprintf("%s %s (%d)", icon, column.Name, len(tasks))
       headerStyle := TitleStyle
       if selected {
           headerStyle = headerStyle.Foreground(PrimaryColor)
       }
       
       content := headerStyle.Render(header) + "\n\n"
       
       // Render tasks or empty state
       if len(tasks) == 0 {
           emptyStyle := lipgloss.NewStyle().
               Foreground(MutedColor).
               Italic(true)
           content += emptyStyle.Render("No tasks")
       } else {
           // ... render tasks
       }
       
       return ColumnStyle.Render(content)
   }
   ```

3. **Add task metadata display** (`internal/tui/components.go`)
   ```go
   func RenderTask(task *models.Task, selected bool) string {
       title := fmt.Sprintf("▸ %s", task.Title)
       id := fmt.Sprintf("  PASO-%d", task.ID)
       
       // Show relative time if recently updated
       timeSince := time.Since(task.UpdatedAt)
       timeStr := ""
       if timeSince < 24*time.Hour {
           if timeSince < time.Hour {
               timeStr = fmt.Sprintf("  %dm ago", int(timeSince.Minutes()))
           } else {
               timeStr = fmt.Sprintf("  %dh ago", int(timeSince.Hours()))
           }
       }
       
       content := title + "\n" + id
       if timeStr != "" {
           timeStyle := lipgloss.NewStyle().Foreground(MutedColor)
           content += "\n" + timeStyle.Render(timeStr)
       }
       
       style := TaskStyle
       if selected {
           style = style.
               Foreground(PrimaryColor).
               Bold(true).
               BorderLeft(true).
               BorderStyle(lipgloss.ThickBorder()).
               BorderForeground(PrimaryColor)
       }
       
       return style.Render(content)
   }
   ```

4. **Create help screen** (`internal/tui/help.go`)
   ```go
   func RenderHelp() string {
       helpText := `
   PASO - Your Personal Kanban Board
   
   Navigation:
     h/←    Move to column on the left
     l/→    Move to column on the right
     j/↓    Move to task below
     k/↑    Move to task above
     [      Scroll viewport left
     ]      Scroll viewport right
   
   Task Operations:
     a      Add new task in current column
     e      Edit selected task
     d      Delete selected task (with confirmation)
     >      Move task to right column
     <      Move task to left column
   
   Other:
     ?      Toggle this help screen
     q      Quit application
     
   Pro Tips:
     - Tasks auto-save to ~/.paso/tasks.db
     - Use > and < for quick workflow progression
     - Press Esc to cancel any input dialog
   `
       
       style := lipgloss.NewStyle().
           Border(lipgloss.RoundedBorder()).
           BorderForeground(PrimaryColor).
           Padding(2).
           Width(60)
       
       return lipgloss.Place(
           100, 30,
           lipgloss.Center, lipgloss.Center,
           style.Render(helpText),
       )
   }
   ```

5. **Add help mode** (`internal/tui/model.go` and `update.go`)
   ```go
   const (
       NormalMode Mode = iota
       AddTaskMode
       EditTaskMode
       DeleteConfirmMode
       HelpMode  // Add this
   )
   
   // In Update:
   case "?":
       if m.mode == NormalMode {
           m.mode = HelpMode
       } else if m.mode == HelpMode {
           m.mode = NormalMode
       }
   
   // In View:
   if m.mode == HelpMode {
       return RenderHelp()
   }
   ```

6. **Improve error display** (`internal/tui/view.go`)
   ```go
   func (m Model) View() string {
       // ... existing rendering
       
       // Error banner at top if present
       if m.errorTimeout > 0 {
           errorStyle := lipgloss.NewStyle().
               Background(DangerColor).
               Foreground(lipgloss.Color("15")).
               Padding(0, 2).
               Width(m.width)
           
           errorBanner := errorStyle.Render("⚠ " + m.errorMessage)
           return lipgloss.JoinVertical(lipgloss.Left, errorBanner, mainView)
       }
       
       return mainView
   }
   ```

7. **Add status bar** (`internal/tui/view.go`)
   ```go
   func (m Model) renderStatusBar() string {
       leftStatus := fmt.Sprintf(" %d columns | %d total tasks", 
           len(m.columns), m.totalTaskCount())
       
       rightStatus := "Press ? for help "
       
       statusStyle := lipgloss.NewStyle().
           Background(lipgloss.Color("236")).
           Foreground(lipgloss.Color("250")).
           Width(m.width)
       
       gap := m.width - lipgloss.Width(leftStatus) - lipgloss.Width(rightStatus)
       return statusStyle.Render(leftStatus + strings.Repeat(" ", gap) + rightStatus)
   }
   ```

8. **Final touches** (`internal/tui/components.go`)
   ```go
   // Add subtle hover effects
   // Add task priority indicators (if you add priority field)
   // Add due date warnings (if you add due dates)
   // Add tags/labels display (if you add tags)
   ```

### Testing Phase 10

1. Open `paso` and appreciate the polished UI
2. Press `?` to view help screen
3. Navigate around - colors should be clear and consistent
4. Observe task counts in column headers
5. Check that recently updated tasks show relative timestamps
6. Verify status bar shows correct information
7. Test error display by triggering a database error (if possible)

**Definition of Done**:
- Color scheme is consistent and visually pleasing
- Column headers show task counts and icons
- Tasks display metadata (ID, relative time)
- Help screen is accessible with `?` key
- Error messages display clearly when things go wrong
- Status bar shows useful information
- Overall UI feels polished and professional
- No visual glitches or layout issues

---

## Final Deliverables

After completing all 10 phases, you should have:

1. **Working Application**
   - Zero-setup kanban board TUI
   - SQLite persistence
   - Full CRUD operations
   - Smooth animations
   - Column viewport scrolling

2. **Code Structure**
   ```
   paso/
   ├── main.go
   ├── go.mod
   ├── go.sum
   ├── internal/
   │   ├── database/
   │   │   ├── db.go
   │   │   ├── migrations.go
   │   │   └── repository.go
   │   ├── models/
   │   │   ├── column.go
   │   │   └── task.go
   │   └── tui/
   │       ├── model.go
   │       ├── update.go
   │       ├── view.go
   │       ├── components.go
   │       ├── styles.go
   │       ├── help.go
   │       └── input.go
   └── README.md
   ```

3. **User Experience**
   - Run `paso` → immediate kanban board
   - Intuitive keyboard shortcuts
   - Smooth, delightful animations
   - No configuration needed
   - Data persists across sessions

4. **Technical Achievements**
   - Pure Go implementation
   - No external dependencies (no CGo)
   - Single binary distribution
   - Efficient SQLite operations
   - 60 FPS animations
   - Responsive to terminal resizing

## Next Steps (Future Enhancements)

After core implementation is complete, consider:

- **Column management**: Add/remove/reorder custom columns
- **Task details**: Add description field, due dates, tags
- **Search/filter**: Quick search across all tasks
- **Task archiving**: Move completed tasks to archive
- **Time tracking**: Built-in Pomodoro timer
- **Export**: Generate reports or export to Markdown
- **Sync**: Optional cloud sync or Git-based sync
- **Themes**: User-configurable color schemes
- **Plugins**: Neovim integration (future phase)

## Build & Distribution

```bash
# Build for current platform
go build -o paso main.go

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o paso-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o paso-macos-amd64
GOOS=darwin GOARCH=arm64 go build -o paso-macos-arm64
GOOS=windows GOARCH=amd64 go build -o paso-windows-amd64.exe

# Install globally
go install
```

## Summary

This implementation guide provides a complete, scaffolded path to building Paso from the ground up. Each phase is self-contained, testable, and builds logically on the previous phase. By following this guide, you'll create a professional-quality terminal kanban board that's fast, beautiful, and delightful to use.

The key principles throughout:
- **Incremental development**: Each phase adds value
- **Test as you go**: Verify each phase works before moving on
- **Clean architecture**: Separation of concerns (data, business logic, UI)
- **User-first design**: Focus on experience, not just features
- **Polish matters**: Animations and details make it special

Good luck building Paso! 🚀
