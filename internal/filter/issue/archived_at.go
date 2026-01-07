package issue

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
)

// ApplyArchivedAt handles --archived-after and --archived-before flags.
func ApplyArchivedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("archived-after")
	before, _ := cmd.Flags().GetString("archived-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.ArchivedAtComparator()
	dr := &filter.DateRange{Gte: comp.Gte, Lt: comp.Lt}

	if err := filter.ApplyDateRange(b.Parser(), dr, after, before, "archived"); err != nil {
		return err
	}

	comp.Gte = dr.Gte
	comp.Lt = dr.Lt
	return nil
}
