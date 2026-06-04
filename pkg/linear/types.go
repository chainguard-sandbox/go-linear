package linear

import intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"

// Priority represents Linear issue priority levels.
//
// These constants provide type-safe, self-documenting priority values.
// The underlying type is int64 to match Linear's API.
//
// Example:
//
//	priority := int64(linear.PriorityUrgent)
//	issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
//	    TeamID:   teamID,
//	    Title:    &title,
//	    Priority: &priority,
//	})
type Priority int64

// Priority constants for Linear issues.
const (
	PriorityNone   Priority = iota // No priority
	PriorityUrgent                 // Urgent
	PriorityHigh                   // High
	PriorityNormal                 // Normal
	PriorityLow                    // Low
)

// Input types re-exported from internal/graphql for discoverability via go doc.
//
// These types are used as parameters to mutation methods. Re-exporting them
// makes them accessible to AI agents and developers using go doc, while the
// actual implementation remains in the generated internal/graphql package.

// IssueCreateInput defines fields for creating a new issue.
//
// Required Fields:
//   - TeamID: Team to create issue in (string, NOT pointer)
//
// Common Optional Fields (pointers, nil = omitted from request):
//   - Title: Issue title (*string)
//   - Description: Issue description in markdown (*string)
//   - Priority: Priority level 0-4 (*int64, 0=none, 1=urgent, 2=high, 3=normal, 4=low)
//   - AssigneeID: User to assign (*string, get from Users() or Viewer())
//   - StateID: Workflow state (*string, get from WorkflowStates())
//   - CycleID: Development cycle/sprint (*string, get from Cycles())
//   - LabelIds: Labels to apply ([]string, get from IssueLabels())
//   - ParentID: Parent issue for sub-issues (*string)
//   - DueDate: Due date (*string, format: YYYY-MM-DD)
//   - ProjectID: Project association (*string, get from Projects())
//
// See internal/graphql/models.go for complete field list (30+ fields available).
type IssueCreateInput = intgraphql.IssueCreateInput

// IssueUpdateInput defines fields for updating an existing issue.
//
// All fields are optional pointers. Nil means leave the field unchanged.
//
// Common Fields (pointers, nil = unchanged):
//   - Title: New title (*string)
//   - Description: New description (*string)
//   - Priority: New priority 0-4 (*int64)
//   - AssigneeID: New assignee (*string, empty string "" to unassign)
//   - StateID: New workflow state (*string)
//   - CycleID: Move to cycle (*string)
//   - LabelIds: Replace all labels ([]string)
//   - AddedLabelIds: Add these labels ([]string)
//   - RemovedLabelIds: Remove these labels ([]string)
//   - ParentID: Change parent (*string)
//
// See internal/graphql/models.go for complete field list.
type IssueUpdateInput = intgraphql.IssueUpdateInput

// CommentCreateInput defines fields for creating a comment on an issue.
//
// Required Fields:
//   - IssueID: Issue to comment on (*string)
//   - Body: Comment text in markdown (*string)
//
// Optional Fields:
//   - CreateAsUser: Create as specific user name (*string, OAuth only)
//   - DisplayIconURL: Custom avatar URL (*string, OAuth only)
//
// See internal/graphql/models.go for complete field list.
type CommentCreateInput = intgraphql.CommentCreateInput

// CommentUpdateInput defines fields for updating a comment.
//
// Common Fields (all optional, nil = unchanged):
//   - Body: New comment text (*string)
//
// See internal/graphql/models.go for complete field list.
type CommentUpdateInput = intgraphql.CommentUpdateInput

// IssueLabelCreateInput defines fields for creating an issue label.
//
// Required Fields:
//   - Name: Label name (string)
//   - Color: Label color hex code (string, format: #RRGGBB)
//
// Optional Fields:
//   - Description: Label description (*string)
//   - TeamID: Scope label to specific team (*string)
//
// See internal/graphql/models.go for complete field list.
type IssueLabelCreateInput = intgraphql.IssueLabelCreateInput

// IssueLabelUpdateInput defines fields for updating a label.
//
// All fields optional (nil = unchanged):
//   - Name: New label name (*string)
//   - Description: New description (*string)
//   - Color: New color hex (*string)
//
// See internal/graphql/models.go for complete field list.
type IssueLabelUpdateInput = intgraphql.IssueLabelUpdateInput

// TeamCreateInput defines fields for creating a new team.
//
// Required Fields:
//   - Name: Team name (string)
//   - Key: Team identifier/key (string, used in URLs)
//
// Optional Fields:
//   - Description: Team description (*string)
//   - Icon: Team icon emoji (*string)
//   - Color: Team color hex code (*string)
//   - Private: Whether team is private (*bool)
//
// See internal/graphql/models.go for complete field list.
type TeamCreateInput = intgraphql.TeamCreateInput

