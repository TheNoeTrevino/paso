package standuplog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		projectID   int
		content     string
		expectedErr error
	}{
		{
			name:        "valid params",
			projectID:   1,
			content:     "Fixed the auth bug",
			expectedErr: nil,
		},
		{
			name:        "zero project ID",
			projectID:   0,
			content:     "Some content",
			expectedErr: ErrInvalidProjectID,
		},
		{
			name:        "negative project ID",
			projectID:   -1,
			content:     "Some content",
			expectedErr: ErrInvalidProjectID,
		},
		{
			name:        "empty content",
			projectID:   1,
			content:     "",
			expectedErr: ErrEmptyContent,
		},
		{
			name:        "both invalid",
			projectID:   0,
			content:     "",
			expectedErr: ErrInvalidProjectID,
		},
		{
			name:        "long content is valid",
			projectID:   1,
			content:     string(make([]byte, 5000)),
			expectedErr: nil,
		},
		{
			name:        "content with newlines",
			projectID:   1,
			content:     "Line 1\nLine 2\nLine 3",
			expectedErr: nil,
		},
		{
			name:        "content with unicode",
			projectID:   1,
			content:     "Fixed bug in 日本語 module",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateCreateParams(tt.projectID, tt.content)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
