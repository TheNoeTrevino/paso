package config

// KeyMappings defines all configurable key bindings, organized by context.
type KeyMappings struct {
	Navigation NavigationKeys `yaml:"navigation"`
	Tasks      TaskKeys       `yaml:"tasks"`
	Kanban     KanbanKeys     `yaml:"kanban"`
	Projects   ProjectKeys    `yaml:"projects"`
	Forms      FormKeys       `yaml:"forms"`
	Pickers    PickerKeys     `yaml:"pickers"`
	General    GeneralKeys    `yaml:"general"`
}

// NavigationKeys defines keybindings for moving around the UI.
type NavigationKeys struct {
	MoveLeft            string `yaml:"move_left"`
	MoveRight           string `yaml:"move_right"`
	MoveUp              string `yaml:"move_up"`
	MoveDown            string `yaml:"move_down"`
	ScrollViewportLeft  string `yaml:"scroll_viewport_left"`
	ScrollViewportRight string `yaml:"scroll_viewport_right"`
	NextProject         string `yaml:"next_project"`
	PrevProject         string `yaml:"prev_project"`
}

// TaskKeys defines keybindings for task actions.
type TaskKeys struct {
	AddTask           string `yaml:"add_task"`
	EditTask          string `yaml:"edit_task"`
	DeleteTask        string `yaml:"delete_task"`
	ViewTask          string `yaml:"view_task"`
	EditLabels        string `yaml:"edit_labels"`
	EditPriority      string `yaml:"edit_priority"`
	EditAssignee      string `yaml:"edit_assignee"`
	EditEstimate      string `yaml:"edit_estimate"`
	EditDueDate       string `yaml:"edit_due_date"`
	EditParentTask    string `yaml:"edit_parent_task"`
	EditChildTask     string `yaml:"edit_child_task"`
	EditType          string `yaml:"edit_type"`
	MoveTaskToProject string `yaml:"move_task_to_project"`
}

// KanbanKeys defines keybindings specific to the kanban board view.
type KanbanKeys struct {
	CreateColumn  string `yaml:"create_column"`
	RenameColumn  string `yaml:"rename_column"`
	DeleteColumn  string `yaml:"delete_column"`
	MoveTaskLeft  string `yaml:"move_task_left"`
	MoveTaskRight string `yaml:"move_task_right"`
	MoveTaskUp    string `yaml:"move_task_up"`
	MoveTaskDown  string `yaml:"move_task_down"`
}

// ProjectKeys defines keybindings for project management.
type ProjectKeys struct {
	CreateProject string `yaml:"create_project"`
	EditProject   string `yaml:"edit_project"`
	DeleteProject string `yaml:"delete_project"`
}

// FormKeys defines keybindings used within form contexts.
// These use ctrl+ modifiers to avoid conflicting with text input.
type FormKeys struct {
	SaveForm         string `yaml:"save_form"`
	OpenCommentsView string `yaml:"open_comments_view"`
	RefreshGitData   string `yaml:"refresh_git_data"`
	EditLabels       string `yaml:"edit_labels"`
	EditParentTask   string `yaml:"edit_parent_task"`
	EditChildTask    string `yaml:"edit_child_task"`
	EditPriority     string `yaml:"edit_priority"`
	EditType         string `yaml:"edit_type"`
	EditAssignee     string `yaml:"edit_assignee"`
	EditEstimate     string `yaml:"edit_estimate"`
	EditDueDate      string `yaml:"edit_due_date"`
	ShowHelp         string `yaml:"show_help"`
}

// PickerKeys defines keybindings used within picker overlays.
type PickerKeys struct {
	DeleteLabel string `yaml:"delete_label"`
}

// GeneralKeys defines global keybindings.
type GeneralKeys struct {
	ShowHelp      string `yaml:"show_help"`
	ConnectRemote string `yaml:"connect_remote"`
	FilterBar     string `yaml:"filter_bar"`
	Quit          string `yaml:"quit"`
	Suspend       string `yaml:"suspend"`
	ToggleView    string `yaml:"toggle_view"`
	Search        string `yaml:"search"`
}

