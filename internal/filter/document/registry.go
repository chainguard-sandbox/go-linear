package document

import (
	"github.com/chainguard-sandbox/go-linear/v2/internal/filter/common"
)

// Filters maps filter names to their apply functions.
// Uses shared filters from common package where applicable.
var Filters = map[string]FilterFunc{
	// Date filters (shared)
	"created-at": common.ApplyCreatedAt[*FilterBuilder],
	"updated-at": common.ApplyUpdatedAt[*FilterBuilder],

	// Entity filters (id is shared)
	"id":         common.ApplyID[*FilterBuilder],
	"creator":    ApplyCreator,
	"initiative": ApplyInitiative,
	"project":    ApplyProject,
	"issue":      ApplyIssue,

	// Text filters (document-specific)
	"title":   ApplyTitle,
	"slug-id": ApplySlugID,
}
