package cycle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/filter/common"
)

// FilterFunc is the signature for cycle filter functions.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// Filters maps filter names to their implementations.
// Uses shared filters from common package where applicable.
var Filters = map[string]FilterFunc{
	// Date filters (created-at and updated-at are shared)
	"created-at":   common.ApplyCreatedAt[*FilterBuilder],
	"updated-at":   common.ApplyUpdatedAt[*FilterBuilder],
	"completed-at": ApplyCompletedAt,
	"starts-at":    ApplyStartsAt,
	"ends-at":      ApplyEndsAt,

	// Boolean state filters (cycle-specific)
	"is-active":      ApplyIsActive,
	"is-future":      ApplyIsFuture,
	"is-past":        ApplyIsPast,
	"is-next":        ApplyIsNext,
	"is-previous":    ApplyIsPrevious,
	"is-in-cooldown": ApplyIsInCooldown,

	// Entity filters (id is shared)
	"id":     common.ApplyID[*FilterBuilder],
	"name":   ApplyName,
	"number": ApplyNumber,
	"team":   ApplyTeam,
}

// FromFlags builds a filter from command flags using the registry.
func (b *FilterBuilder) FromFlags(ctx context.Context, cmd *cobra.Command) error {
	for _, apply := range Filters {
		if err := apply(ctx, cmd, b); err != nil {
			return err
		}
	}
	return nil
}
