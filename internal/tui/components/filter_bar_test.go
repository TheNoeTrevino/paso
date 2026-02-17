package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestBuildAllChipTexts tests the construction of all filter chip display texts.
func TestBuildAllChipTexts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		props    FilterBarProps
		expected map[state.FilterChip]string
	}{
		{
			name: "all filters empty",
			props: FilterBarProps{
				Filter:       state.NewFilterState(),
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: All",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "search query active",
			props: FilterBarProps{
				Filter: &state.FilterState{
					SearchQuery: "bug fix",
				},
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: bug fix",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "search input active shows cursor",
			props: FilterBarProps{
				Filter: &state.FilterState{
					SearchQuery:       "hel",
					SearchInputActive: true,
				},
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: hel▎",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "search input active empty query shows cursor only",
			props: FilterBarProps{
				Filter: &state.FilterState{
					SearchInputActive: true,
				},
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: ▎",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "single label",
			props: FilterBarProps{
				Filter:       state.NewFilterState(),
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{"bug"},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: All",
				state.FilterChipLabel:    " Label: bug",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "multiple labels",
			props: FilterBarProps{
				Filter:       state.NewFilterState(),
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{"bug", "frontend", "urgent"},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: All",
				state.FilterChipLabel:    " Label: bug, frontend, urgent",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "priority filter set",
			props: FilterBarProps{
				Filter:       state.NewFilterState(),
				PriorityName: "High",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: All",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: High",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "type filter set",
			props: FilterBarProps{
				Filter:       state.NewFilterState(),
				PriorityName: "",
				TypeName:     "Bug",
				AssigneeName: "",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: All",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: Bug",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "assignee filter set",
			props: FilterBarProps{
				Filter:       state.NewFilterState(),
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "John Doe",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: All",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: John Doe",
				state.FilterChipArchived: "󱝋 Archived: Off",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "archived filter on",
			props: FilterBarProps{
				Filter: &state.FilterState{
					ShowArchived: true,
				},
				PriorityName: "",
				TypeName:     "",
				AssigneeName: "",
				LabelNames:   []string{},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: All",
				state.FilterChipLabel:    " Label: All",
				state.FilterChipPriority: " Priority: All",
				state.FilterChipType:     "󰉺 Type: All",
				state.FilterChipAssignee: " Assignee: All",
				state.FilterChipArchived: "󱝍 Archived: On",
				state.FilterChipClearAll: " Clear All",
			},
		},
		{
			name: "all filters set",
			props: FilterBarProps{
				Filter: &state.FilterState{
					ShowArchived: true,
					SearchQuery:  "deploy",
				},
				PriorityName: "Critical",
				TypeName:     "Feature",
				AssigneeName: "Jane Smith",
				LabelNames:   []string{"backend", "database"},
			},
			expected: map[state.FilterChip]string{
				state.FilterChipSearch:   " Search: deploy",
				state.FilterChipLabel:    " Label: backend, database",
				state.FilterChipPriority: " Priority: Critical",
				state.FilterChipType:     "󰉺 Type: Feature",
				state.FilterChipAssignee: " Assignee: Jane Smith",
				state.FilterChipArchived: "󱝍 Archived: On",
				state.FilterChipClearAll: " Clear All",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildAllChipTexts(tt.props)

			assert.Len(t, result, int(state.FilterChipCount), "should return correct number of chip texts")

			for chip, expectedText := range tt.expected {
				assert.Equal(t, expectedText, result[chip], "chip %d text mismatch", chip)
			}
		})
	}
}

// TestBuildSearchChipText tests the search chip text construction.
func TestBuildSearchChipText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		query       string
		inputActive bool
		expected    string
	}{
		{
			name:        "empty query not active",
			query:       "",
			inputActive: false,
			expected:    " Search: All",
		},
		{
			name:        "query set not in input mode",
			query:       "bug",
			inputActive: false,
			expected:    " Search: bug",
		},
		{
			name:        "empty query in input mode shows cursor",
			query:       "",
			inputActive: true,
			expected:    " Search: ▎",
		},
		{
			name:        "query in input mode shows cursor after text",
			query:       "hello",
			inputActive: true,
			expected:    " Search: hello▎",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildSearchChipText(tt.query, tt.inputActive)

			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildChipText tests the basic chip text construction.
func TestBuildChipText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		value    string
		expected string
	}{
		{
			name:     "empty value shows All",
			field:    " Label",
			value:    "",
			expected: " Label: All",
		},
		{
			name:     "non-empty value",
			field:    " Priority",
			value:    "High",
			expected: " Priority: High",
		},
		{
			name:     "value with comma-separated items",
			field:    " Label",
			value:    "bug, frontend",
			expected: " Label: bug, frontend",
		},
		{
			name:     "type field with icon",
			field:    "󰉺 Type",
			value:    "Feature",
			expected: "󰉺 Type: Feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildChipText(tt.field, tt.value)

			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildArchivedComponents tests the archived chip component construction.
func TestBuildArchivedComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		isArchived     bool
		expectedPrefix string
		expectedValue  string
	}{
		{
			name:           "archived off",
			isArchived:     false,
			expectedPrefix: "󱝋 ",
			expectedValue:  "Off",
		},
		{
			name:           "archived on",
			isArchived:     true,
			expectedPrefix: "󱝍 ",
			expectedValue:  "On",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildArchivedComponents(tt.isArchived)

			assert.Equal(t, tt.expectedPrefix, result.prefix, "prefix mismatch")
			assert.Equal(t, tt.expectedValue, result.value, "value mismatch")
		})
	}
}
