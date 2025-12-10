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
func (b *IssueFilterBuilder) FromFlags(ctx context.Context, cmd *cobra.Command) error {
	// Team filter
	if team, _ := cmd.Flags().GetString("team"); team != "" {
		teamID, err := b.resolver.ResolveTeam(ctx, team)
		if err != nil {
			return fmt.Errorf("failed to resolve team: %w", err)
		}
		b.filter.Team = &intgraphql.TeamFilter{
			ID: &intgraphql.IDComparator{
				Eq: &teamID,
			},
		}
	}

	// Assignee filter
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
	if b.filter.Team == nil &&
		b.filter.Assignee == nil &&
		b.filter.State == nil &&
		b.filter.Priority == nil &&
		b.filter.CreatedAt == nil &&
		b.filter.CompletedAt == nil &&
		b.filter.UpdatedAt == nil &&
		b.filter.Labels == nil {
		return nil
	}

	return b.filter
}
