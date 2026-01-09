package initiative

import (
	"github.com/chainguard-sandbox/go-linear/internal/filter/common"
)

// Filters maps filter names to their apply functions.
// Uses shared filters from common package where applicable.
var Filters = map[string]FilterFunc{
	// Date filters (created-at is shared)
	"created-at":  common.ApplyCreatedAt[*FilterBuilder],
	"target-date": ApplyTargetDate,

	// Entity filters (id is shared)
	"id":      common.ApplyID[*FilterBuilder],
	"creator": ApplyCreator,
	"owner":   ApplyOwner,
	// "parent":  ApplyParent, // TODO: Linear API doesn't support parent filtering yet

	// Text filters (initiative-specific)
	"name":    ApplyName,
	"slug-id": ApplySlugID,

	// State filters (initiative-specific)
	"health": ApplyHealth,
	"status": ApplyStatus,
}
