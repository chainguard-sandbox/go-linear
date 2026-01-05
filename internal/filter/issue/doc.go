// Package issue provides filter implementations for Linear issues.
//
// This package implements 44 filters following the per-filter file pattern:
//   - Date filters: created-at, updated-at, completed-at, started-at, archived-at, etc.
//   - Entity filters: team, assignee, creator, cycle, project, parent, labels, etc.
//   - Boolean filters: has-blocked-by, has-blocking, has-suggested-teams, etc.
//   - Text filters: title, description
//   - Numeric filters: priority, estimate, number, customer-count, etc.
//
// Each filter is a standalone function in its own file, registered in registry.go.
// The IssueFilterBuilder in builder.go constructs the GraphQL filter.
//
// Usage:
//
//	builder := issue.NewIssueFilterBuilder(resolver)
//	if err := builder.FromFlags(ctx, cmd); err != nil {
//	    return err
//	}
//	filter := builder.Build()
package issue
