package linear

import intgraphql "github.com/eslerm/go-linear/internal/graphql"

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
//   - Priority: Priority level 0-4 (*int, 0=none, 1=urgent, 2=high, 3=normal, 4=low)
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
//   - Priority: New priority 0-4 (*int)
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