// TeamUpdateInput defines fields for updating a team.
//
// All fields optional (nil = unchanged):
//   - Name: New team name (*string)
//   - Description: New description (*string)
//   - Icon: New icon (*string)
//   - Color: New color (*string)
//
// See internal/graphql/models.go for complete field list.
type TeamUpdateInput = intgraphql.TeamUpdateInput

// ProjectCreateInput defines fields for creating a new project.
//
// Required Fields:
//   - Name: Project name (string)
//
// Optional Fields:
//   - Description: Project description (*string)
//   - State: Project state (*string)
//   - TargetDate: Target completion date (*string, format: YYYY-MM-DD)
//   - LeadID: Project lead user ID (*string)
//   - TeamIds: Associated teams ([]string)
//
// See internal/graphql/models.go for complete field list.
type ProjectCreateInput = intgraphql.ProjectCreateInput

// ProjectUpdateInput defines fields for updating a project.
//
// All fields optional (nil = unchanged):
//   - Name: New name (*string)
//   - Description: New description (*string)
//   - State: New state (*string)
//   - TargetDate: New target date (*string)
//
// See internal/graphql/models.go for complete field list.
type ProjectUpdateInput = intgraphql.ProjectUpdateInput

// Server-side filter and result types re-exported from internal/graphql.
//
// IssuesFiltered, SearchIssues, and ProjectsFiltered take these filter types
// and return these result types. Because internal/graphql cannot be imported
// from outside the module, external consumers could see the method signatures
// but never construct the arguments or use the returns. Re-exporting via type
// aliases closes that gap without changing the generated implementation (#85).

// IssueFilter is the server-side filter for IssuesFiltered and SearchIssues.
// See internal/graphql/models.go for the complete field list (And/Or plus
// per-field comparators).
type IssueFilter = intgraphql.IssueFilter

// ProjectFilter is the server-side filter for ProjectsFiltered.
// See internal/graphql/models.go for the complete field list.
type ProjectFilter = intgraphql.ProjectFilter

// Result-type alias names intentionally mirror the generated gqlgenc type
// names so go doc shows the same identifier for the method return and the
// alias; ST1003 (underscores) is suppressed for that reason.

// ListIssuesFiltered_Issues is the paginated result of IssuesFiltered
// (Nodes + PageInfo).
type ListIssuesFiltered_Issues = intgraphql.ListIssuesFiltered_Issues //nolint:staticcheck // ST1003: mirrors generated gqlgenc result type name

// SearchIssues_SearchIssues is the paginated result of SearchIssues
// (Nodes + PageInfo + TotalCount).
type SearchIssues_SearchIssues = intgraphql.SearchIssues_SearchIssues //nolint:staticcheck // ST1003: mirrors generated gqlgenc result type name

// ListProjectsFiltered_Projects is the paginated result of ProjectsFiltered
// (Nodes + PageInfo).
type ListProjectsFiltered_Projects = intgraphql.ListProjectsFiltered_Projects //nolint:staticcheck // ST1003: mirrors generated gqlgenc result type name

// Filter leaf types — the nested-filter and comparator types used by the fields
// of IssueFilter and ProjectFilter. Without these, the filter aliases above are
// only nameable, not populatable: a consumer could declare an IssueFilter but
// not set a single condition, because every field is typed with an internal
// type. This is the common subset covering project / state / assignee / priority
// / id / date / string filtering. The full transitive set is large and tracks
// the generated models, so it will be emitted automatically alongside the
// gqlgenc generation in a follow-up (see #85); these hand-written aliases cover
// the frequent cases until then.
//
// The first group are nested filters reachable from common IssueFilter /
// ProjectFilter fields; the second are terminal scalar comparators whose own
// fields are plain scalars/slices.
type (
	WorkflowStateFilter   = intgraphql.WorkflowStateFilter   // IssueFilter.State
	NullableProjectFilter = intgraphql.NullableProjectFilter // IssueFilter.Project
	NullableUserFilter    = intgraphql.NullableUserFilter    // IssueFilter.Assignee / Creator / ...

	IDComparator             = intgraphql.IDComparator
	IssueIDComparator        = intgraphql.IssueIDComparator
	NumberComparator         = intgraphql.NumberComparator
	NullableNumberComparator = intgraphql.NullableNumberComparator
	StringComparator         = intgraphql.StringComparator
	NullableStringComparator = intgraphql.NullableStringComparator
	DateComparator           = intgraphql.DateComparator
	NullableDateComparator   = intgraphql.NullableDateComparator
	BooleanComparator        = intgraphql.BooleanComparator
)
