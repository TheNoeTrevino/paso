package config

// KeyMappings defines all configurable key bindings
type KeyMappings struct {
	// Tasks
	AddTask       string `yaml:"add_task"`
	EditTask      string `yaml:"edit_task"`
	DeleteTask    string `yaml:"delete_task"`
	MoveTaskLeft  string `yaml:"move_task_left"`
	MoveTaskRight string `yaml:"move_task_right"`
	MoveTaskUp    string `yaml:"move_task_up"`
	MoveTaskDown  string `yaml:"move_task_down"`
	ViewTask      string `yaml:"view_task"`
	EditLabels    string `yaml:"edit_labels"`
	EditParentTask string `yaml:"edit_parent_task"`
	EditChildTask  string `yaml:"edit_child_task"`

	// Forms
	SaveForm string `yaml:"save_form"`

	// Columns
	CreateColumn string `yaml:"create_column"`
	RenameColumn string `yaml:"rename_column"`
	DeleteColumn string `yaml:"delete_column"`

	// Projects
	CreateProject string `yaml:"create_project"`

	// Navigation
	PrevColumn          string `yaml:"prev_column"`
	NextColumn          string `yaml:"next_column"`
	PrevTask            string `yaml:"prev_task"`
	NextTask            string `yaml:"next_task"`
	ScrollViewportLeft  string `yaml:"scroll_viewport_left"`
	ScrollViewportRight string `yaml:"scroll_viewport_right"`
	NextProject         string `yaml:"next_project"`
	PrevProject         string `yaml:"prev_project"`

	// Other
	ShowHelp string `yaml:"show_help"`
	Quit     string `yaml:"quit"`
}

// DefaultKeyMappings returns the default key mappings
func DefaultKeyMappings() KeyMappings {
	return KeyMappings{
		// Tasks
		AddTask:       "a",
		EditTask:      "e",
		DeleteTask:    "d",
		MoveTaskLeft:  "L",
		MoveTaskRight: "H",
		MoveTaskUp:    "K",
		MoveTaskDown:  "J",
		ViewTask:      " ",
		EditLabels:    "l",
		EditParentTask: "p",
		EditChildTask:  "c",
		SaveForm:       "ctrl+s",

		// Columns
		CreateColumn: "C",
		RenameColumn: "R",
		DeleteColumn: "X",

		// Projects
		CreateProject: "P",

		// Navigation
		PrevColumn:          "h",
		NextColumn:          "l",
		PrevTask:            "k",
		NextTask:            "j",
		ScrollViewportLeft:  "[",
		ScrollViewportRight: "]",
		NextProject:         "}",
		PrevProject:         "{",

		// Other
		ShowHelp: "?",
		Quit:     "q",
	}
}

// applyDefaults fills in missing key mappings with defaults
func (k *KeyMappings) applyDefaults() {
	defaults := DefaultKeyMappings()

	if k.AddTask == "" {
		k.AddTask = defaults.AddTask
	}
	if k.EditTask == "" {
		k.EditTask = defaults.EditTask
	}
	if k.DeleteTask == "" {
		k.DeleteTask = defaults.DeleteTask
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
	if k.ViewTask == "" {
		k.ViewTask = defaults.ViewTask
	}
	if k.EditLabels == "" {
		k.EditLabels = defaults.EditLabels
	}
	if k.EditParentTask == "" {
		k.EditParentTask = defaults.EditParentTask
	}
	if k.EditChildTask == "" {
		k.EditChildTask = defaults.EditChildTask
	}
	if k.SaveForm == "" {
		k.SaveForm = defaults.SaveForm
	}
	if k.CreateColumn == "" {
		k.CreateColumn = defaults.CreateColumn
	}
	if k.RenameColumn == "" {
		k.RenameColumn = defaults.RenameColumn
	}
	if k.DeleteColumn == "" {
		k.DeleteColumn = defaults.DeleteColumn
	}
	if k.CreateProject == "" {
		k.CreateProject = defaults.CreateProject
	}
	if k.PrevColumn == "" {
		k.PrevColumn = defaults.PrevColumn
	}
	if k.NextColumn == "" {
		k.NextColumn = defaults.NextColumn
	}
	if k.PrevTask == "" {
		k.PrevTask = defaults.PrevTask
	}
	if k.NextTask == "" {
		k.NextTask = defaults.NextTask
	}
	if k.ScrollViewportLeft == "" {
		k.ScrollViewportLeft = defaults.ScrollViewportLeft
	}
	if k.ScrollViewportRight == "" {
		k.ScrollViewportRight = defaults.ScrollViewportRight
	}
	if k.NextProject == "" {
		k.NextProject = defaults.NextProject
	}
	if k.PrevProject == "" {
		k.PrevProject = defaults.PrevProject
	}
	if k.ShowHelp == "" {
		k.ShowHelp = defaults.ShowHelp
	}
	if k.Quit == "" {
		k.Quit = defaults.Quit
	}
}
