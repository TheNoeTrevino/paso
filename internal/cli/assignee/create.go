package assignee

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/handler"
)

// CreateCmd returns the assignee create subcommand
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new assignee",
		Long: `Create a new assignee with a name.

Examples:
  # Create assignee using shorthand flags
  paso assignee create -n "john"

  # JSON output for agents
  paso assignee create -n "john" -j

  # Quiet mode for bash capture
  ASSIGNEE_ID=$(paso assignee create -n "john" -q)

  # Long-form flags also supported
  paso assignee create --name="john"
`,
		RunE: handler.Command(&createHandler{}, parseCreateFlags),
	}

	// Required flags
	cmd.Flags().StringP("name", "n", "", "Assignee name (required)")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		slog.Error("failed to marking flag as required", "error", err)
	}

	// Agent-friendly flags
	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().BoolP("quiet", "q", false, "Minimal output (ID only)")

	return cmd
}

// createHandler implements handler.Handler for assignee creation
type createHandler struct{}

// Execute implements the Handler interface
func (h *createHandler) Execute(ctx context.Context, args *handler.Arguments) (any, error) {
	assigneeName := args.MustGetString("name")

	cliInstance, err := cli.GetCLIFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization error: %w", err)
	}
	defer func() {
		if err := cliInstance.Close(); err != nil {
			slog.Error("failed to closing CLI", "error", err)
		}
	}()

	assignee, err := cliInstance.App.AssigneeService.Create(ctx, assigneeName)
	if err != nil {
		return nil, fmt.Errorf("assignee creation error: %w", err)
	}

	return &assigneeCreateResult{
		ID:   assignee.ID,
		Name: assignee.Name,
	}, nil
}

// assigneeCreateResult represents the result of assignee creation
type assigneeCreateResult struct {
	ID   int
	Name string
}

// GetID implements the GetID interface for quiet mode output
func (r *assigneeCreateResult) GetID() int {
	return r.ID
}

func parseCreateFlags(cmd *cobra.Command) error {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return fmt.Errorf("assignee name is required")
	}

	return nil
}
