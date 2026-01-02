package project

import (
	"context"

	"github.com/spf13/cobra"
)

// FilterFunc applies a filter based on command flags.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// Filters is the registry of all project filter functions.
var Filters = map[string]FilterFunc{
	// Date filters
	"created-at":   ApplyCreatedAt,
	"updated-at":   ApplyUpdatedAt,
	"completed-at": ApplyCompletedAt,
	"canceled-at":  ApplyCanceledAt,
	"started-at":   ApplyStartedAt,
	"target-date":  ApplyTargetDate,

	// Entity filters
	"id":      ApplyID,
	"creator": ApplyCreator,
	"lead":    ApplyLead,

	// Text filters
	"name":    ApplyName,
	"slug-id": ApplySlugID,

	// State filters
	"health": ApplyHealth,

	// Numeric filters
	"priority": ApplyPriority,

	// Relation filters
	"has-blocked-by": ApplyHasBlockedBy,
	"has-blocking":   ApplyHasBlocking,
	"has-related":    ApplyHasRelated,
}
