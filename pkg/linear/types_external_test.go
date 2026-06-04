package linear_test

import (
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// TestFilterTypesUsableExternally is a regression for #85: the server-side
// filter and result types for IssuesFiltered / SearchIssues / ProjectsFiltered
// must be both nameable AND populatable purely through the public pkg/linear,
// since external consumers cannot import internal/graphql.
//
// This is a compile-time check — it fails to build (not at runtime) if any of
// the re-exported aliases are removed.
func TestFilterTypesUsableExternally(t *testing.T) {
	projectID := "proj_123"
	stateType := "started"
	assignedToMe := true
	var minPriority float64 = 2

	// A consumer can build a *populated* issue filter — project + state +
	// assignee + priority — without touching internal/graphql.
	filter := linear.IssueFilter{
		Project:  &linear.NullableProjectFilter{ID: &linear.IDComparator{Eq: &projectID}},
		State:    &linear.WorkflowStateFilter{Type: &linear.StringComparator{Eq: &stateType}},
		Assignee: &linear.NullableUserFilter{IsMe: &linear.BooleanComparator{Eq: &assignedToMe}},
		Priority: &linear.NullableNumberComparator{Gte: &minPriority},
	}
	// Compound And/Or recursion is usable through the public type too.
	filter.And = []*linear.IssueFilter{{ID: &linear.IssueIDComparator{Eq: &projectID}}}
	_ = filter

	// ProjectFilter is likewise populatable (filter projects by name).
	name := "Platform"
	pf := linear.ProjectFilter{Name: &linear.StringComparator{Contains: &name}}
	_ = pf

	// Paginated result types remain nameable through the public package.
	var (
		_ *linear.ListIssuesFiltered_Issues
		_ *linear.SearchIssues_SearchIssues
		_ *linear.ListProjectsFiltered_Projects
	)
}
