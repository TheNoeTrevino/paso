package colors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPreset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectedPreset string
	}{
		{
			name:           "default preset",
			input:          "default",
			expectedPreset: "default",
		},
		{
			name:           "empty string returns default",
			input:          "",
			expectedPreset: "default",
		},
		{
			name:           "wave preset",
			input:          "wave",
			expectedPreset: "wave",
		},
		{
			name:           "dragon preset",
			input:          "dragon",
			expectedPreset: "dragon",
		},
		{
			name:           "lotus preset",
			input:          "lotus",
			expectedPreset: "lotus",
		},
		{
			name:           "monochrome preset",
			input:          "monochrome",
			expectedPreset: "monochrome",
		},
		{
			name:           "unknown name falls back to default",
			input:          "nonexistent",
			expectedPreset: "default",
		},
		{
			name:           "random string falls back to default",
			input:          "foobar",
			expectedPreset: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			result := GetPreset(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedPreset, result.Preset)
		})
	}
}

func TestGetPreset_AllPresetsFullyPopulated(t *testing.T) {
	t.Parallel()

	presets := []string{"default", "wave", "dragon", "lotus", "monochrome"}

	for _, name := range presets {
		t.Run(name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			scheme := GetPreset(name)
			require.NotNil(t, scheme)

			assert.NotEmpty(t, scheme.Preset, "Preset")
			assert.NotEmpty(t, scheme.Accent, "Accent")
			assert.NotEmpty(t, scheme.Background, "Background")
			assert.NotEmpty(t, scheme.Create, "Create")
			assert.NotEmpty(t, scheme.Edit, "Edit")
			assert.NotEmpty(t, scheme.Delete, "Delete")
			assert.NotEmpty(t, scheme.Blocked, "Blocked")
			assert.NotEmpty(t, scheme.ColumnBorder, "ColumnBorder")
			assert.NotEmpty(t, scheme.ColumnBackground, "ColumnBackground")
			assert.NotEmpty(t, scheme.TaskBorder, "TaskBorder")
			assert.NotEmpty(t, scheme.TaskBackground, "TaskBackground")
			assert.NotEmpty(t, scheme.SelectedBorder, "SelectedBorder")
			assert.NotEmpty(t, scheme.SelectedBg, "SelectedBg")
			assert.NotEmpty(t, scheme.Title, "Title")
			assert.NotEmpty(t, scheme.Subtle, "Subtle")
			assert.NotEmpty(t, scheme.Normal, "Normal")
			assert.NotEmpty(t, scheme.InfoFg, "InfoFg")
			assert.NotEmpty(t, scheme.InfoBg, "InfoBg")
			assert.NotEmpty(t, scheme.WarningFg, "WarningFg")
			assert.NotEmpty(t, scheme.WarningBg, "WarningBg")
			assert.NotEmpty(t, scheme.ErrorFg, "ErrorFg")
			assert.NotEmpty(t, scheme.ErrorBg, "ErrorBg")
			assert.NotEmpty(t, scheme.StatusBarBg, "StatusBarBg")
			assert.NotEmpty(t, scheme.StatusBarText, "StatusBarText")
		})
	}
}