// DefaultKeyMappings returns the default key mappings.
func DefaultKeyMappings() KeyMappings {
	return KeyMappings{
		Navigation: NavigationKeys{
			MoveLeft:            "h",
			MoveRight:           "l",
			MoveUp:              "k",
			MoveDown:            "j",
			ScrollViewportLeft:  "[",
			ScrollViewportRight: "]",
			NextProject:         "}",
			PrevProject:         "{",
		},
		Tasks: TaskKeys{
			AddTask:           "a",
			EditTask:          "e",
			DeleteTask:        "d",
			ViewTask:          " ",
			EditLabels:        "ctrl+l",
			EditPriority:      "ctrl+r",
			EditAssignee:      "ctrl+a",
			EditEstimate:      "ctrl+e",
			EditDueDate:       "ctrl+d",
			EditParentTask:    "p",
			EditChildTask:     "c",
			EditType:          "ctrl+t",
			MoveTaskToProject: "s",
		},
		Kanban: KanbanKeys{
			CreateColumn:  "C",
			RenameColumn:  "R",
			DeleteColumn:  "X",
			MoveTaskLeft:  "H",
			MoveTaskRight: "L",
			MoveTaskUp:    "K",
			MoveTaskDown:  "J",
		},
		Projects: ProjectKeys{
			CreateProject: "P",
			EditProject:   "E",
			DeleteProject: "D",
		},
		Forms: FormKeys{
			SaveForm:         "ctrl+s",
			OpenCommentsView: "ctrl+n",
			RefreshGitData:   "f5",
			EditLabels:       "ctrl+l",
			EditParentTask:   "ctrl+p",
			EditChildTask:    "ctrl+c",
			EditPriority:     "ctrl+r",
			EditType:         "ctrl+t",
			EditAssignee:     "ctrl+a",
			EditEstimate:     "ctrl+e",
			EditDueDate:      "ctrl+d",
			ShowHelp:         "ctrl+/",
		},
		Pickers: PickerKeys{
			DeleteLabel: "ctrl+d",
		},
		General: GeneralKeys{
			ShowHelp:      "?",
			ConnectRemote: "ctrl+h",
			FilterBar:     "ctrl+f",
			Quit:          "q",
			Suspend:       "ctrl+z",
			ToggleView:    "v",
			Search:        "/",
		},
	}
}

func (n *NavigationKeys) applyDefaults() {
	defaults := DefaultKeyMappings().Navigation
	if n.MoveLeft == "" {
		n.MoveLeft = defaults.MoveLeft
	}
	if n.MoveRight == "" {
		n.MoveRight = defaults.MoveRight
	}
	if n.MoveUp == "" {
		n.MoveUp = defaults.MoveUp
	}
	if n.MoveDown == "" {
		n.MoveDown = defaults.MoveDown
	}
	if n.ScrollViewportLeft == "" {
		n.ScrollViewportLeft = defaults.ScrollViewportLeft
	}
	if n.ScrollViewportRight == "" {
		n.ScrollViewportRight = defaults.ScrollViewportRight
	}
	if n.NextProject == "" {
		n.NextProject = defaults.NextProject
	}
	if n.PrevProject == "" {
		n.PrevProject = defaults.PrevProject
	}
}

func (t *TaskKeys) applyDefaults() {
	defaults := DefaultKeyMappings().Tasks
	if t.AddTask == "" {
		t.AddTask = defaults.AddTask
	}
	if t.EditTask == "" {
		t.EditTask = defaults.EditTask
	}
	if t.DeleteTask == "" {
		t.DeleteTask = defaults.DeleteTask
	}
	if t.ViewTask == "" {
		t.ViewTask = defaults.ViewTask
	}
	if t.EditLabels == "" {
		t.EditLabels = defaults.EditLabels
	}
	if t.EditPriority == "" {
		t.EditPriority = defaults.EditPriority
	}
	if t.EditAssignee == "" {
		t.EditAssignee = defaults.EditAssignee
	}
	if t.EditEstimate == "" {
		t.EditEstimate = defaults.EditEstimate
	}
	if t.EditDueDate == "" {
		t.EditDueDate = defaults.EditDueDate
	}
	if t.EditParentTask == "" {
		t.EditParentTask = defaults.EditParentTask
	}
	if t.EditChildTask == "" {
		t.EditChildTask = defaults.EditChildTask
	}
	if t.EditType == "" {
		t.EditType = defaults.EditType
	}
	if t.MoveTaskToProject == "" {
		t.MoveTaskToProject = defaults.MoveTaskToProject
	}
}

