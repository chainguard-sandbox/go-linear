// Package filter provides utilities to build GraphQL filters from CLI flags.
package filter

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// IssueFilterBuilder builds an IssueFilter from Cobra command flags.
type IssueFilterBuilder struct {
	filter   *intgraphql.IssueFilter
	resolver *resolver.Resolver
	parser   *dateparser.Parser
}

// NewIssueFilterBuilder creates a new filter builder.
func NewIssueFilterBuilder(res *resolver.Resolver) *IssueFilterBuilder {
	return &IssueFilterBuilder{
		filter:   &intgraphql.IssueFilter{},
		resolver: res,
		parser:   dateparser.New(),
	}
}

// FromFlags builds a filter from command flags.
// Resolves names to IDs and parses dates.
// Filters processed in alphabetical order for maintainability.
func (b *IssueFilterBuilder) FromFlags(ctx context.Context, cmd *cobra.Command) error {
	// 1. accumulatedStateUpdatedAt - [Internal] - skip

	// 2. addedToCycleAt
	if after, _ := cmd.Flags().GetString("added-to-cycle-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid added-to-cycle-after date: %w", err)
		}
		if b.filter.AddedToCycleAt == nil {
			b.filter.AddedToCycleAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.AddedToCycleAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("added-to-cycle-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid added-to-cycle-before date: %w", err)
		}
		if b.filter.AddedToCycleAt == nil {
			b.filter.AddedToCycleAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.AddedToCycleAt.Lt = &tStr
	}

	// 3. addedToCyclePeriod
	if period, _ := cmd.Flags().GetString("added-to-cycle-period"); period != "" {
		b.filter.AddedToCyclePeriod = &intgraphql.CyclePeriodComparator{
			Eq: (*intgraphql.CyclePeriod)(&period),
		}
	}

	// 4. ageTime - [Internal] - skip

	// 5. and - compound filter [IssueFilter!] - skip (requires recursive filter building)

	// 6. archivedAt
	if after, _ := cmd.Flags().GetString("archived-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid archived-after date: %w", err)
		}
		if b.filter.ArchivedAt == nil {
			b.filter.ArchivedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.ArchivedAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("archived-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid archived-before date: %w", err)
		}
		if b.filter.ArchivedAt == nil {
			b.filter.ArchivedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.ArchivedAt.Lt = &tStr
	}

	// 7. assignee
	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		userID, err := b.resolver.ResolveUser(ctx, assignee)
		if err != nil {
			return fmt.Errorf("failed to resolve assignee: %w", err)
		}
		b.filter.Assignee = &intgraphql.NullableUserFilter{
			ID: &intgraphql.IDComparator{
				Eq: &userID,
			},
		}
	}

	// 8. attachments (by creator)
	attachmentBy, _ := cmd.Flags().GetStringArray("attachment-by")
	if len(attachmentBy) > 0 {
		userIDs := make([]string, 0, len(attachmentBy))
		for _, user := range attachmentBy {
			userID, err := b.resolver.ResolveUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to resolve attachment-by user %q: %w", user, err)
			}
			userIDs = append(userIDs, userID)
		}

		// Has attachments by any of these users
		// Optimize: single user uses Eq, multiple uses In
		if len(userIDs) == 1 {
			b.filter.Attachments = &intgraphql.AttachmentCollectionFilter{
				Some: &intgraphql.AttachmentFilter{
					Creator: &intgraphql.NullableUserFilter{
						ID: &intgraphql.IDComparator{
							Eq: &userIDs[0],
						},
					},
				},
			}
		} else {
			b.filter.Attachments = &intgraphql.AttachmentCollectionFilter{
				Some: &intgraphql.AttachmentFilter{
					Creator: &intgraphql.NullableUserFilter{
						ID: &intgraphql.IDComparator{
							In: userIDs,
						},
					},
				},
			}
		}
	}

	// 9. autoArchivedAt
	if after, _ := cmd.Flags().GetString("auto-archived-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid auto-archived-after date: %w", err)
		}
		if b.filter.AutoArchivedAt == nil {
			b.filter.AutoArchivedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.AutoArchivedAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("auto-archived-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid auto-archived-before date: %w", err)
		}
		if b.filter.AutoArchivedAt == nil {
			b.filter.AutoArchivedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.AutoArchivedAt.Lt = &tStr
	}

	// 10. autoClosedAt
	if after, _ := cmd.Flags().GetString("auto-closed-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid auto-closed-after date: %w", err)
		}
		if b.filter.AutoClosedAt == nil {
			b.filter.AutoClosedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.AutoClosedAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("auto-closed-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid auto-closed-before date: %w", err)
		}
		if b.filter.AutoClosedAt == nil {
			b.filter.AutoClosedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.AutoClosedAt.Lt = &tStr
	}

	// 11. canceledAt
	if after, _ := cmd.Flags().GetString("canceled-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid canceled-after date: %w", err)
		}
		if b.filter.CanceledAt == nil {
			b.filter.CanceledAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.CanceledAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("canceled-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid canceled-before date: %w", err)
		}
		if b.filter.CanceledAt == nil {
			b.filter.CanceledAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.CanceledAt.Lt = &tStr
	}

	// 12. children (has sub-issues)
	if hasChildren, _ := cmd.Flags().GetBool("has-children"); hasChildren {
		// Has at least one child issue
		b.filter.Children = &intgraphql.IssueCollectionFilter{
			Length: &intgraphql.NumberComparator{
				Gt: ptrFloat(0),
			},
		}
	}

	// 13. comments (by user)
	commentBy, _ := cmd.Flags().GetStringArray("comment-by")
	if len(commentBy) > 0 {
		userIDs := make([]string, 0, len(commentBy))
		for _, user := range commentBy {
			userID, err := b.resolver.ResolveUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to resolve comment-by user %q: %w", user, err)
			}
			userIDs = append(userIDs, userID)
		}

		// Has comments by any of these users
		// Optimize: single user uses Eq, multiple uses In
		if len(userIDs) == 1 {
			b.filter.Comments = &intgraphql.CommentCollectionFilter{
				Some: &intgraphql.CommentFilter{
					User: &intgraphql.UserFilter{
						ID: &intgraphql.IDComparator{
							Eq: &userIDs[0],
						},
					},
				},
			}
		} else {
			b.filter.Comments = &intgraphql.CommentCollectionFilter{
				Some: &intgraphql.CommentFilter{
					User: &intgraphql.UserFilter{
						ID: &intgraphql.IDComparator{
							In: userIDs,
						},
					},
				},
			}
		}
	}

	// 14. completedAt (already implemented below)

	// 15. createdAt (already implemented below)

	// 16. creator
	if creator, _ := cmd.Flags().GetString("creator"); creator != "" {
		userID, err := b.resolver.ResolveUser(ctx, creator)
		if err != nil {
			return fmt.Errorf("failed to resolve creator: %w", err)
		}
		b.filter.Creator = &intgraphql.NullableUserFilter{
			ID: &intgraphql.IDComparator{
				Eq: &userID,
			},
		}
	}

	// 17. customerCount
	if count, _ := cmd.Flags().GetInt("customer-count"); count >= 0 {
		c := float64(count)
		b.filter.CustomerCount = &intgraphql.NumberComparator{
			Eq: &c,
		}
	}

	// 18. customerImportantCount
	if count, _ := cmd.Flags().GetInt("customer-important-count"); count >= 0 {
		c := float64(count)
		b.filter.CustomerImportantCount = &intgraphql.NumberComparator{
			Eq: &c,
		}
	}

	// 19. cycle
	if cycle, _ := cmd.Flags().GetString("cycle"); cycle != "" {
		b.filter.Cycle = &intgraphql.NullableCycleFilter{
			ID: &intgraphql.IDComparator{
				Eq: &cycle,
			},
		}
	}

	// 20. cycleTime - [Internal] - skip

	// 21. delegate
	if delegate, _ := cmd.Flags().GetString("delegate"); delegate != "" {
		userID, err := b.resolver.ResolveUser(ctx, delegate)
		if err != nil {
			return fmt.Errorf("failed to resolve delegate: %w", err)
		}
		b.filter.Delegate = &intgraphql.NullableUserFilter{
			ID: &intgraphql.IDComparator{
				Eq: &userID,
			},
		}
	}

	// 22. description
	if description, _ := cmd.Flags().GetString("description"); description != "" {
		b.filter.Description = &intgraphql.NullableStringComparator{
			Contains: &description,
		}
	}

	// 23. dueDate
	if after, _ := cmd.Flags().GetString("due-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid due-after date: %w", err)
		}
		if b.filter.DueDate == nil {
			b.filter.DueDate = &intgraphql.NullableTimelessDateComparator{}
		}
		// TimelessDate format is just YYYY-MM-DD
		tStr := t.Format("2006-01-02")
		b.filter.DueDate.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("due-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid due-before date: %w", err)
		}
		if b.filter.DueDate == nil {
			b.filter.DueDate = &intgraphql.NullableTimelessDateComparator{}
		}
		tStr := t.Format("2006-01-02")
		b.filter.DueDate.Lt = &tStr
	}

	// 24. estimate
	if estimate, _ := cmd.Flags().GetInt("estimate"); estimate >= 0 {
		e := float64(estimate)
		b.filter.Estimate = &intgraphql.EstimateComparator{
			Eq: &e,
		}
	}

	// 25. hasBlockedByRelations
	if hasBlocked, _ := cmd.Flags().GetBool("has-blocked-by"); hasBlocked {
		trueVal := true
		b.filter.HasBlockedByRelations = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 26. hasBlockingRelations
	if hasBlocking, _ := cmd.Flags().GetBool("has-blocking"); hasBlocking {
		trueVal := true
		b.filter.HasBlockingRelations = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 27. hasDuplicateRelations
	if hasDupe, _ := cmd.Flags().GetBool("has-duplicate"); hasDupe {
		trueVal := true
		b.filter.HasDuplicateRelations = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 28. hasRelatedRelations
	if hasRelated, _ := cmd.Flags().GetBool("has-related"); hasRelated {
		trueVal := true
		b.filter.HasRelatedRelations = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 29. hasSuggestedAssignees
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-assignees"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedAssignees = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 30. hasSuggestedLabels
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-labels"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedLabels = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 31. hasSuggestedProjects
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-projects"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedProjects = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 32. hasSuggestedRelatedIssues - [Internal] - skip

	// 33. hasSuggestedSimilarIssues - [Internal] - skip

	// 34. hasSuggestedTeams
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-teams"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedTeams = &intgraphql.RelationExistsComparator{
			Eq: &trueVal,
		}
	}

	// 35. id
	if id, _ := cmd.Flags().GetString("id"); id != "" {
		b.filter.ID = &intgraphql.IDComparator{
			Eq: &id,
		}
	}

	// 36. labels (already implemented below)

	// 37. lastAppliedTemplate
	if template, _ := cmd.Flags().GetString("last-applied-template"); template != "" {
		b.filter.LastAppliedTemplate = &intgraphql.NullableTemplateFilter{
			ID: &intgraphql.IDComparator{
				Eq: &template,
			},
		}
	}

	// 38. leadTime - [Internal] - skip

	// 39. needs (has customer needs)
	if hasNeeds, _ := cmd.Flags().GetBool("has-needs"); hasNeeds {
		// Has at least one customer need
		b.filter.Needs = &intgraphql.CustomerNeedCollectionFilter{
			Length: &intgraphql.NumberComparator{
				Gt: ptrFloat(0),
			},
		}
	}

	// 40. number
	if number, _ := cmd.Flags().GetInt("number"); number >= 0 {
		n := float64(number)
		b.filter.Number = &intgraphql.NumberComparator{
			Eq: &n,
		}
	}

	// 41. or - compound filter [IssueFilter!] - skip (requires recursive filter building)

	// 42. parent
	if parent, _ := cmd.Flags().GetString("parent"); parent != "" {
		b.filter.Parent = &intgraphql.NullableIssueFilter{
			ID: &intgraphql.IDComparator{
				Eq: &parent,
			},
		}
	}

	// 43. priority (already implemented below)

	// 44. project
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		b.filter.Project = &intgraphql.NullableProjectFilter{
			ID: &intgraphql.IDComparator{
				Eq: &project,
			},
		}
	}

	// 45. projectMilestone
	if milestone, _ := cmd.Flags().GetString("project-milestone"); milestone != "" {
		b.filter.ProjectMilestone = &intgraphql.NullableProjectMilestoneFilter{
			ID: &intgraphql.IDComparator{
				Eq: &milestone,
			},
		}
	}

	// 46. reactions (has reactions)
	if hasReactions, _ := cmd.Flags().GetBool("has-reactions"); hasReactions {
		// Has at least one reaction
		b.filter.Reactions = &intgraphql.ReactionCollectionFilter{
			Length: &intgraphql.NumberComparator{
				Gt: ptrFloat(0),
			},
		}
	}

	// 47. recurringIssueTemplate - [ALPHA] - skip

	// 48. searchableContent - [Internal] - skip

	// 49. slaStatus - SlaStatusComparator - skip (complex SLA logic)

	// 50. snoozedBy
	if snoozedBy, _ := cmd.Flags().GetString("snoozed-by"); snoozedBy != "" {
		userID, err := b.resolver.ResolveUser(ctx, snoozedBy)
		if err != nil {
			return fmt.Errorf("failed to resolve snoozed-by: %w", err)
		}
		b.filter.SnoozedBy = &intgraphql.NullableUserFilter{
			ID: &intgraphql.IDComparator{
				Eq: &userID,
			},
		}
	}

	// 51. snoozedUntilAt
	if after, _ := cmd.Flags().GetString("snoozed-until-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid snoozed-until-after date: %w", err)
		}
		if b.filter.SnoozedUntilAt == nil {
			b.filter.SnoozedUntilAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.SnoozedUntilAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("snoozed-until-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid snoozed-until-before date: %w", err)
		}
		if b.filter.SnoozedUntilAt == nil {
			b.filter.SnoozedUntilAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.SnoozedUntilAt.Lt = &tStr
	}

	// 52. sourceMetadata - SourceMetadataComparator - skip (complex integration metadata)

	// 53. startedAt
	if after, _ := cmd.Flags().GetString("started-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid started-after date: %w", err)
		}
		if b.filter.StartedAt == nil {
			b.filter.StartedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.StartedAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("started-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid started-before date: %w", err)
		}
		if b.filter.StartedAt == nil {
			b.filter.StartedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.StartedAt.Lt = &tStr
	}

	// 54. state (already implemented below)

	// 55. subscribers
	subscribers, _ := cmd.Flags().GetStringArray("subscriber")
	if len(subscribers) > 0 {
		userIDs := make([]string, 0, len(subscribers))
		for _, user := range subscribers {
			userID, err := b.resolver.ResolveUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to resolve subscriber %q: %w", user, err)
			}
			userIDs = append(userIDs, userID)
		}

		// Has subscribers matching any of these users
		// Optimize: single user uses Eq, multiple uses In
		if len(userIDs) == 1 {
			b.filter.Subscribers = &intgraphql.UserCollectionFilter{
				Some: &intgraphql.UserFilter{
					ID: &intgraphql.IDComparator{
						Eq: &userIDs[0],
					},
				},
			}
		} else {
			b.filter.Subscribers = &intgraphql.UserCollectionFilter{
				Some: &intgraphql.UserFilter{
					ID: &intgraphql.IDComparator{
						In: userIDs,
					},
				},
			}
		}
	}

	// 56. suggestions - [Internal] - skip

	// 57. team (moved after state below for existing code order)

	// 58. title
	if title, _ := cmd.Flags().GetString("title"); title != "" {
		b.filter.Title = &intgraphql.StringComparator{
			Contains: &title,
		}
	}

	// 59. triageTime - [Internal] - skip

	// 60. triagedAt
	if after, _ := cmd.Flags().GetString("triaged-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid triaged-after date: %w", err)
		}
		if b.filter.TriagedAt == nil {
			b.filter.TriagedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.TriagedAt.Gte = &tStr
	}
	if before, _ := cmd.Flags().GetString("triaged-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid triaged-before date: %w", err)
		}
		if b.filter.TriagedAt == nil {
			b.filter.TriagedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.TriagedAt.Lt = &tStr
	}

	// 61. updatedAt (already implemented below)

	// State filter
	if state, _ := cmd.Flags().GetString("state"); state != "" {
		stateID, err := b.resolver.ResolveState(ctx, state)
		if err != nil {
			return fmt.Errorf("failed to resolve state: %w", err)
		}
		b.filter.State = &intgraphql.WorkflowStateFilter{
			ID: &intgraphql.IDComparator{
				Eq: &stateID,
			},
		}
	}

	// Priority filter
	if priority, _ := cmd.Flags().GetInt("priority"); priority >= 0 {
		p := float64(priority)
		b.filter.Priority = &intgraphql.NullableNumberComparator{
			Eq: &p,
		}
	}

	// Created date filters
	if after, _ := cmd.Flags().GetString("created-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid created-after date: %w", err)
		}
		if b.filter.CreatedAt == nil {
			b.filter.CreatedAt = &intgraphql.DateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.CreatedAt.Gte = &tStr
	}

	if before, _ := cmd.Flags().GetString("created-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid created-before date: %w", err)
		}
		if b.filter.CreatedAt == nil {
			b.filter.CreatedAt = &intgraphql.DateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.CreatedAt.Lt = &tStr
	}

	// Completed date filters
	if after, _ := cmd.Flags().GetString("completed-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid completed-after date: %w", err)
		}
		if b.filter.CompletedAt == nil {
			b.filter.CompletedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.CompletedAt.Gte = &tStr
	}

	if before, _ := cmd.Flags().GetString("completed-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid completed-before date: %w", err)
		}
		if b.filter.CompletedAt == nil {
			b.filter.CompletedAt = &intgraphql.NullableDateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.CompletedAt.Lt = &tStr
	}

	// Updated date filters
	if after, _ := cmd.Flags().GetString("updated-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid updated-after date: %w", err)
		}
		if b.filter.UpdatedAt == nil {
			b.filter.UpdatedAt = &intgraphql.DateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.UpdatedAt.Gte = &tStr
	}

	if before, _ := cmd.Flags().GetString("updated-before"); before != "" {
		t, err := b.parser.Parse(before)
		if err != nil {
			return fmt.Errorf("invalid updated-before date: %w", err)
		}
		if b.filter.UpdatedAt == nil {
			b.filter.UpdatedAt = &intgraphql.DateComparator{}
		}
		tStr := t.Format("2006-01-02T15:04:05.000Z")
		b.filter.UpdatedAt.Lt = &tStr
	}

	// Label filters
	labels, _ := cmd.Flags().GetStringArray("label")
	if len(labels) > 0 {
		labelIDs := make([]string, 0, len(labels))
		for _, label := range labels {
			labelID, err := b.resolver.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}

		// Create OR filter for labels
		if len(labelIDs) > 0 {
			filters := make([]*intgraphql.IssueLabelFilter, len(labelIDs))
			for i, id := range labelIDs {
				idCopy := id
				filters[i] = &intgraphql.IssueLabelFilter{
					ID: &intgraphql.IDComparator{
						Eq: &idCopy,
					},
				}
			}
			b.filter.Labels = &intgraphql.IssueLabelCollectionFilter{
				Some: &intgraphql.IssueLabelFilter{
					Or: filters,
				},
			}
		}
	}

	return nil
}