func TestMergeFrom_NonEmptyValuesOverride(t *testing.T) {
	t.Parallel()

	base := ColorScheme{
		Preset:           "default",
		Accent:           "#000000",
		Background:       "#111111",
		Create:           "#222222",
		Edit:             "#333333",
		Delete:           "#444444",
		Blocked:          "#555555",
		ColumnBorder:     "#666666",
		ColumnBackground: "#777777",
		TaskBorder:       "#888888",
		TaskBackground:   "#999999",
		SelectedBorder:   "#AAAAAA",
		SelectedBg:       "#BBBBBB",
		Title:            "#CCCCCC",
		Subtle:           "#DDDDDD",
		Normal:           "#EEEEEE",
		InfoFg:           "#110000",
		InfoBg:           "#220000",
		WarningFg:        "#330000",
		WarningBg:        "#440000",
		ErrorFg:          "#550000",
		ErrorBg:          "#660000",
		StatusBarBg:      "#770000",
		StatusBarText:    "#880000",
	}

	other := ColorScheme{
		Preset:           "wave",
		Accent:           "#FF0000",
		Background:       "#FF1111",
		Create:           "#FF2222",
		Edit:             "#FF3333",
		Delete:           "#FF4444",
		Blocked:          "#FF5555",
		ColumnBorder:     "#FF6666",
		ColumnBackground: "#FF7777",
		TaskBorder:       "#FF8888",
		TaskBackground:   "#FF9999",
		SelectedBorder:   "#FFAAAA",
		SelectedBg:       "#FFBBBB",
		Title:            "#FFCCCC",
		Subtle:           "#FFDDDD",
		Normal:           "#FFEEEE",
		InfoFg:           "#FF1100",
		InfoBg:           "#FF2200",
		WarningFg:        "#FF3300",
		WarningBg:        "#FF4400",
		ErrorFg:          "#FF5500",
		ErrorBg:          "#FF6600",
		StatusBarBg:      "#FF7700",
		StatusBarText:    "#FF8800",
	}

	base.MergeFrom(other)

	assert.Equal(t, "wave", base.Preset)
	assert.Equal(t, "#FF0000", base.Accent)
	assert.Equal(t, "#FF1111", base.Background)
	assert.Equal(t, "#FF2222", base.Create)
	assert.Equal(t, "#FF3333", base.Edit)
	assert.Equal(t, "#FF4444", base.Delete)
	assert.Equal(t, "#FF5555", base.Blocked)
	assert.Equal(t, "#FF6666", base.ColumnBorder)
	assert.Equal(t, "#FF7777", base.ColumnBackground)
	assert.Equal(t, "#FF8888", base.TaskBorder)
	assert.Equal(t, "#FF9999", base.TaskBackground)
	assert.Equal(t, "#FFAAAA", base.SelectedBorder)
	assert.Equal(t, "#FFBBBB", base.SelectedBg)
	assert.Equal(t, "#FFCCCC", base.Title)
	assert.Equal(t, "#FFDDDD", base.Subtle)
	assert.Equal(t, "#FFEEEE", base.Normal)
	assert.Equal(t, "#FF1100", base.InfoFg)
	assert.Equal(t, "#FF2200", base.InfoBg)
	assert.Equal(t, "#FF3300", base.WarningFg)
	assert.Equal(t, "#FF4400", base.WarningBg)
	assert.Equal(t, "#FF5500", base.ErrorFg)
	assert.Equal(t, "#FF6600", base.ErrorBg)
	assert.Equal(t, "#FF7700", base.StatusBarBg)
	assert.Equal(t, "#FF8800", base.StatusBarText)
}

func TestMergeFrom_EmptyValuesDoNotOverride(t *testing.T) {
	t.Parallel()

	base := ColorScheme{
		Preset:           "default",
		Accent:           "#000000",
		Background:       "#111111",
		Create:           "#222222",
		Edit:             "#333333",
		Delete:           "#444444",
		Blocked:          "#555555",
		ColumnBorder:     "#666666",
		ColumnBackground: "#777777",
		TaskBorder:       "#888888",
		TaskBackground:   "#999999",
		SelectedBorder:   "#AAAAAA",
		SelectedBg:       "#BBBBBB",
		Title:            "#CCCCCC",
		Subtle:           "#DDDDDD",
		Normal:           "#EEEEEE",
		InfoFg:           "#110000",
		InfoBg:           "#220000",
		WarningFg:        "#330000",
		WarningBg:        "#440000",
		ErrorFg:          "#550000",
		ErrorBg:          "#660000",
		StatusBarBg:      "#770000",
		StatusBarText:    "#880000",
	}

	original := base
	base.MergeFrom(ColorScheme{})

	assert.Equal(t, original.Preset, base.Preset)
	assert.Equal(t, original.Accent, base.Accent)
	assert.Equal(t, original.Background, base.Background)
	assert.Equal(t, original.Create, base.Create)
	assert.Equal(t, original.Edit, base.Edit)
	assert.Equal(t, original.Delete, base.Delete)
	assert.Equal(t, original.Blocked, base.Blocked)
	assert.Equal(t, original.ColumnBorder, base.ColumnBorder)
	assert.Equal(t, original.ColumnBackground, base.ColumnBackground)
	assert.Equal(t, original.TaskBorder, base.TaskBorder)
	assert.Equal(t, original.TaskBackground, base.TaskBackground)
	assert.Equal(t, original.SelectedBorder, base.SelectedBorder)
	assert.Equal(t, original.SelectedBg, base.SelectedBg)
	assert.Equal(t, original.Title, base.Title)
	assert.Equal(t, original.Subtle, base.Subtle)
	assert.Equal(t, original.Normal, base.Normal)
	assert.Equal(t, original.InfoFg, base.InfoFg)
	assert.Equal(t, original.InfoBg, base.InfoBg)
	assert.Equal(t, original.WarningFg, base.WarningFg)
	assert.Equal(t, original.WarningBg, base.WarningBg)
	assert.Equal(t, original.ErrorFg, base.ErrorFg)
	assert.Equal(t, original.ErrorBg, base.ErrorBg)
	assert.Equal(t, original.StatusBarBg, base.StatusBarBg)
	assert.Equal(t, original.StatusBarText, base.StatusBarText)
}

func TestMergeFrom_PartialOverride(t *testing.T) {
	t.Parallel()

	base := *Default()
	other := ColorScheme{
		Accent:     "#CUSTOM1",
		Background: "#CUSTOM2",
	}

	base.MergeFrom(other)

	assert.Equal(t, "#CUSTOM1", base.Accent, "Accent should be overridden")
	assert.Equal(t, "#CUSTOM2", base.Background, "Background should be overridden")
	assert.Equal(t, Default().Create, base.Create, "Create should remain from base")
	assert.Equal(t, Default().Edit, base.Edit, "Edit should remain from base")
}

