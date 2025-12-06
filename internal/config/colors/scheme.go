package colors

// ColorScheme defines all configurable color values
type ColorScheme struct {
	// Preset name (e.g., "default", "monochrome")
	Preset string `yaml:"preset"`

	// Primary accent color (used for selections, titles, highlights)
	Accent string `yaml:"accent"`

	// Background color
	Background string `yaml:"background"`

	// Semantic colors
	Create string `yaml:"create"` // Green - creation dialogs
	Edit   string `yaml:"edit"`   // Blue - edit dialogs
	Delete string `yaml:"delete"` // Red - delete confirmations

	// UI element colors
	ColumnBorder     string `yaml:"column_border"`
	ColumnBackground string `yaml:"column_background"`
	TaskBorder       string `yaml:"task_border"`
	TaskBackground   string `yaml:"task_background"`
	SelectedBorder   string `yaml:"selected_border"`
	SelectedBg       string `yaml:"selected_bg"`

	// Text colors
	Title  string `yaml:"title"`
	Subtle string `yaml:"subtle"` // Muted/placeholder text
	Normal string `yaml:"normal"`

	// Notification colors (foreground/background pairs)
	InfoFg    string `yaml:"info_fg"`
	InfoBg    string `yaml:"info_bg"`
	WarningFg string `yaml:"warning_fg"`
	WarningBg string `yaml:"warning_bg"`
	ErrorFg   string `yaml:"error_fg"`
	ErrorBg   string `yaml:"error_bg"`
}

// GetPreset returns a preset color scheme by name
func GetPreset(name string) *ColorScheme {
	switch name {
	case "monochrome":
		return Monochrome()
	case "wave":
		return Wave()
	case "dragon":
		return Dragon()
	case "lotus":
		return Lotus()
	case "default", "":
		return Default()
	default:
		return Default()
	}
}

// MergeFrom merges non-empty color values from another ColorScheme
// Non-empty values from 'other' will override this scheme's values
func (c *ColorScheme) MergeFrom(other ColorScheme) {
	if other.Preset != "" {
		c.Preset = other.Preset
	}
	if other.Background == "" {
		c.Background = other.Background
	}
	if other.Accent != "" {
		c.Accent = other.Accent
	}
	if other.Create != "" {
		c.Create = other.Create
	}
	if other.Edit != "" {
		c.Edit = other.Edit
	}
	if other.Delete != "" {
		c.Delete = other.Delete
	}
	if other.ColumnBorder != "" {
		c.ColumnBorder = other.ColumnBorder
	}
	if other.ColumnBackground != "" {
		c.ColumnBackground = other.ColumnBackground
	}
	if other.TaskBorder != "" {
		c.TaskBorder = other.TaskBorder
	}
	if other.TaskBackground != "" {
		c.TaskBackground = other.TaskBackground
	}
	if other.SelectedBorder != "" {
		c.SelectedBorder = other.SelectedBorder
	}
	if other.SelectedBg != "" {
		c.SelectedBg = other.SelectedBg
	}
	if other.Title != "" {
		c.Title = other.Title
	}
	if other.Subtle != "" {
		c.Subtle = other.Subtle
	}
	if other.Normal != "" {
		c.Normal = other.Normal
	}
	if other.InfoFg != "" {
		c.InfoFg = other.InfoFg
	}
	if other.InfoBg != "" {
		c.InfoBg = other.InfoBg
	}
	if other.WarningFg != "" {
		c.WarningFg = other.WarningFg
	}
	if other.WarningBg != "" {
		c.WarningBg = other.WarningBg
	}
	if other.ErrorFg != "" {
		c.ErrorFg = other.ErrorFg
	}
	if other.ErrorBg != "" {
		c.ErrorBg = other.ErrorBg
	}
}

// ApplyDefaults fills in missing color values using the preset as base
// If preset is specified, loads that preset first, then overrides with custom values
func (c *ColorScheme) ApplyDefaults() {
	// Get the base preset
	preset := GetPreset(c.Preset)

	// Override with custom values (only if not empty)
	if c.Accent == "" {
		c.Accent = preset.Accent
	}
	if c.Background == "" {
		c.Background = preset.Background
	}
	if c.Create == "" {
		c.Create = preset.Create
	}
	if c.Edit == "" {
		c.Edit = preset.Edit
	}
	if c.Delete == "" {
		c.Delete = preset.Delete
	}
	if c.ColumnBorder == "" {
		c.ColumnBorder = preset.ColumnBorder
	}
	if c.ColumnBackground != "" {
		c.ColumnBackground = preset.ColumnBackground
	}
	if c.TaskBorder == "" {
		c.TaskBorder = preset.TaskBorder
	}
	if c.TaskBackground == "" {
		c.TaskBackground = preset.TaskBackground
	}
	if c.SelectedBorder == "" {
		c.SelectedBorder = preset.SelectedBorder
	}
	if c.SelectedBg == "" {
		c.SelectedBg = preset.SelectedBg
	}
	if c.Title == "" {
		c.Title = preset.Title
	}
	if c.Subtle == "" {
		c.Subtle = preset.Subtle
	}
	if c.Normal == "" {
		c.Normal = preset.Normal
	}
	if c.InfoFg == "" {
		c.InfoFg = preset.InfoFg
	}
	if c.InfoBg == "" {
		c.InfoBg = preset.InfoBg
	}
	if c.WarningFg == "" {
		c.WarningFg = preset.WarningFg
	}
	if c.WarningBg == "" {
		c.WarningBg = preset.WarningBg
	}
	if c.ErrorFg == "" {
		c.ErrorFg = preset.ErrorFg
	}
	if c.ErrorBg == "" {
		c.ErrorBg = preset.ErrorBg
	}
}