// Build returns the constructed filter.
// Returns nil if no filters were set.
func (b *IssueFilterBuilder) Build() *intgraphql.IssueFilter {
	// Check if any filters were actually set
	// With 40+ filters now, use reflection would be cleaner but checking common ones
	if b.filter.Team == nil &&
		b.filter.Assignee == nil &&
		b.filter.State == nil &&
		b.filter.Priority == nil &&
		b.filter.CreatedAt == nil &&
		b.filter.CompletedAt == nil &&
		b.filter.UpdatedAt == nil &&
		b.filter.Labels == nil &&
		b.filter.Creator == nil &&
		b.filter.Cycle == nil &&
		b.filter.Project == nil &&
		b.filter.Parent == nil &&
		b.filter.ArchivedAt == nil &&
		b.filter.CanceledAt == nil &&
		b.filter.StartedAt == nil &&
		b.filter.TriagedAt == nil &&
		b.filter.HasSuggestedTeams == nil &&
		b.filter.HasSuggestedAssignees == nil &&
		b.filter.HasSuggestedProjects == nil &&
		b.filter.HasSuggestedLabels == nil &&
		b.filter.Comments == nil &&
		b.filter.Attachments == nil &&
		b.filter.Subscribers == nil &&
		b.filter.Children == nil &&
		b.filter.Needs == nil &&
		b.filter.Reactions == nil &&
		b.filter.DueDate == nil &&
		b.filter.Estimate == nil &&
		b.filter.Title == nil &&
		b.filter.Description == nil &&
		b.filter.Number == nil &&
		b.filter.ID == nil &&
		b.filter.Delegate == nil &&
		b.filter.SnoozedBy == nil &&
		b.filter.SnoozedUntilAt == nil &&
		b.filter.AutoArchivedAt == nil &&
		b.filter.AutoClosedAt == nil &&
		b.filter.AddedToCycleAt == nil &&
		b.filter.AddedToCyclePeriod == nil &&
		b.filter.ProjectMilestone == nil &&
		b.filter.LastAppliedTemplate == nil &&
		b.filter.CustomerCount == nil &&
		b.filter.CustomerImportantCount == nil &&
		b.filter.HasBlockedByRelations == nil &&
		b.filter.HasBlockingRelations == nil &&
		b.filter.HasDuplicateRelations == nil &&
		b.filter.HasRelatedRelations == nil {
		return nil
	}

	return b.filter
}

// ptrFloat returns a pointer to a float64
func ptrFloat(f float64) *float64 {
	return &f
}
