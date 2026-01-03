package project

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter/common"
)

// FilterFunc applies a filter based on command flags.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// Filters is the registry of all project filter functions.
// Uses shared filters from common package where applicable.
var Filters = map[string]FilterFunc{
	// Date filters (created-at and updated-at are shared)
	"created-at":   common.ApplyCreatedAt[*FilterBuilder],
	"updated-at":   common.ApplyUpdatedAt[*FilterBuilder],
	"completed-at": ApplyCompletedAt,
	"canceled-at":  ApplyCanceledAt,
	"started-at":   ApplyStartedAt,
	"target-date":  ApplyTargetDate,

	// Entity filters (id is shared)
	"id":      common.ApplyID[*FilterBuilder],
	"creator": ApplyCreator,
	"lead":    ApplyLead,

	// Text filters (project-specific)
	"name":    ApplyName,
	"slug-id": ApplySlugID,

	// State filters (project-specific)
	"health": ApplyHealth,

	// Numeric filters (project-specific)
	"priority": ApplyPriority,

	// Relation filters (project-specific)
	"has-blocked-by": ApplyHasBlockedBy,
	"has-blocking":   ApplyHasBlocking,
	"has-related":    ApplyHasRelated,
}
