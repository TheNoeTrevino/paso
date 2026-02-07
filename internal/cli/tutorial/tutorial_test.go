package tutorial

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestTutorialCmd(t *testing.T) {
	t.Run("Outputs tutorial content", func(t *testing.T) {
		cmd := TutorialCmd()

		output := testutil.CaptureOutput(t, func() {
			cmd.Run(cmd, []string{})
		})

		assert.NotEmpty(t, output)
		// The tutorial content should contain paso-related information
		assert.Contains(t, output, "paso")
	})

	t.Run("Command metadata is correct", func(t *testing.T) {
		cmd := TutorialCmd()
		assert.Equal(t, "tutorial", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})
}
