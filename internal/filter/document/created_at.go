package document

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
)

// ApplyCreatedAt handles --created-after and --created-before flags.
func ApplyCreatedAt(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	after, _ := cmd.Flags().GetString("created-after")
	before, _ := cmd.Flags().GetString("created-before")
	if after == "" && before == "" {
		return nil
	}

	comp := b.CreatedAtComparator()
	dr := &filter.DateRange{Gte: comp.Gte, Lt: comp.Lt}
	if err := filter.ApplyDateRange(b.Parser(), dr, after, before, "created"); err != nil {
		return err
	}
	comp.Gte, comp.Lt = dr.Gte, dr.Lt
	return nil
}
