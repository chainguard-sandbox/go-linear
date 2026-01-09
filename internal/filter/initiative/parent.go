package initiative

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyParent handles --parent flag.
// TODO: Linear API doesn't currently support filtering initiatives by parent in InitiativeFilter.
// The parentInitiative field exists on Initiative but not as a filter option.
func ApplyParent(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	parent, _ := cmd.Flags().GetString("parent")
	if parent == "" {
		return nil
	}

	// Placeholder - filtering by parent is not supported by Linear API yet
	return fmt.Errorf("filtering by parent initiative is not currently supported by the Linear API")
}
