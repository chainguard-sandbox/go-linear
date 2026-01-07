package common

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyUpdatedAt handles --updated-after and --updated-before flags.
// Works with any filter builder that implements UpdateDateFilterable.
func ApplyUpdatedAt[T UpdateDateFilterable](ctx context.Context, cmd *cobra.Command, b T) error {
	after, _ := cmd.Flags().GetString("updated-after")
	before, _ := cmd.Flags().GetString("updated-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.UpdatedAtComparator()

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid --updated-after: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Gte = &formatted
	}

	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid --updated-before: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Lte = &formatted
	}

	return nil
}
