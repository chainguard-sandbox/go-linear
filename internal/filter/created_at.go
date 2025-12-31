package filter

import (
	"context"

	"github.com/spf13/cobra"
)

// ApplyCreatedAt handles --created-after and --created-before flags for issues.
func ApplyCreatedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("created-after")
	before, _ := cmd.Flags().GetString("created-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.CreatedAtComparator()
	dr := &DateRange{Gte: comp.Gte, Lt: comp.Lt}

	if err := ApplyDateRange(b.Parser(), dr, after, before, "created"); err != nil {
		return err
	}

	comp.Gte = dr.Gte
	comp.Lt = dr.Lt
	return nil
}
