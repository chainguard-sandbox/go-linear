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
// Each filter is handled by its own method for maintainability.
func (b *IssueFilterBuilder) FromFlags(ctx context.Context, cmd *cobra.Command) error {
	// Apply each filter in alphabetical order
	filters := []func(context.Context, *cobra.Command) error{
		b.applyAddedToCycleAt,
		b.applyAddedToCyclePeriod,
		b.applyArchivedAt,
		b.applyAssignee,
		b.applyAttachments,
		b.applyAutoArchivedAt,
		b.applyAutoClosedAt,
		b.applyCanceledAt,
		b.applyChildren,
		b.applyComments,
		b.applyCompletedAt,
		b.applyCreatedAt,
		b.applyCreator,
		b.applyCustomerCount,
		b.applyCustomerImportantCount,
		b.applyCycle,
		b.applyDelegate,
		b.applyDescription,
		b.applyDueDate,
		b.applyEstimate,
		b.applyHasBlockedByRelations,
		b.applyHasBlockingRelations,
		b.applyHasDuplicateRelations,
		b.applyHasRelatedRelations,
		b.applyHasSuggestedAssignees,
		b.applyHasSuggestedLabels,
		b.applyHasSuggestedProjects,
		b.applyHasSuggestedTeams,
		b.applyID,
		b.applyLabels,
		b.applyLastAppliedTemplate,
		b.applyNeeds,
		b.applyNumber,
		b.applyParent,
		b.applyPriority,
		b.applyProject,
		b.applyProjectMilestone,
		b.applyReactions,
		b.applySLAStatus,
		b.applySnoozedBy,
		b.applySnoozedUntilAt,
		b.applyStartedAt,
		b.applyState,
		b.applySubscribers,
		b.applyTitle,
		b.applyTriagedAt,
		b.applyUpdatedAt,
	}

	for _, apply := range filters {
		if err := apply(ctx, cmd); err != nil {
			return err
		}
	}

	return nil
}

// applyAddedToCycleAt handles --added-to-cycle-after and --added-to-cycle-before flags.
func (b *IssueFilterBuilder) applyAddedToCycleAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyAddedToCyclePeriod handles --added-to-cycle-period flag.
func (b *IssueFilterBuilder) applyAddedToCyclePeriod(_ context.Context, cmd *cobra.Command) error {
	if period, _ := cmd.Flags().GetString("added-to-cycle-period"); period != "" {
		b.filter.AddedToCyclePeriod = &intgraphql.CyclePeriodComparator{
			Eq: (*intgraphql.CyclePeriod)(&period),
		}
	}
	return nil
}

// applyArchivedAt handles --archived-after and --archived-before flags.
func (b *IssueFilterBuilder) applyArchivedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyAssignee handles --assignee flag.
func (b *IssueFilterBuilder) applyAssignee(ctx context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyAttachments handles --attachment-by and --attachment-source-type flags.
func (b *IssueFilterBuilder) applyAttachments(ctx context.Context, cmd *cobra.Command) error {
	attachmentBy, _ := cmd.Flags().GetStringArray("attachment-by")
	attachmentSource, _ := cmd.Flags().GetString("attachment-source-type")

	if len(attachmentBy) == 0 && attachmentSource == "" {
		return nil
	}

	attachmentFilter := &intgraphql.AttachmentFilter{}

	if len(attachmentBy) > 0 {
		userIDs := make([]string, 0, len(attachmentBy))
		for _, user := range attachmentBy {
			userID, err := b.resolver.ResolveUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to resolve attachment-by user %q: %w", user, err)
			}
			userIDs = append(userIDs, userID)
		}

		if len(userIDs) == 1 {
			attachmentFilter.Creator = &intgraphql.NullableUserFilter{
				ID: &intgraphql.IDComparator{Eq: &userIDs[0]},
			}
		} else {
			attachmentFilter.Creator = &intgraphql.NullableUserFilter{
				ID: &intgraphql.IDComparator{In: userIDs},
			}
		}
	}

	if attachmentSource != "" {
		attachmentFilter.SourceType = &intgraphql.SourceTypeComparator{
			Contains: &attachmentSource,
		}
	}

	b.filter.Attachments = &intgraphql.AttachmentCollectionFilter{
		Some: attachmentFilter,
	}
	return nil
}

