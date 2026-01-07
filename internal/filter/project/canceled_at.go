package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyCanceledAt handles --canceled-after and --canceled-before flags.
func ApplyCanceledAt(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	after, _ := cmd.Flags().GetString("canceled-after")
	before, _ := cmd.Flags().GetString("canceled-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.CanceledAtComparator()

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid --canceled-after: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Gte = &formatted
	}

	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid --canceled-before: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Lte = &formatted
	}

	return nil
}
