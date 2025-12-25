package label

import (
	"github.com/spf13/cobra"
)

// LabelCmd returns the label parent command
func LabelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label",
		Short: "Manage labels",
	}

	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(UpdateCmd())
	cmd.AddCommand(DeleteCmd())
	cmd.AddCommand(AttachCmd())
	cmd.AddCommand(DetachCmd())

	return cmd
}