func TestApplyDefaults_EmptySchemeGetsAllDefaults(t *testing.T) {
	t.Parallel()

	scheme := &ColorScheme{}
	scheme.ApplyDefaults()

	defaults := Default()

	assert.Equal(t, defaults.Accent, scheme.Accent)
	assert.Equal(t, defaults.Background, scheme.Background)
	assert.Equal(t, defaults.Create, scheme.Create)
	assert.Equal(t, defaults.Edit, scheme.Edit)
	assert.Equal(t, defaults.Delete, scheme.Delete)
	assert.Equal(t, defaults.Blocked, scheme.Blocked)
	assert.Equal(t, defaults.ColumnBorder, scheme.ColumnBorder)
	assert.Equal(t, defaults.ColumnBackground, scheme.ColumnBackground)
	assert.Equal(t, defaults.TaskBorder, scheme.TaskBorder)
	assert.Equal(t, defaults.TaskBackground, scheme.TaskBackground)
	assert.Equal(t, defaults.SelectedBorder, scheme.SelectedBorder)
	assert.Equal(t, defaults.SelectedBg, scheme.SelectedBg)
	assert.Equal(t, defaults.Title, scheme.Title)
	assert.Equal(t, defaults.Subtle, scheme.Subtle)
	assert.Equal(t, defaults.Normal, scheme.Normal)
	assert.Equal(t, defaults.InfoFg, scheme.InfoFg)
	assert.Equal(t, defaults.InfoBg, scheme.InfoBg)
	assert.Equal(t, defaults.WarningFg, scheme.WarningFg)
	assert.Equal(t, defaults.WarningBg, scheme.WarningBg)
	assert.Equal(t, defaults.ErrorFg, scheme.ErrorFg)
	assert.Equal(t, defaults.ErrorBg, scheme.ErrorBg)
	assert.Equal(t, defaults.StatusBarBg, scheme.StatusBarBg)
	assert.Equal(t, defaults.StatusBarText, scheme.StatusBarText)
}

func TestApplyDefaults_CustomValuesNotOverridden(t *testing.T) {
	t.Parallel()

	scheme := &ColorScheme{
		Accent:     "#CUSTOM_ACCENT",
		Background: "#CUSTOM_BG",
		Create:     "#CUSTOM_CREATE",
	}
	scheme.ApplyDefaults()

	assert.Equal(t, "#CUSTOM_ACCENT", scheme.Accent, "custom Accent should be preserved")
	assert.Equal(t, "#CUSTOM_BG", scheme.Background, "custom Background should be preserved")
	assert.Equal(t, "#CUSTOM_CREATE", scheme.Create, "custom Create should be preserved")

	defaults := Default()
	assert.Equal(t, defaults.Edit, scheme.Edit, "Edit should get default value")
	assert.Equal(t, defaults.Delete, scheme.Delete, "Delete should get default value")
}

func TestApplyDefaults_UsesNamedPreset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		preset string
	}{
		{name: "wave preset", preset: "wave"},
		{name: "dragon preset", preset: "dragon"},
		{name: "lotus preset", preset: "lotus"},
		{name: "monochrome preset", preset: "monochrome"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			scheme := &ColorScheme{Preset: tt.preset}
			scheme.ApplyDefaults()

			expected := GetPreset(tt.preset)
			assert.Equal(t, expected.Accent, scheme.Accent)
			assert.Equal(t, expected.Background, scheme.Background)
			assert.Equal(t, expected.ColumnBackground, scheme.ColumnBackground)
			assert.Equal(t, expected.Normal, scheme.Normal)
		})
	}
}

func TestApplyDefaults_CustomValueWithPreset(t *testing.T) {
	t.Parallel()

	scheme := &ColorScheme{
		Preset: "wave",
		Accent: "#MY_CUSTOM_ACCENT",
	}
	scheme.ApplyDefaults()

	wave := Wave()
	assert.Equal(t, "#MY_CUSTOM_ACCENT", scheme.Accent, "custom Accent should be preserved")
	assert.Equal(t, wave.Background, scheme.Background, "Background should come from wave preset")
	assert.Equal(t, wave.ColumnBackground, scheme.ColumnBackground, "ColumnBackground should come from wave preset")
}

func TestPresetConstructors_ReturnDistinctSchemes(t *testing.T) {
	t.Parallel()

	d := Default()
	w := Wave()
	dr := Dragon()
	l := Lotus()
	m := Monochrome()

	assert.NotEqual(t, d.Accent, w.Accent, "default and wave should have different accents")
	assert.NotEqual(t, d.Accent, dr.Accent, "default and dragon should have different accents")
	assert.NotEqual(t, d.Accent, l.Accent, "default and lotus should have different accents")
	assert.NotEqual(t, d.Accent, m.Accent, "default and monochrome should have different accents")
}
