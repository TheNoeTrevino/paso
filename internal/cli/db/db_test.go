package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatAddMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		dbName    string
		connected bool
		want      string
	}{
		{
			name:      "connected successfully",
			dbName:    "production",
			connected: true,
			want:      "Database 'production' added successfully",
		},
		{
			name:      "connection failed",
			dbName:    "staging",
			connected: false,
			want:      "Database 'staging' saved (connection test failed)",
		},
		{
			name:      "db with special chars connected",
			dbName:    "my-db_123",
			connected: true,
			want:      "Database 'my-db_123' added successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatAddMessage(tt.dbName, tt.connected)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFormatConnectMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		dbName    string
		connected bool
		want      string
	}{
		{
			name:      "switched and connected",
			dbName:    "production",
			connected: true,
			want:      "Connected to 'production'",
		},
		{
			name:      "switched but connection failed",
			dbName:    "staging",
			connected: false,
			want:      "Switched to 'staging' (connection test failed)",
		},
		{
			name:      "db with special chars",
			dbName:    "test_db-2",
			connected: true,
			want:      "Connected to 'test_db-2'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatConnectMessage(tt.dbName, tt.connected)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestAddCmd_FlagValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "both remote and local",
			args:    []string{"postgres://localhost/db", "--name", "test", "--remote", "--local"},
			wantErr: true,
			errMsg:  "cannot specify both --remote and --local",
		},
		{
			name:    "neither remote nor local",
			args:    []string{"postgres://localhost/db", "--name", "test"},
			wantErr: true,
			errMsg:  "either --remote or --local must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := AddCmd()
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDbCmd_Subcommands(t *testing.T) {
	t.Parallel()
	cmd := DbCmd()

	// Verify all expected subcommands are registered
	expectedSubcommands := map[string]bool{
		"add":     false,
		"connect": false,
		"list":    false,
		"remove":  false,
	}

	for _, subCmd := range cmd.Commands() {
		if _, exists := expectedSubcommands[subCmd.Name()]; exists {
			expectedSubcommands[subCmd.Name()] = true
		}
	}

	for name, found := range expectedSubcommands {
		assert.True(t, found, "Expected subcommand '%s' not found", name)
	}
}
