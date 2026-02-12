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
