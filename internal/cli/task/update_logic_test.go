package task

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUpdateID(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    int
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ID",
			args: []string{"42"},
			want: 42,
		},
		{
			name: "valid ID - single digit",
			args: []string{"1"},
			want: 1,
		},
		{
			name: "valid ID - large number",
			args: []string{"999999"},
			want: 999999,
		},
		{
			name:    "missing ID",
			args:    []string{},
			wantErr: true,
			errMsg:  "task ID is required",
		},
		{
			name:    "invalid ID - not a number",
			args:    []string{"abc"},
			wantErr: true,
			errMsg:  "invalid ID 'abc': must be a number",
		},
		{
			name:    "invalid ID - float",
			args:    []string{"42.5"},
			wantErr: true,
			errMsg:  "invalid ID '42.5': must be a number",
		},
		{
			name:    "invalid ID - negative",
			args:    []string{"-1"},
			wantErr: true,
			errMsg:  "task ID must be a positive integer",
		},
		{
			name:    "invalid ID - zero",
			args:    []string{"0"},
			wantErr: true,
			errMsg:  "task ID must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUpdateID(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseUpdateFlags(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*cobra.Command)
		want    *UpdateInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "title only",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("title", "New Title"))
			},
			want: &UpdateInput{
				Title: new("New Title"),
			},
		},
		{
			name: "description only",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("description", "New Description"))
			},
			want: &UpdateInput{
				Description: new("New Description"),
			},
		},
		{
			name: "priority only",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("priority", "high"))
			},
			want: &UpdateInput{
				Priority: new("high"),
			},
		},
		{
			name: "title and description",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("title", "New Title"))
				require.NoError(t, cmd.Flags().Set("description", "New Description"))
			},
			want: &UpdateInput{
				Title:       new("New Title"),
				Description: new("New Description"),
			},
		},
		{
			name: "title and priority",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("title", "New Title"))
				require.NoError(t, cmd.Flags().Set("priority", "critical"))
			},
			want: &UpdateInput{
				Title:    new("New Title"),
				Priority: new("critical"),
			},
		},
		{
			name: "description and priority",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("description", "New Description"))
				require.NoError(t, cmd.Flags().Set("priority", "low"))
			},
			want: &UpdateInput{
				Description: new("New Description"),
				Priority:    new("low"),
			},
		},
		{
			name: "all fields",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("title", "New Title"))
				require.NoError(t, cmd.Flags().Set("description", "New Description"))
				require.NoError(t, cmd.Flags().Set("priority", "medium"))
			},
			want: &UpdateInput{
				Title:       new("New Title"),
				Description: new("New Description"),
				Priority:    new("medium"),
			},
		},
		{
			name: "empty title",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("title", ""))
			},
			want: &UpdateInput{
				Title: new(""),
			},
		},
		{
			name: "empty description",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("description", ""))
			},
			want: &UpdateInput{
				Description: new(""),
			},
		},
		{
			name: "empty priority",
			setup: func(cmd *cobra.Command) {
				require.NoError(t, cmd.Flags().Set("priority", ""))
			},
			want: &UpdateInput{
				Priority: new(""),
			},
		},
		{
			name:    "no flags set",
			setup:   func(cmd *cobra.Command) {},
			wantErr: true,
			errMsg:  "at least one of --title, --description, or --priority must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().StringP("title", "t", "", "Task title")
			cmd.Flags().StringP("description", "d", "", "Task description")
			cmd.Flags().StringP("priority", "r", "", "Task priority")

			tt.setup(cmd)

			got, err := ParseUpdateFlags(cmd)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestHasContentUpdate(t *testing.T) {
	tests := []struct {
		name  string
		input *UpdateInput
		want  bool
	}{
		{
			name: "has title",
			input: &UpdateInput{
				Title: new("New Title"),
			},
			want: true,
		},
		{
			name: "has description",
			input: &UpdateInput{
				Description: new("New Description"),
			},
			want: true,
		},
		{
			name: "has both title and description",
			input: &UpdateInput{
				Title:       new("New Title"),
				Description: new("New Description"),
			},
			want: true,
		},
		{
			name: "has only priority",
			input: &UpdateInput{
				Priority: new("high"),
			},
			want: false,
		},
		{
			name:  "has nothing",
			input: &UpdateInput{},
			want:  false,
		},
		{
			name: "has empty title",
			input: &UpdateInput{
				Title: new(""),
			},
			want: true,
		},
		{
			name: "has empty description",
			input: &UpdateInput{
				Description: new(""),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasContentUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasPriorityUpdate(t *testing.T) {
	tests := []struct {
		name  string
		input *UpdateInput
		want  bool
	}{
		{
			name: "has priority",
			input: &UpdateInput{
				Priority: new("high"),
			},
			want: true,
		},
		{
			name: "has empty priority",
			input: &UpdateInput{
				Priority: new(""),
			},
			want: true,
		},
		{
			name: "has only title",
			input: &UpdateInput{
				Title: new("New Title"),
			},
			want: false,
		},
		{
			name: "has only description",
			input: &UpdateInput{
				Description: new("New Description"),
			},
			want: false,
		},
		{
			name:  "has nothing",
			input: &UpdateInput{},
			want:  false,
		},
		{
			name: "has title and description but no priority",
			input: &UpdateInput{
				Title:       new("New Title"),
				Description: new("New Description"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPriorityUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatUpdateOutput(t *testing.T) {
	tests := []struct {
		name   string
		result *UpdateResult
		want   string
	}{
		{
			name: "single digit ID",
			result: &UpdateResult{
				TaskID: 1,
			},
			want: "Task 1 updated successfully",
		},
		{
			name: "multi digit ID",
			result: &UpdateResult{
				TaskID: 42,
			},
			want: "Task 42 updated successfully",
		},
		{
			name: "large ID",
			result: &UpdateResult{
				TaskID: 999999,
			},
			want: "Task 999999 updated successfully",
		},
		{
			name: "zero ID",
			result: &UpdateResult{
				TaskID: 0,
			},
			want: "Task 0 updated successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatUpdateOutput(tt.result)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatUpdateJSON(t *testing.T) {
	tests := []struct {
		name   string
		result *UpdateResult
		want   map[string]any
	}{
		{
			name: "single digit ID",
			result: &UpdateResult{
				TaskID: 1,
			},
			want: map[string]any{
				"success": true,
				"task_id": 1,
			},
		},
		{
			name: "multi digit ID",
			result: &UpdateResult{
				TaskID: 42,
			},
			want: map[string]any{
				"success": true,
				"task_id": 42,
			},
		},
		{
			name: "large ID",
			result: &UpdateResult{
				TaskID: 999999,
			},
			want: map[string]any{
				"success": true,
				"task_id": 999999,
			},
		},
		{
			name: "zero ID",
			result: &UpdateResult{
				TaskID: 0,
			},
			want: map[string]any{
				"success": true,
				"task_id": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatUpdateJSON(tt.result)
			assert.Equal(t, tt.want, got)
		})
	}
}
