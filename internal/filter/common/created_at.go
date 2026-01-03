package common

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyCreatedAt handles --created-after and --created-before flags.
// Works with any filter builder that implements DateFilterable.
func ApplyCreatedAt[T DateFilterable](ctx context.Context, cmd *cobra.Command, b T) error {
	after, _ := cmd.Flags().GetString("created-after")
	before, _ := cmd.Flags().GetString("created-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.CreatedAtComparator()

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid --created-after: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Gte = &formatted
	}

	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid --created-before: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Lte = &formatted
	}

	return nil
}
