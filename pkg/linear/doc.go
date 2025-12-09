//go:generate go run github.com/Yamashou/gqlgenc

// Package linear provides a Go client library for the Linear API.
//
// The Linear API is a GraphQL API for managing issues, projects, teams,
// and other resources in Linear (https://linear.app).
//
// # Getting Your API Key
//
// Before using this SDK, you need a Linear API key.
//
// Get your API key: https://linear.app/settings/account/security
//
// The key format is: lin_api_xxx (starts with "lin_api_")
// OAuth tokens are also supported and work automatically.
//
// # Quick Start
//
// Create a client with your API key:
//
//	client, err := linear.NewClient("lin_api_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Verify authentication:
//
//	viewer, err := client.Viewer(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Authenticated as: %s\n", viewer.Email)
//
// # API Structure
//
// The SDK provides two types of operations:
//
//	Queries (Read):  Viewer, Issue, Issues, Team, Teams, Project, Projects, etc.
//	Mutations (Write): IssueCreate, IssueUpdate, IssueDelete, CommentCreate, etc.
//
// Single resource methods (Issue, Team, Project) return one item by ID.
// List methods (Issues, Teams, Projects) return paginated collections.
//
// # Pagination
//
// List methods use cursor-based pagination:
//
//	first := int64(50)
//	issues, err := client.Issues(ctx, &first, nil) // First page
//	if issues.PageInfo.HasNextPage {
//	    // Get next page
//	    issues, err = client.Issues(ctx, &first, issues.PageInfo.EndCursor)
//	}
//
// Or use iterators for automatic pagination:
//
//	iter := NewIssueIterator(client, 100)
//	for iter.Next(ctx) {
//	    issue := iter.Issue()
//	    fmt.Println(issue.Title)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
//
// # Mutations
//
// Mutations modify data and require Write permission:
//
//	title := "Fix login bug"
//	issue, err := client.IssueCreate(ctx, IssueCreateInput{
//	    TeamID: "team-uuid",
//	    Title: &title,
//	})
//
// All input fields (except required IDs) are pointers.
// Nil means omit the field (for Create) or leave unchanged (for Update).
//
// # Getting Required IDs
//
// Mutations require resource UUIDs. Get them from query methods:
//
//	TeamID:
//	  teams, _ := client.Teams(ctx, nil, nil)
//	  teamID := teams.Nodes[0].ID
//
//	UserID:
//	  viewer, _ := client.Viewer(ctx)
//	  userID := viewer.ID
//
//	StateID:
//	  states, _ := client.WorkflowStates(ctx, nil, nil)
//	  stateID := states.Nodes[0].ID  // Find by state.Type or state.Name
//
//	LabelID:
//	  labels, _ := client.IssueLabels(ctx, nil, nil)
//	  labelID := labels.Nodes[0].ID  // Find by label.Name
//
// # Permissions
//
// Linear API uses permission-based access:
//
//	Read:  Query methods (Viewer, Issues, Teams, etc.)
//	Write: Mutation methods (IssueCreate, CommentCreate, etc.)
//
// API keys can be scoped to specific permissions at
// https://linear.app/settings/account/security
//
// # Error Handling
//
// All methods return errors for:
//   - Authentication failures (401)
//   - Permission denied (403)
//   - Resource not found (404)
//   - Rate limiting (429)
//   - Network errors
//
// Mutations check a Success boolean. If Success is false, an error
// is returned automatically - you only need to check err != nil.
//
// # Testing
//
// The SDK provides build tags for different test types:
//
//	make test             # Mock tests (httptest.Server, no API key)
//	make test-read        # Live read-only tests (needs LINEAR_API_KEY)
//	make test-write       # Live mutations (needs LINEAR_API_KEY, modifies test-server)
//
// See README.md for more details.
package linear
