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
// Use errors.As to check for specific error types:
//
//	var rateLimitErr *linear.RateLimitError
//	if errors.As(err, &rateLimitErr) {
//	    // Handle rate limiting with retry info
//	    time.Sleep(time.Duration(rateLimitErr.RetryAfter) * time.Second)
//	}
//
// # Production Deployment
//
// For production use, configure the client with reliability features:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	client, err := linear.NewClient(apiKey,
//	    linear.WithLogger(logger),           // Structured logging
//	    linear.WithRetry(5, 500*time.Millisecond, 60*time.Second),
//	    linear.WithTimeout(30*time.Second),
//	    linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
//	        // Monitor rate limits for metrics/alerting
//	        log.Printf("Rate limit: %d/%d requests remaining",
//	            info.RequestsRemaining, info.RequestsLimit)
//	    }),
//	    linear.WithTLSConfig(&tls.Config{
//	        MinVersion: tls.VersionTLS12,    // Enforce TLS 1.2+
//	    }),
//	)
//	defer client.Close()
//
// Production features:
//   - Automatic retry with exponential backoff for transient failures
//   - 429 rate limit handling with Retry-After header support
//   - Rate limit monitoring via callback for metrics collection
//   - Structured logging of all requests, responses, and retries
//   - Context timeout/cancellation support on all operations
//   - TLS configuration for security requirements
//
// See examples/production/main.go for a complete production example.
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
