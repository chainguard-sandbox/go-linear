package filter

import (
	"context"

	"github.com/spf13/cobra"
)

// ApplyUpdatedAt handles --updated-after and --updated-before flags.
func ApplyUpdatedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("updated-after")
	before, _ := cmd.Flags().GetString("updated-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.UpdatedAtComparator()
	dr := &DateRange{Gte: comp.Gte, Lt: comp.Lt}

	if err := ApplyDateRange(b.Parser(), dr, after, before, "updated"); err != nil {
		return err
	}

	comp.Gte = dr.Gte
	comp.Lt = dr.Lt
	return nil
}