// applyAutoArchivedAt handles --auto-archived-after and --auto-archived-before flags.
func (b *IssueFilterBuilder) applyAutoArchivedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyAutoClosedAt handles --auto-closed-after and --auto-closed-before flags.
func (b *IssueFilterBuilder) applyAutoClosedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyCanceledAt handles --canceled-after and --canceled-before flags.
func (b *IssueFilterBuilder) applyCanceledAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyChildren handles --has-children flag.
func (b *IssueFilterBuilder) applyChildren(_ context.Context, cmd *cobra.Command) error {
	if hasChildren, _ := cmd.Flags().GetBool("has-children"); hasChildren {
		b.filter.Children = &intgraphql.IssueCollectionFilter{
			Length: &intgraphql.NumberComparator{
				Gt: ptrFloat(0),
			},
		}
	}
	return nil
}

// applyComments handles --comment-by and --comment-contains flags.
func (b *IssueFilterBuilder) applyComments(ctx context.Context, cmd *cobra.Command) error {
	commentBy, _ := cmd.Flags().GetStringArray("comment-by")
	commentContains, _ := cmd.Flags().GetString("comment-contains")

	if len(commentBy) == 0 && commentContains == "" {
		return nil
	}

	commentFilter := &intgraphql.CommentFilter{}

	if len(commentBy) > 0 {
		userIDs := make([]string, 0, len(commentBy))
		for _, user := range commentBy {
			userID, err := b.resolver.ResolveUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to resolve comment-by user %q: %w", user, err)
			}
			userIDs = append(userIDs, userID)
		}

		if len(userIDs) == 1 {
			commentFilter.User = &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{Eq: &userIDs[0]},
			}
		} else {
			commentFilter.User = &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{In: userIDs},
			}
		}
	}

	if commentContains != "" {
		commentFilter.Body = &intgraphql.StringComparator{
			Contains: &commentContains,
		}
	}

	b.filter.Comments = &intgraphql.CommentCollectionFilter{
		Some: commentFilter,
	}
	return nil
}

// applyCompletedAt handles --completed-after and --completed-before flags.
func (b *IssueFilterBuilder) applyCompletedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyCreatedAt handles --created-after and --created-before flags.
func (b *IssueFilterBuilder) applyCreatedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyCreator handles --creator flag.
func (b *IssueFilterBuilder) applyCreator(ctx context.Context, cmd *cobra.Command) error {
	if creator, _ := cmd.Flags().GetString("creator"); creator != "" {
		userID, err := b.resolver.ResolveUser(ctx, creator)
		if err != nil {
			return fmt.Errorf("failed to resolve creator: %w", err)
		}
		b.filter.Creator = &intgraphql.NullableUserFilter{
			ID: &intgraphql.IDComparator{Eq: &userID},
		}
	}
	return nil
}

// applyCustomerCount handles --customer-count flag.
func (b *IssueFilterBuilder) applyCustomerCount(_ context.Context, cmd *cobra.Command) error {
	if count, _ := cmd.Flags().GetInt("customer-count"); count >= 0 {
		c := float64(count)
		b.filter.CustomerCount = &intgraphql.NumberComparator{Eq: &c}
	}
	return nil
}

// applyCustomerImportantCount handles --customer-important-count flag.
func (b *IssueFilterBuilder) applyCustomerImportantCount(_ context.Context, cmd *cobra.Command) error {
	if count, _ := cmd.Flags().GetInt("customer-important-count"); count >= 0 {
		c := float64(count)
		b.filter.CustomerImportantCount = &intgraphql.NumberComparator{Eq: &c}
	}
	return nil
}

// applyCycle handles --cycle flag.
func (b *IssueFilterBuilder) applyCycle(_ context.Context, cmd *cobra.Command) error {
	if cycle, _ := cmd.Flags().GetString("cycle"); cycle != "" {
		b.filter.Cycle = &intgraphql.NullableCycleFilter{
			ID: &intgraphql.IDComparator{Eq: &cycle},
		}
	}
	return nil
}

