package handler

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestParseFlagsToMap(t *testing.T) {
	t.Run("Parses string flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("name", "", "test name")
		cmd.SetArgs([]string{"--name", "hello"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, "hello", flags["name"])
	})

	t.Run("Parses int flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().Int("count", 0, "test count")
		cmd.SetArgs([]string{"--count", "42"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, 42, flags["count"])
	})

	t.Run("Parses bool flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().Bool("verbose", false, "verbose output")
		cmd.SetArgs([]string{"--verbose"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, true, flags["verbose"])
	})

	t.Run("Parses float64 flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().Float64("rate", 0.0, "rate")
		cmd.SetArgs([]string{"--rate", "3.14"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, 3.14, flags["rate"])
	})

	t.Run("Parses string slice flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().StringSlice("tags", nil, "tags")
		cmd.SetArgs([]string{"--tags", "a,b,c"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, []string{"a", "b", "c"}, flags["tags"])
	})

	t.Run("Parses int slice flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().IntSlice("ids", nil, "ids")
		cmd.SetArgs([]string{"--ids", "1,2,3"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, []int{1, 2, 3}, flags["ids"])
	})

	t.Run("Only includes explicitly set flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("set-flag", "", "set")
		cmd.Flags().String("unset-flag", "default", "unset")
		cmd.SetArgs([]string{"--set-flag", "value"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, "value", flags["set-flag"])
		_, exists := flags["unset-flag"]
		assert.False(t, exists, "Unset flags should not appear in map")
	})

	t.Run("Parses int64 flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test", Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().Int64("big", 0, "big number")
		cmd.SetArgs([]string{"--big", "9999999999"})
		_ = cmd.Execute()

		flags := parseFlagsToMap(cmd)
		assert.Equal(t, int64(9999999999), flags["big"])
	})
}

func TestArguments_GetString(t *testing.T) {
	args := &Arguments{
		Flags: map[string]any{
			"name": "hello",
			"bad":  42,
		},
	}

	t.Run("Returns value when present", func(t *testing.T) {
		assert.Equal(t, "hello", args.GetString("name", "default"))
	})

	t.Run("Returns default when missing", func(t *testing.T) {
		assert.Equal(t, "default", args.GetString("missing", "default"))
	})

	t.Run("Returns default when wrong type", func(t *testing.T) {
		assert.Equal(t, "default", args.GetString("bad", "default"))
	})
}

func TestArguments_GetInt(t *testing.T) {
	args := &Arguments{
		Flags: map[string]any{
			"count": 42,
			"bad":   "not-int",
		},
	}

	t.Run("Returns value when present", func(t *testing.T) {
		assert.Equal(t, 42, args.GetInt("count", 0))
	})

	t.Run("Returns default when missing", func(t *testing.T) {
		assert.Equal(t, 99, args.GetInt("missing", 99))
	})

	t.Run("Returns default when wrong type", func(t *testing.T) {
		assert.Equal(t, 0, args.GetInt("bad", 0))
	})
}

func TestArguments_GetBool(t *testing.T) {
	args := &Arguments{
		Flags: map[string]any{
			"verbose": true,
			"bad":     "not-bool",
		},
	}

	t.Run("Returns value when present", func(t *testing.T) {
		assert.True(t, args.GetBool("verbose"))
	})

	t.Run("Returns false when missing", func(t *testing.T) {
		assert.False(t, args.GetBool("missing"))
	})

	t.Run("Returns false when wrong type", func(t *testing.T) {
		assert.False(t, args.GetBool("bad"))
	})
}

func TestArguments_GetStringSlice(t *testing.T) {
	args := &Arguments{
		Flags: map[string]any{
			"tags": []string{"a", "b"},
			"bad":  "not-slice",
		},
	}

	t.Run("Returns value when present", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b"}, args.GetStringSlice("tags", nil))
	})

	t.Run("Returns default when missing", func(t *testing.T) {
		assert.Equal(t, []string{"default"}, args.GetStringSlice("missing", []string{"default"}))
	})

	t.Run("Returns default when wrong type", func(t *testing.T) {
		assert.Nil(t, args.GetStringSlice("bad", nil))
	})
}

func TestArguments_GetIntSlice(t *testing.T) {
	args := &Arguments{
		Flags: map[string]any{
			"ids": []int{1, 2, 3},
			"bad": "not-slice",
		},
	}

	t.Run("Returns value when present", func(t *testing.T) {
		assert.Equal(t, []int{1, 2, 3}, args.GetIntSlice("ids", nil))
	})

	t.Run("Returns default when missing", func(t *testing.T) {
		assert.Equal(t, []int{99}, args.GetIntSlice("missing", []int{99}))
	})

	t.Run("Returns default when wrong type", func(t *testing.T) {
		assert.Nil(t, args.GetIntSlice("bad", nil))
	})
}

func TestArguments_GetCmd(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	args := &Arguments{
		Flags: map[string]any{},
		cmd:   cmd,
	}

	assert.Equal(t, cmd, args.GetCmd())
}

func TestCommand_WithMockHandler(t *testing.T) {
	t.Run("Executes handler and formats output", func(t *testing.T) {
		handler := &mockHandler{result: map[string]any{"id": 1}}
		parseFlags := func(cmd *cobra.Command) error { return nil }

		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("json", false, "json output")
		cmd.Flags().Bool("quiet", false, "quiet mode")
		cmd.RunE = Command(handler, parseFlags)

		cmd.SetArgs([]string{})
		err := cmd.ExecuteContext(context.Background())

		assert.NoError(t, err)
		assert.True(t, handler.executed)
	})

	t.Run("Returns error from handler", func(t *testing.T) {
		handler := &mockHandler{err: assert.AnError}
		parseFlags := func(cmd *cobra.Command) error { return nil }

		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("json", false, "json output")
		cmd.Flags().Bool("quiet", false, "quiet mode")
		cmd.RunE = Command(handler, parseFlags)
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		cmd.SetArgs([]string{})
		err := cmd.ExecuteContext(context.Background())

		assert.Error(t, err)
	})

	t.Run("Returns error from parseFlags", func(t *testing.T) {
		handler := &mockHandler{result: "ok"}
		parseFlags := func(cmd *cobra.Command) error { return assert.AnError }

		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("json", false, "json output")
		cmd.Flags().Bool("quiet", false, "quiet mode")
		cmd.RunE = Command(handler, parseFlags)
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		cmd.SetArgs([]string{})
		err := cmd.ExecuteContext(context.Background())

		assert.Error(t, err)
		assert.False(t, handler.executed)
	})
}

func TestSimpleCommand(t *testing.T) {
	handler := &mockHandler{result: map[string]any{"ok": true}}

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("json", false, "json output")
	cmd.Flags().Bool("quiet", false, "quiet mode")
	cmd.RunE = SimpleCommand(handler)

	cmd.SetArgs([]string{})
	err := cmd.ExecuteContext(context.Background())

	assert.NoError(t, err)
	assert.True(t, handler.executed)
}

type mockHandler struct {
	result   any
	err      error
	executed bool
}

func (m *mockHandler) Execute(ctx context.Context, args *Arguments) (any, error) {
	m.executed = true
	return m.result, m.err
}