func (k *KanbanKeys) applyDefaults() {
	defaults := DefaultKeyMappings().Kanban
	if k.CreateColumn == "" {
		k.CreateColumn = defaults.CreateColumn
	}
	if k.RenameColumn == "" {
		k.RenameColumn = defaults.RenameColumn
	}
	if k.DeleteColumn == "" {
		k.DeleteColumn = defaults.DeleteColumn
	}
	if k.MoveTaskLeft == "" {
		k.MoveTaskLeft = defaults.MoveTaskLeft
	}
	if k.MoveTaskRight == "" {
		k.MoveTaskRight = defaults.MoveTaskRight
	}
	if k.MoveTaskUp == "" {
		k.MoveTaskUp = defaults.MoveTaskUp
	}
	if k.MoveTaskDown == "" {
		k.MoveTaskDown = defaults.MoveTaskDown
	}
}

func (p *ProjectKeys) applyDefaults() {
	defaults := DefaultKeyMappings().Projects
	if p.CreateProject == "" {
		p.CreateProject = defaults.CreateProject
	}
	if p.EditProject == "" {
		p.EditProject = defaults.EditProject
	}
	if p.DeleteProject == "" {
		p.DeleteProject = defaults.DeleteProject
	}
}

func (f *FormKeys) applyDefaults() {
	defaults := DefaultKeyMappings().Forms
	if f.SaveForm == "" {
		f.SaveForm = defaults.SaveForm
	}
	if f.OpenCommentsView == "" {
		f.OpenCommentsView = defaults.OpenCommentsView
	}
	if f.RefreshGitData == "" {
		f.RefreshGitData = defaults.RefreshGitData
	}
	if f.EditLabels == "" {
		f.EditLabels = defaults.EditLabels
	}
	if f.EditParentTask == "" {
		f.EditParentTask = defaults.EditParentTask
	}
	if f.EditChildTask == "" {
		f.EditChildTask = defaults.EditChildTask
	}
	if f.EditPriority == "" {
		f.EditPriority = defaults.EditPriority
	}
	if f.EditType == "" {
		f.EditType = defaults.EditType
	}
	if f.EditAssignee == "" {
		f.EditAssignee = defaults.EditAssignee
	}
	if f.EditEstimate == "" {
		f.EditEstimate = defaults.EditEstimate
	}
	if f.EditDueDate == "" {
		f.EditDueDate = defaults.EditDueDate
	}
	if f.ShowHelp == "" {
		f.ShowHelp = defaults.ShowHelp
	}
}

func (p *PickerKeys) applyDefaults() {
	defaults := DefaultKeyMappings().Pickers
	if p.DeleteLabel == "" {
		p.DeleteLabel = defaults.DeleteLabel
	}
}

func (g *GeneralKeys) applyDefaults() {
	defaults := DefaultKeyMappings().General
	if g.ShowHelp == "" {
		g.ShowHelp = defaults.ShowHelp
	}
	if g.ConnectRemote == "" {
		g.ConnectRemote = defaults.ConnectRemote
	}
	if g.FilterBar == "" {
		g.FilterBar = defaults.FilterBar
	}
	if g.Quit == "" {
		g.Quit = defaults.Quit
	}
	if g.Suspend == "" {
		g.Suspend = defaults.Suspend
	}
	if g.ToggleView == "" {
		g.ToggleView = defaults.ToggleView
	}
	if g.Search == "" {
		g.Search = defaults.Search
	}
}

// applyDefaults fills in missing key mappings with defaults.
func (k *KeyMappings) applyDefaults() {
	k.Navigation.applyDefaults()
	k.Tasks.applyDefaults()
	k.Kanban.applyDefaults()
	k.Projects.applyDefaults()
	k.Forms.applyDefaults()
	k.Pickers.applyDefaults()
	k.General.applyDefaults()
}
