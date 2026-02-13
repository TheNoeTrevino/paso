package task

import (
	"testing"
)

func TestParseReadyMoveArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        *ReadyMoveInput
		wantErr     bool
		errContains string
	}{
		{
			name: "valid task ID",
			args: []string{"42"},
			want: &ReadyMoveInput{
				TaskID: 42,
			},
			wantErr: false,
		},
		{
			name: "valid task ID - single digit",
			args: []string{"1"},
			want: &ReadyMoveInput{
				TaskID: 1,
			},
			wantErr: false,
		},
		{
			name: "valid task ID - large number",
			args: []string{"999999"},
			want: &ReadyMoveInput{
				TaskID: 999999,
			},
			wantErr: false,
		},
		{
			name:        "empty args",
			args:        []string{},
			want:        nil,
			wantErr:     true,
			errContains: "task ID is required",
		},
		{
			name:        "nil args",
			args:        nil,
			want:        nil,
			wantErr:     true,
			errContains: "task ID is required",
		},
		{
			name:        "invalid task ID - not a number",
			args:        []string{"abc"},
			want:        nil,
			wantErr:     true,
			errContains: "invalid task ID: abc",
		},
		{
			name:        "invalid task ID - empty string",
			args:        []string{""},
			want:        nil,
			wantErr:     true,
			errContains: "invalid task ID: ",
		},
		{
			name:        "invalid task ID - negative number",
			args:        []string{"-1"},
			want:        nil,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "invalid task ID - decimal",
			args:        []string{"42.5"},
			want:        nil,
			wantErr:     true,
			errContains: "invalid task ID: 42.5",
		},
		{
			name:        "invalid task ID - with spaces",
			args:        []string{"42 "},
			want:        nil,
			wantErr:     true,
			errContains: "invalid task ID: 42 ",
		},
		{
			name: "extra args ignored",
			args: []string{"42", "extra", "args"},
			want: &ReadyMoveInput{
				TaskID: 42,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReadyMoveArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReadyMoveArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseReadyMoveArgs() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if !tt.wantErr && got != nil && tt.want != nil {
				if got.TaskID != tt.want.TaskID {
					t.Errorf("ParseReadyMoveArgs() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFormatReadyMoveOutput(t *testing.T) {
	tests := []struct {
		name   string
		result *ReadyMoveResult
		want   string
	}{
		{
			name: "basic move",
			result: &ReadyMoveResult{
				TaskID:     42,
				FromColumn: "Backlog",
				ToColumn:   "Ready",
			},
			want: "Task 42 moved to 'Ready'",
		},
		{
			name: "move with special characters in column name",
			result: &ReadyMoveResult{
				TaskID:     1,
				FromColumn: "To Do",
				ToColumn:   "Ready & Waiting",
			},
			want: "Task 1 moved to 'Ready & Waiting'",
		},
		{
			name: "large task ID",
			result: &ReadyMoveResult{
				TaskID:     999999,
				FromColumn: "Backlog",
				ToColumn:   "Ready",
			},
			want: "Task 999999 moved to 'Ready'",
		},
		{
			name: "empty column names",
			result: &ReadyMoveResult{
				TaskID:     42,
				FromColumn: "",
				ToColumn:   "",
			},
			want: "Task 42 moved to ''",
		},
		{
			name: "unicode in column names",
			result: &ReadyMoveResult{
				TaskID:     42,
				FromColumn: "待办",
				ToColumn:   "就绪",
			},
			want: "Task 42 moved to '就绪'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatReadyMoveOutput(tt.result)
			if got != tt.want {
				t.Errorf("FormatReadyMoveOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatReadyMoveJSON(t *testing.T) {
	tests := []struct {
		name   string
		result *ReadyMoveResult
		want   map[string]any
	}{
		{
			name: "basic move",
			result: &ReadyMoveResult{
				TaskID:     42,
				FromColumn: "Backlog",
				ToColumn:   "Ready",
			},
			want: map[string]any{
				"success":     true,
				"task_id":     42,
				"from_column": "Backlog",
				"to_column":   "Ready",
			},
		},
		{
			name: "move with empty columns",
			result: &ReadyMoveResult{
				TaskID:     1,
				FromColumn: "",
				ToColumn:   "",
			},
			want: map[string]any{
				"success":     true,
				"task_id":     1,
				"from_column": "",
				"to_column":   "",
			},
		},
		{
			name: "move with special characters",
			result: &ReadyMoveResult{
				TaskID:     999,
				FromColumn: "To-Do & Backlog",
				ToColumn:   "Ready/Waiting",
			},
			want: map[string]any{
				"success":     true,
				"task_id":     999,
				"from_column": "To-Do & Backlog",
				"to_column":   "Ready/Waiting",
			},
		},
		{
			name: "move with unicode",
			result: &ReadyMoveResult{
				TaskID:     42,
				FromColumn: "待办",
				ToColumn:   "就绪",
			},
			want: map[string]any{
				"success":     true,
				"task_id":     42,
				"from_column": "待办",
				"to_column":   "就绪",
			},
		},
		{
			name: "zero task ID",
			result: &ReadyMoveResult{
				TaskID:     0,
				FromColumn: "Backlog",
				ToColumn:   "Ready",
			},
			want: map[string]any{
				"success":     true,
				"task_id":     0,
				"from_column": "Backlog",
				"to_column":   "Ready",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatReadyMoveJSON(tt.result)
			if !mapsEqual(got, tt.want) {
				t.Errorf("FormatReadyMoveJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatReadyMoveQuiet(t *testing.T) {
	tests := []struct {
		name   string
		result *ReadyMoveResult
		want   string
	}{
		{
			name: "basic task ID",
			result: &ReadyMoveResult{
				TaskID:     42,
				FromColumn: "In Progress",
				ToColumn:   "Ready",
			},
			want: "42\n",
		},
		{
			name: "single digit task ID",
			result: &ReadyMoveResult{
				TaskID:     1,
				FromColumn: "Backlog",
				ToColumn:   "Ready",
			},
			want: "1\n",
		},
		{
			name: "large task ID",
			result: &ReadyMoveResult{
				TaskID:     999999,
				FromColumn: "Done",
				ToColumn:   "Ready",
			},
			want: "999999\n",
		},
		{
			name: "zero task ID",
			result: &ReadyMoveResult{
				TaskID:     0,
				FromColumn: "Backlog",
				ToColumn:   "Ready",
			},
			want: "0\n",
		},
		{
			name: "negative task ID",
			result: &ReadyMoveResult{
				TaskID:     -1,
				FromColumn: "Backlog",
				ToColumn:   "Ready",
			},
			want: "-1\n",
		},
		{
			name: "columns don't matter",
			result: &ReadyMoveResult{
				TaskID:     42,
				FromColumn: "",
				ToColumn:   "",
			},
			want: "42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatReadyMoveQuiet(tt.result)
			if got != tt.want {
				t.Errorf("FormatReadyMoveQuiet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
