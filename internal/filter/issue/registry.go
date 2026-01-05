package issue

import (
	"context"

	"github.com/spf13/cobra"
)

// IssueFilterFunc is the signature for issue filter functions.
type IssueFilterFunc func(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error

// IssueFilters maps filter names to their implementations.
// Each filter is a standalone function in its own file.
var IssueFilters = map[string]IssueFilterFunc{
	// Date filters (shared pattern, reusable for other entities)
	"created-at":   ApplyCreatedAt,
	"updated-at":   ApplyUpdatedAt,
	"completed-at": ApplyCompletedAt,
	"started-at":   ApplyStartedAt,
	"archived-at":  ApplyArchivedAt,

	// Issue-specific filters
	"added-to-cycle-at":        ApplyAddedToCycleAt,
	"added-to-cycle-period":    ApplyAddedToCyclePeriod,
	"assignee":                 ApplyAssignee,
	"attachments":              ApplyAttachments,
	"auto-archived-at":         ApplyAutoArchivedAt,
	"auto-closed-at":           ApplyAutoClosedAt,
	"canceled-at":              ApplyCanceledAt,
	"children":                 ApplyChildren,
	"comments":                 ApplyComments,
	"creator":                  ApplyCreator,
	"customer-count":           ApplyCustomerCount,
	"customer-important-count": ApplyCustomerImportantCount,
	"cycle":                    ApplyCycle,
	"delegate":                 ApplyDelegate,
	"description":              ApplyDescription,
	"due-date":                 ApplyDueDate,
	"estimate":                 ApplyEstimate,
	"has-blocked-by":           ApplyHasBlockedByRelations,
	"has-blocking":             ApplyHasBlockingRelations,
	"has-duplicate":            ApplyHasDuplicateRelations,
	"has-related":              ApplyHasRelatedRelations,
	"has-suggested-assignees":  ApplyHasSuggestedAssignees,
	"has-suggested-labels":     ApplyHasSuggestedLabels,
	"has-suggested-projects":   ApplyHasSuggestedProjects,
	"has-suggested-teams":      ApplyHasSuggestedTeams,
	"id":                       ApplyID,
	"labels":                   ApplyLabels,
	"last-applied-template":    ApplyLastAppliedTemplate,
	"needs":                    ApplyNeeds,
	"number":                   ApplyNumber,
	"parent":                   ApplyParent,
	"priority":                 ApplyPriority,
	"project":                  ApplyProject,
	"project-milestone":        ApplyProjectMilestone,
	"reactions":                ApplyReactions,
	"sla-status":               ApplySLAStatus,
	"snoozed-by":               ApplySnoozedBy,
	"snoozed-until-at":         ApplySnoozedUntilAt,
	"state":                    ApplyState,
	"subscribers":              ApplySubscribers,
	"team":                     ApplyTeam,
	"title":                    ApplyTitle,
	"triaged-at":               ApplyTriagedAt,
}

// FromFlags builds a filter from command flags using the registry.
// Iterates through all registered filters and applies them.
func (b *IssueFilterBuilder) FromFlags(ctx context.Context, cmd *cobra.Command) error {
	for _, apply := range IssueFilters {
		if err := apply(ctx, cmd, b); err != nil {
			return err
		}
	}
	return nil
}