// applyDelegate handles --delegate flag.
func (b *IssueFilterBuilder) applyDelegate(ctx context.Context, cmd *cobra.Command) error {
	if delegate, _ := cmd.Flags().GetString("delegate"); delegate != "" {
		userID, err := b.resolver.ResolveUser(ctx, delegate)
		if err != nil {
			return fmt.Errorf("failed to resolve delegate: %w", err)
		}
		b.filter.Delegate = &intgraphql.NullableUserFilter{
			ID: &intgraphql.IDComparator{Eq: &userID},
		}
	}
	return nil
}

// applyDescription handles --description flag.
func (b *IssueFilterBuilder) applyDescription(_ context.Context, cmd *cobra.Command) error {
	if description, _ := cmd.Flags().GetString("description"); description != "" {
		b.filter.Description = &intgraphql.NullableStringComparator{
			Contains: &description,
		}
	}
	return nil
}

// applyDueDate handles --due-after and --due-before flags.
func (b *IssueFilterBuilder) applyDueDate(_ context.Context, cmd *cobra.Command) error {
	if after, _ := cmd.Flags().GetString("due-after"); after != "" {
		t, err := b.parser.Parse(after)
		if err != nil {
			return fmt.Errorf("invalid due-after date: %w", err)
		}
		if b.filter.DueDate == nil {
			b.filter.DueDate = &intgraphql.NullableTimelessDateComparator{}
		}
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
	return nil
}

// applyEstimate handles --estimate flag.
func (b *IssueFilterBuilder) applyEstimate(_ context.Context, cmd *cobra.Command) error {
	if estimate, _ := cmd.Flags().GetInt("estimate"); estimate >= 0 {
		e := float64(estimate)
		b.filter.Estimate = &intgraphql.EstimateComparator{Eq: &e}
	}
	return nil
}

// applyHasBlockedByRelations handles --has-blocked-by flag.
func (b *IssueFilterBuilder) applyHasBlockedByRelations(_ context.Context, cmd *cobra.Command) error {
	if hasBlocked, _ := cmd.Flags().GetBool("has-blocked-by"); hasBlocked {
		trueVal := true
		b.filter.HasBlockedByRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyHasBlockingRelations handles --has-blocking flag.
func (b *IssueFilterBuilder) applyHasBlockingRelations(_ context.Context, cmd *cobra.Command) error {
	if hasBlocking, _ := cmd.Flags().GetBool("has-blocking"); hasBlocking {
		trueVal := true
		b.filter.HasBlockingRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyHasDuplicateRelations handles --has-duplicate flag.
func (b *IssueFilterBuilder) applyHasDuplicateRelations(_ context.Context, cmd *cobra.Command) error {
	if hasDupe, _ := cmd.Flags().GetBool("has-duplicate"); hasDupe {
		trueVal := true
		b.filter.HasDuplicateRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyHasRelatedRelations handles --has-related flag.
func (b *IssueFilterBuilder) applyHasRelatedRelations(_ context.Context, cmd *cobra.Command) error {
	if hasRelated, _ := cmd.Flags().GetBool("has-related"); hasRelated {
		trueVal := true
		b.filter.HasRelatedRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyHasSuggestedAssignees handles --has-suggested-assignees flag.
func (b *IssueFilterBuilder) applyHasSuggestedAssignees(_ context.Context, cmd *cobra.Command) error {
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-assignees"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedAssignees = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyHasSuggestedLabels handles --has-suggested-labels flag.
func (b *IssueFilterBuilder) applyHasSuggestedLabels(_ context.Context, cmd *cobra.Command) error {
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-labels"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedLabels = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyHasSuggestedProjects handles --has-suggested-projects flag.
func (b *IssueFilterBuilder) applyHasSuggestedProjects(_ context.Context, cmd *cobra.Command) error {
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-projects"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedProjects = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyHasSuggestedTeams handles --has-suggested-teams flag.
func (b *IssueFilterBuilder) applyHasSuggestedTeams(_ context.Context, cmd *cobra.Command) error {
	if hasSuggested, _ := cmd.Flags().GetBool("has-suggested-teams"); hasSuggested {
		trueVal := true
		b.filter.HasSuggestedTeams = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	}
	return nil
}

// applyID handles --id flag.
func (b *IssueFilterBuilder) applyID(_ context.Context, cmd *cobra.Command) error {
	if id, _ := cmd.Flags().GetString("id"); id != "" {
		b.filter.ID = &intgraphql.IDComparator{Eq: &id}
	}
	return nil
}

// applyLabels handles --label flag.
func (b *IssueFilterBuilder) applyLabels(ctx context.Context, cmd *cobra.Command) error {
	labels, _ := cmd.Flags().GetStringArray("label")
	if len(labels) == 0 {
		return nil
	}

	labelIDs := make([]string, 0, len(labels))
	for _, label := range labels {
		labelID, err := b.resolver.ResolveLabel(ctx, label)
		if err != nil {
			return fmt.Errorf("failed to resolve label %q: %w", label, err)
		}
		labelIDs = append(labelIDs, labelID)
	}

	if len(labelIDs) > 0 {
		filters := make([]*intgraphql.IssueLabelFilter, len(labelIDs))
		for i, id := range labelIDs {
			idCopy := id
			filters[i] = &intgraphql.IssueLabelFilter{
				ID: &intgraphql.IDComparator{Eq: &idCopy},
			}
		}
		b.filter.Labels = &intgraphql.IssueLabelCollectionFilter{
			Some: &intgraphql.IssueLabelFilter{Or: filters},
		}
	}
	return nil
}

// applyLastAppliedTemplate handles --last-applied-template flag.
func (b *IssueFilterBuilder) applyLastAppliedTemplate(_ context.Context, cmd *cobra.Command) error {
	if template, _ := cmd.Flags().GetString("last-applied-template"); template != "" {
		b.filter.LastAppliedTemplate = &intgraphql.NullableTemplateFilter{
			ID: &intgraphql.IDComparator{Eq: &template},
		}
	}
	return nil
}

// applyNeeds handles --has-needs flag.
func (b *IssueFilterBuilder) applyNeeds(_ context.Context, cmd *cobra.Command) error {
	if hasNeeds, _ := cmd.Flags().GetBool("has-needs"); hasNeeds {
		b.filter.Needs = &intgraphql.CustomerNeedCollectionFilter{
			Length: &intgraphql.NumberComparator{Gt: ptrFloat(0)},
		}
	}
	return nil
}

// applyNumber handles --number flag.
func (b *IssueFilterBuilder) applyNumber(_ context.Context, cmd *cobra.Command) error {
	if number, _ := cmd.Flags().GetInt("number"); number >= 0 {
		n := float64(number)
		b.filter.Number = &intgraphql.NumberComparator{Eq: &n}
	}
	return nil
}

// applyParent handles --parent flag.
func (b *IssueFilterBuilder) applyParent(_ context.Context, cmd *cobra.Command) error {
	if parent, _ := cmd.Flags().GetString("parent"); parent != "" {
		b.filter.Parent = &intgraphql.NullableIssueFilter{
			ID: &intgraphql.IDComparator{Eq: &parent},
		}
	}
	return nil
}

// applyPriority handles --priority flag.
func (b *IssueFilterBuilder) applyPriority(_ context.Context, cmd *cobra.Command) error {
	if priority, _ := cmd.Flags().GetInt("priority"); priority >= 0 {
		p := float64(priority)
		b.filter.Priority = &intgraphql.NullableNumberComparator{Eq: &p}
	}
	return nil
}

// applyProject handles --project flag.
func (b *IssueFilterBuilder) applyProject(_ context.Context, cmd *cobra.Command) error {
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		b.filter.Project = &intgraphql.NullableProjectFilter{
			ID: &intgraphql.IDComparator{Eq: &project},
		}
	}
	return nil
}

// applyProjectMilestone handles --project-milestone flag.
func (b *IssueFilterBuilder) applyProjectMilestone(_ context.Context, cmd *cobra.Command) error {
	if milestone, _ := cmd.Flags().GetString("project-milestone"); milestone != "" {
		b.filter.ProjectMilestone = &intgraphql.NullableProjectMilestoneFilter{
			ID: &intgraphql.IDComparator{Eq: &milestone},
		}
	}
	return nil
}

// applyReactions handles --has-reactions flag.
func (b *IssueFilterBuilder) applyReactions(_ context.Context, cmd *cobra.Command) error {
	if hasReactions, _ := cmd.Flags().GetBool("has-reactions"); hasReactions {
		b.filter.Reactions = &intgraphql.ReactionCollectionFilter{
			Length: &intgraphql.NumberComparator{Gt: ptrFloat(0)},
		}
	}
	return nil
}

// applySLAStatus handles --sla-status flag.
func (b *IssueFilterBuilder) applySLAStatus(_ context.Context, cmd *cobra.Command) error {
	if slaStatus, _ := cmd.Flags().GetString("sla-status"); slaStatus != "" {
		b.filter.SLAStatus = &intgraphql.SLAStatusComparator{
			Eq: (*intgraphql.SLAStatus)(&slaStatus),
		}
	}
	return nil
}

// applySnoozedBy handles --snoozed-by flag.
func (b *IssueFilterBuilder) applySnoozedBy(ctx context.Context, cmd *cobra.Command) error {
	if snoozedBy, _ := cmd.Flags().GetString("snoozed-by"); snoozedBy != "" {
		userID, err := b.resolver.ResolveUser(ctx, snoozedBy)
		if err != nil {
			return fmt.Errorf("failed to resolve snoozed-by: %w", err)
		}
		b.filter.SnoozedBy = &intgraphql.NullableUserFilter{
			ID: &intgraphql.IDComparator{Eq: &userID},
		}
	}
	return nil
}

// applySnoozedUntilAt handles --snoozed-until-after and --snoozed-until-before flags.
func (b *IssueFilterBuilder) applySnoozedUntilAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyStartedAt handles --started-after and --started-before flags.
func (b *IssueFilterBuilder) applyStartedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyState handles --state flag.
func (b *IssueFilterBuilder) applyState(ctx context.Context, cmd *cobra.Command) error {
	if state, _ := cmd.Flags().GetString("state"); state != "" {
		stateID, err := b.resolver.ResolveState(ctx, state)
		if err != nil {
			return fmt.Errorf("failed to resolve state: %w", err)
		}
		b.filter.State = &intgraphql.WorkflowStateFilter{
			ID: &intgraphql.IDComparator{Eq: &stateID},
		}
	}
	return nil
}

// applySubscribers handles --subscriber flag.
func (b *IssueFilterBuilder) applySubscribers(ctx context.Context, cmd *cobra.Command) error {
	subscribers, _ := cmd.Flags().GetStringArray("subscriber")
	if len(subscribers) == 0 {
		return nil
	}

	userIDs := make([]string, 0, len(subscribers))
	for _, user := range subscribers {
		userID, err := b.resolver.ResolveUser(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to resolve subscriber %q: %w", user, err)
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 1 {
		b.filter.Subscribers = &intgraphql.UserCollectionFilter{
			Some: &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{Eq: &userIDs[0]},
			},
		}
	} else {
		b.filter.Subscribers = &intgraphql.UserCollectionFilter{
			Some: &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{In: userIDs},
			},
		}
	}
	return nil
}

// applyTitle handles --title flag.
func (b *IssueFilterBuilder) applyTitle(_ context.Context, cmd *cobra.Command) error {
	if title, _ := cmd.Flags().GetString("title"); title != "" {
		b.filter.Title = &intgraphql.StringComparator{Contains: &title}
	}
	return nil
}

// applyTriagedAt handles --triaged-after and --triaged-before flags.
func (b *IssueFilterBuilder) applyTriagedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// applyUpdatedAt handles --updated-after and --updated-before flags.
func (b *IssueFilterBuilder) applyUpdatedAt(_ context.Context, cmd *cobra.Command) error {
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
	return nil
}

// Build returns the constructed filter.
// Returns nil if no filters were set.
func (b *IssueFilterBuilder) Build() *intgraphql.IssueFilter {
	// Check if any filters were actually set
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
		b.filter.HasRelatedRelations == nil &&
		b.filter.SLAStatus == nil {
		return nil
	}

	return b.filter
}

// ptrFloat returns a pointer to a float64.
func ptrFloat(f float64) *float64 {
	return &f
}
