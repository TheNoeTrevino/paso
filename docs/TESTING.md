# Testing Conventions

## Naming

Format:
`TestActionEntity_Scenario` 

For Example: 
- `TestCreateTask_BasicCreate`
- `TestDeleteProject_NegativeId`

## Test Structure

Use table driven tests where applicable, especially for testing multiple scenarios of the same action.

For example:

``` go
func TestUpdateLabel_InvalidLabelID_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		labelID  int
		newName  *string
		newColor *string
		wantErr  error
	}{
		{
			name:    "negative label ID",
			labelID: -1,
			newName: ptrStr("Updated Bug"),
			wantErr: ErrInvalidLabelID,
		},
		{
			name:    "zero label ID",
			labelID: 0,
			newName: ptrStr("Updated Bug"),
			wantErr: ErrInvalidLabelID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := fixtures.SetupTestDB(t)

			svc := newTestService(t, db)
			req := UpdateLabelRequest{
				ID:    tt.labelID,
				Name:  tt.newName,
				Color: tt.newColor,
			}

			err := svc.UpdateLabel(context.Background(), req)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
```

## Snapshot / Golden File Tests

All snapshot and golden file tests live under `internal/testing/snapshots/`:

```
internal/testing/snapshots/
  snapshot.go               -- Shared Helper (package snapshots)

internal/testing/snapshots/cli/
  help_test.go              -- CLI help output golden tests
  styles_test.go            -- CLI styled output golden tests (all color themes)
  testdata/help/*.golden    -- Help output golden files
  testdata/styles/*.golden  -- Styled output golden files

internal/testing/snapshots/tui/
  snapshots_test.go         -- TUI rendering snapshot tests
  testdata/*.golden         -- TUI rendering golden files
```

All snapshot tests use a single shared helper (`snapshots.NewHelper`) and a unified
environment variable for updating golden files:

```bash
# Update all snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/...

# Update only CLI snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/cli/...

# Update only TUI snapshots
UPDATE_SNAPSHOTS=1 go test ./internal/testing/snapshots/tui/...
```

When a snapshot test fails, it means the output has changed. Either update the golden
file (if the change is intentional) or fix the regression.
