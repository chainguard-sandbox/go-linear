package linear

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	"go.opentelemetry.io/otel/trace"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Client manages communication with the Linear API.
type Client struct {
	gqlClient          intgraphql.LinearGraphQLClient
	config             *ClientConfig
	credentialProvider *credentialCache
}

// NewClient creates a new Linear API client for GraphQL operations.
//
// Parameters:
//   - apiKey: Linear API key from https://linear.app/settings/account/security
//     Format: "lin_api_xxx" for API keys, or OAuth token (Bearer prefix auto-added)
//     Required: yes
//   - opts: Optional configuration via functional options
//     Available: WithTimeout, WithHTTPClient, WithBaseURL, WithUserAgent
//
// Returns:
//   - *Client: Configured client ready for API calls
//   - error: Non-nil if apiKey is empty
//
// Authentication:
//   - API keys: Sent as-is in Authorization header
//   - OAuth tokens: "Bearer " prefix added automatically if not present
//   - Validates on first API call, not during NewClient
//
// Related: [WithTimeout], [WithHTTPClient], [Viewer]
//
// Example:
//
//	client, err := linear.NewClient("lin_api_xxx")
//	if err != nil {
//	    return err
//	}
//
// Example with options:
//
//	client, err := linear.NewClient(apiKey,
//	    linear.WithTimeout(60*time.Second),
//	    linear.WithBaseURL("https://api.linear.app/graphql"),
//	)
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	// Create default configs (apiKey can be empty if WithCredentialProvider is used)
	config := NewDefaultClientConfig(apiKey)
	config.Transport = NewDefaultTransportConfig()

	// Create static credential provider by default (empty apiKey handled)
	credProvider := newCredentialCache(&staticCredentialProvider{apiKey: apiKey})

	// Initialize client with config
	c := &Client{
		config:             config,
		credentialProvider: credProvider,
	}

	// Apply options (they mutate config through client reference)
	for _, opt := range opts {
		opt(c)
	}

	// Validate that we have a way to authenticate
	if apiKey == "" && c.config.CredentialProvider == nil {
		return nil, fmt.Errorf("apiKey is required (or use WithCredentialProvider)")
	}

	// Build and assign transport (conditional wrapping)
	c.config.HTTPClient.Transport = buildTransport(c.config)

	// Create GraphQL client with auth interceptor
	authInterceptor := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res any, next clientv2.RequestInterceptorFunc) error {
		// Get current credential (supports rotation)
		authValue, err := c.credentialProvider.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get credential: %w", err)
		}

		// Normalize authorization header
		if len(authValue) > 7 && authValue[:7] != "Bearer " {
			if len(authValue) > 8 && authValue[:8] != "lin_api_" {
				authValue = "Bearer " + authValue
			}
		}
		req.Header.Set("Authorization", authValue)
		req.Header.Set("User-Agent", c.config.UserAgent)

		// Execute request
		err = next(ctx, req, gqlInfo, res)

		// On 401, refresh credentials and retry once
		if err != nil && isAuthError(err) {
			if _, refreshErr := c.credentialProvider.Refresh(ctx); refreshErr == nil {
				// Retry with fresh credential
				newCred, _ := c.credentialProvider.Get(ctx)
				if len(newCred) > 7 && newCred[:7] != "Bearer " {
					if len(newCred) > 8 && newCred[:8] != "lin_api_" {
						newCred = "Bearer " + newCred
					}
				}
				req.Header.Set("Authorization", newCred)
				return next(ctx, req, gqlInfo, res)
			}
		}

		return err
	}

	c.gqlClient = intgraphql.NewClient(
		c.config.HTTPClient,
		c.config.BaseURL,
		nil,
		authInterceptor,
	)

	return c, nil
}

// Close closes idle connections and cleans up resources.
// It's safe to call Close multiple times.
func (c *Client) Close() error {
	if c.config.HTTPClient != nil {
		c.config.HTTPClient.CloseIdleConnections()
	}
	return nil
}

// isAuthError checks if an error is an authentication error (401).
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error message contains 401 or authentication keywords
	errStr := err.Error()
	return contains(errStr, "401") || contains(errStr, "authentication") || contains(errStr, "unauthorized")
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Viewer returns the currently authenticated user information.
//
// Use this to:
//   - Verify API authentication is working
//   - Get authenticated user details
//   - Check user permissions (admin status)
//
// Returns:
//   - Viewer.ID: User UUID (always populated)
//   - Viewer.Name: Full name (always populated)
//   - Viewer.Email: Email address (always populated)
//   - Viewer.DisplayName: Display/nick name (always populated)
//   - Viewer.Admin: Whether user is org admin (always populated)
//   - error: Non-nil if authentication fails (401) or network error
//
// Permissions Required: Read
//
// Related: [Users], [User]
//
// Example:
//
//	viewer, err := client.Viewer(ctx)
//	if err != nil {
//	    return fmt.Errorf("auth failed: %w", err)
//	}
//	log.Printf("Authenticated as: %s", viewer.Email)
func (c *Client) Viewer(ctx context.Context) (*intgraphql.Viewer_Viewer, error) {
	if c.config.Transport.TracingEnabled {
		var span trace.Span
		ctx, span = startSpan(ctx, "Viewer")
		defer span.End()
	}

	resp, err := c.gqlClient.Viewer(ctx)
	if err != nil {
		return nil, wrapGraphQLError("viewer query", err)
	}
	return &resp.Viewer, nil
}

// Issue retrieves a single issue by ID.
//
// Returns:
//   - Issue.ID: Issue UUID (always populated)
//   - Issue.Title: Issue title (always populated)
//   - Issue.Description: Markdown description (may be empty)
//   - Issue.Priority: 0-4 priority level (always populated)
//   - Issue.Estimate: Story point estimate (may be 0)
//   - Issue.Number: Issue number in team (always populated)
//   - Issue.URL: Linear web URL (always populated)
//   - Issue.State: Workflow state with ID, Name, Type (always populated)
//   - Issue.Team: Team with ID, Name, Key (always populated)
//   - Issue.Assignee: User with ID, Name, DisplayName (nil if unassigned)
//   - Issue.CreatedAt: Creation timestamp (always populated)
//   - Issue.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if issue not found or query fails
//
// Permissions Required: Read
//
// Related: [Issues], [IssueCreate], [IssueUpdate]
func (c *Client) Issue(ctx context.Context, id string) (*intgraphql.GetIssue_Issue, error) {
	resp, err := c.gqlClient.GetIssue(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("issue query", err)
	}
	return &resp.Issue, nil
}

// Issues retrieves a paginated list of issues.
//
// Parameters:
//   - first: Number of issues to return (pointer, nil = server default ~50)
//   - after: Cursor for pagination (pointer, nil = start from beginning)
//
// Returns:
//   - Issues.Nodes: Array of issues (may be empty)
//   - Issues.PageInfo.HasNextPage: true if more results available
//   - Issues.PageInfo.EndCursor: Cursor for next page (pass to after parameter)
//   - error: Non-nil if query fails
//
// Pagination Pattern:
//  1. Call with first=50, after=nil for first page
//  2. Check HasNextPage
//  3. Call again with after=EndCursor for next page
//  4. Repeat until HasNextPage is false
//
// Permissions Required: Read
//
// Related: [Issue], [NewIssueIterator], [IssueSearch]
//
// Example:
//
//	// Get first 10 issues
//	first := int64(10)
//	issues, err := client.Issues(ctx, &first, nil)
//	if err != nil {
//	    return err
//	}
//	for _, issue := range issues.Nodes {
//	    fmt.Println(issue.Title)
//	}
//
// Example pagination:
//
//	cursor := (*string)(nil)
//	for {
//	    issues, err := client.Issues(ctx, &first, cursor)
//	    if err != nil { return err }
//	    // Process issues.Nodes
//	    if !issues.PageInfo.HasNextPage { break }
//	    cursor = issues.PageInfo.EndCursor
//	}
func (c *Client) Issues(ctx context.Context, first *int64, after *string) (*intgraphql.ListIssues_Issues, error) {
	resp, err := c.gqlClient.ListIssues(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("issues query", err)
	}
	return &resp.Issues, nil
}

// IssuesFiltered retrieves a paginated list of issues with filtering.
//
// Parameters:
//   - first: Number of issues to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//   - filter: Issue filter (team, assignee, state, priority, dates, labels)
//
// Returns:
//   - Issues.Nodes: Array of issues matching filter (may be empty)
//   - Issues.PageInfo.HasNextPage: true if more results available
//   - Issues.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Issues], [SearchIssues], [Issue]
func (c *Client) IssuesFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.IssueFilter) (*intgraphql.ListIssuesFiltered_Issues, error) {
	resp, err := c.gqlClient.ListIssuesFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("issues filtered query", err)
	}
	return &resp.Issues, nil
}

// Team retrieves a single team by ID.
//
// Returns:
//   - Team.ID: Team UUID (always populated)
//   - Team.Name: Team name (always populated)
//   - Team.Key: Team identifier used in URLs (always populated)
//   - Team.Description: Team description (may be empty)
//   - Team.Private: Whether team is private (always populated)
//   - Team.Timezone: Team timezone (always populated)
//   - Team.Icon: Team icon emoji (may be empty)
//   - Team.Color: Team color hex code (may be empty)
//   - Team.CreatedAt: Creation timestamp (always populated)
//   - Team.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if team not found or query fails
//
// Permissions Required: Read
//
// Related: [Teams], [TeamMemberships]
func (c *Client) Team(ctx context.Context, id string) (*intgraphql.GetTeam_Team, error) {
	resp, err := c.gqlClient.GetTeam(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("team query", err)
	}
	return &resp.Team, nil
}

// Teams retrieves a paginated list of teams.
//
// Parameters:
//   - first: Number of teams to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Teams.Nodes: Array of teams (may be empty)
//   - Teams.PageInfo.HasNextPage: true if more results available
//   - Teams.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Team], [NewTeamIterator]
func (c *Client) Teams(ctx context.Context, first *int64, after *string) (*intgraphql.ListTeams_Teams, error) {
	resp, err := c.gqlClient.ListTeams(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("teams query", err)
	}
	return &resp.Teams, nil
}

// Project retrieves a single project by ID.
//
// Returns:
//   - Project.ID: Project UUID (always populated)
//   - Project.Name: Project name (always populated)
//   - Project.Description: Project description (may be empty)
//   - Project.State: Project state (backlog/planned/started/completed/canceled)
//   - Project.Progress: Completion progress 0.0-1.0 (always populated)
//   - Project.TargetDate: Target completion date (may be empty)
//   - Project.URL: Linear web URL (always populated)
//   - Project.Icon: Project icon (may be empty)
//   - Project.Color: Project color hex code (may be empty)
//   - Project.Lead: Project lead user with ID, Name, Email (may be nil)
//   - Project.CreatedAt: Creation timestamp (always populated)
//   - Project.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if project not found or query fails
//
// Permissions Required: Read
//
// Related: [Projects], [ProjectCreate], [ProjectUpdate]
func (c *Client) Project(ctx context.Context, id string) (*intgraphql.GetProject_Project, error) {
	resp, err := c.gqlClient.GetProject(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("project query", err)
	}
	return &resp.Project, nil
}

// Projects retrieves a paginated list of projects.
//
// Parameters:
//   - first: Number of projects to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Projects.Nodes: Array of projects (may be empty)
//   - Projects.PageInfo.HasNextPage: true if more results available
//   - Projects.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Project], [NewProjectIterator]
func (c *Client) Projects(ctx context.Context, first *int64, after *string) (*intgraphql.ListProjects_Projects, error) {
	resp, err := c.gqlClient.ListProjects(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("projects query", err)
	}
	return &resp.Projects, nil
}

// User retrieves a single user by ID.
//
// Returns:
//   - User.ID: User UUID (always populated)
//   - User.Name: Full name (always populated)
//   - User.Email: Email address (always populated)
//   - User.DisplayName: Display/nick name (always populated)
//   - User.Admin: Whether user is org admin (always populated)
//   - User.Active: Whether account is active (always populated)
//   - User.AvatarURL: Avatar image URL (may be empty)
//   - User.CreatedAt: Account creation timestamp (always populated)
//   - User.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if user not found or query fails
//
// Permissions Required: Read
//
// Related: [Users], [Viewer]
func (c *Client) User(ctx context.Context, id string) (*intgraphql.GetUser_User, error) {
	resp, err := c.gqlClient.GetUser(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("user query", err)
	}
	return &resp.User, nil
}

// Users retrieves a paginated list of users in the organization.
//
// Parameters:
//   - first: Number of users to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Users.Nodes: Array of users (may be empty)
//   - Users.PageInfo.HasNextPage: true if more results available
//   - Users.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [User], [Viewer]
func (c *Client) Users(ctx context.Context, first *int64, after *string) (*intgraphql.ListUsers_Users, error) {
	resp, err := c.gqlClient.ListUsers(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("users query", err)
	}
	return &resp.Users, nil
}

// Comment retrieves a single comment by ID.
//
// Returns:
//   - Comment.ID: Comment UUID (always populated)
//   - Comment.Body: Comment text in markdown (always populated)
//   - Comment.URL: Linear web URL to comment (always populated)
//   - Comment.User: Comment author with ID, Name (always populated)
//   - Comment.Issue: Parent issue with ID, Title (always populated)
//   - Comment.CreatedAt: Creation timestamp (always populated)
//   - Comment.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if comment not found or query fails
//
// Permissions Required: Read
//
// Related: [Comments], [CommentCreate], [CommentUpdate]
func (c *Client) Comment(ctx context.Context, id string) (*intgraphql.GetComment_Comment, error) {
	resp, err := c.gqlClient.GetComment(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("comment query", err)
	}
	return &resp.Comment, nil
}

// Comments retrieves a paginated list of comments.
//
// Parameters:
//   - first: Number of comments to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Comments.Nodes: Array of comments (may be empty)
//   - Comments.PageInfo.HasNextPage: true if more results available
//   - Comments.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Comment], [NewCommentIterator]
func (c *Client) Comments(ctx context.Context, first *int64, after *string) (*intgraphql.ListComments_Comments, error) {
	resp, err := c.gqlClient.ListComments(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("comments query", err)
	}
	return &resp.Comments, nil
}

// IssueCreate creates a new issue in Linear.
//
// Required Input Fields:
//   - TeamID: Team identifier (string, get from Teams() or Team())
//
// Common Optional Fields (all are pointers, nil = omitted):
//   - Title: Issue title (*string)
//   - Description: Markdown description (*string)
//   - Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low (*int)
//   - AssigneeID: User to assign (*string, get from Users())
//   - StateID: Workflow state (*string, get from WorkflowStates())
//   - LabelIDs: Label identifiers ([]string)
//   - DueDate: Due date (*string, format: YYYY-MM-DD)
//
// Returns:
//   - Issue.ID: Created issue UUID (always populated)
//   - Issue.Number: Issue number in team (always populated)
//   - Issue.Title: Issue title (always populated)
//   - Issue.Team: Team relationship (always populated)
//   - Issue.State: Workflow state (always populated)
//   - error: Non-nil if mutation fails or Success is false
//
// Permissions Required: Write (or issues:create)
//
// Related: [IssueUpdate], [IssueDelete], [Issues]
//
// Example:
//
//	title := "Fix login bug"
//	desc := "Users can't log in on Safari"
//	priority := int64(linear.PriorityUrgent)
//
//	issue, err := client.IssueCreate(ctx, IssueCreateInput{
//	    TeamID: "team-uuid",
//	    Title: &title,
//	    Description: &desc,
//	    Priority: &priority,
//	})
//	if err != nil {
//	    return fmt.Errorf("create failed: %w", err)
//	}
//	log.Printf("Created issue #%d: %s", issue.Number, issue.ID)
func (c *Client) IssueCreate(ctx context.Context, input intgraphql.IssueCreateInput) (*intgraphql.CreateIssue_IssueCreate_Issue, error) {
	resp, err := c.gqlClient.CreateIssue(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueCreate", err)
	}

	if !resp.IssueCreate.Success {
		return nil, errMutationFailed("IssueCreate")
	}

	return resp.IssueCreate.Issue, nil
}

// IssueUpdate updates an existing issue's fields.
//
// Parameters:
//   - id: Issue UUID to update (required)
//   - input: Fields to update (all fields are optional pointers)
//
// Input Fields (all optional, nil = unchanged):
//   - Title: New title (*string)
//   - Description: New description (*string)
//   - Priority: New priority 0-4 (*int)
//   - StateID: New workflow state (*string)
//   - AssigneeID: New assignee (*string, empty string to unassign)
//
// Returns:
//   - Updated issue with modified fields
//   - error: Non-nil if issue not found, permission denied, or mutation fails
//
// Permissions Required: Write
//
// Related: [IssueCreate], [IssueDelete], [Issue]
//
// Example:
//
//	updatedTitle := "Updated: Fix critical bug"
//	priority := int64(linear.PriorityUrgent)
//
//	updated, err := client.IssueUpdate(ctx, issueID, IssueUpdateInput{
//	    Title: &updatedTitle,
//	    Priority: &priority,
//	})
func (c *Client) IssueUpdate(ctx context.Context, id string, input intgraphql.IssueUpdateInput) (*intgraphql.UpdateIssue_IssueUpdate_Issue, error) {
	resp, err := c.gqlClient.UpdateIssue(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueUpdate", err)
	}

	if !resp.IssueUpdate.Success {
		return nil, errMutationFailed("IssueUpdate")
	}

	return resp.IssueUpdate.Issue, nil
}

// IssueBatchUpdate updates multiple issues at once (max 50).
//
// Parameters:
//   - ids: Issue UUIDs to update (max 50)
//   - input: Update to apply to all issues
//
// Returns:
//   - IssueBatchPayload.Issues: Updated issues
//   - IssueBatchPayload.Success: true if successful
//   - error: Non-nil if mutation fails
//
// Permissions Required: Write
//
// Related: [IssueUpdate], [IssuesFiltered]
func (c *Client) IssueBatchUpdate(ctx context.Context, ids []string, input intgraphql.IssueUpdateInput) (*intgraphql.BatchUpdateIssues_IssueBatchUpdate, error) {
	resp, err := c.gqlClient.BatchUpdateIssues(ctx, ids, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueBatchUpdate", err)
	}

	if !resp.IssueBatchUpdate.Success {
		return nil, errMutationFailed("IssueBatchUpdate")
	}

	return &resp.IssueBatchUpdate, nil
}

// IssueDelete permanently deletes an issue.
//
// Parameters:
//   - id: Issue UUID to delete (required)
//
// Returns:
//   - nil: Issue successfully deleted
//   - error: Non-nil if issue not found, permission denied, or deletion fails
//
// Warning: This is permanent and cannot be undone. The issue will be removed
// from all projects, cycles, and relationships.
//
// Permissions Required: Write
//
// Related: [IssueCreate], [IssueUpdate]
//
// Example:
//
//	if err := client.IssueDelete(ctx, issueID); err != nil {
//	    return fmt.Errorf("delete failed: %w", err)
//	}
func (c *Client) IssueDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteIssue(ctx, id)
	if err != nil {
		return wrapGraphQLError("IssueDelete", err)
	}

	if !resp.IssueDelete.Success {
		return errMutationFailed("IssueDelete")
	}

	return nil
}

// Organization retrieves the workspace organization information.
//
// Returns:
//   - Organization.ID: Organization UUID (always populated)
//   - Organization.Name: Organization name (always populated)
//   - Organization.URLKey: URL identifier (always populated)
//   - Organization.LogoURL: Organization logo URL (may be empty)
//   - Organization.RoadmapEnabled: Whether roadmaps are enabled (always populated)
//   - Organization.CreatedAt: Creation timestamp (always populated)
//   - Organization.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Teams], [Users]
func (c *Client) Organization(ctx context.Context) (*intgraphql.GetOrganization_Organization, error) {
	resp, err := c.gqlClient.GetOrganization(ctx)
	if err != nil {
		return nil, wrapGraphQLError("organization query", err)
	}
	return &resp.Organization, nil
}

// IssueLabel retrieves a single issue label by ID.
//
// Returns:
//   - IssueLabel.ID: Label UUID (always populated)
//   - IssueLabel.Name: Label name (always populated)
//   - IssueLabel.Description: Label description (may be empty)
//   - IssueLabel.Color: Label color hex code (always populated)
//   - IssueLabel.CreatedAt: Creation timestamp (always populated)
//   - IssueLabel.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if label not found or query fails
//
// Permissions Required: Read
//
// Related: [IssueLabels], [IssueLabelCreate]
func (c *Client) IssueLabel(ctx context.Context, id string) (*intgraphql.GetLabel_IssueLabel, error) {
	resp, err := c.gqlClient.GetLabel(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("label query", err)
	}
	return &resp.IssueLabel, nil
}

// IssueLabels retrieves a paginated list of issue labels.
//
// Parameters:
//   - first: Number of labels to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - IssueLabels.Nodes: Array of labels (may be empty)
//   - IssueLabels.PageInfo.HasNextPage: true if more results available
//   - IssueLabels.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [IssueLabel], [IssueLabelCreate]
func (c *Client) IssueLabels(ctx context.Context, first *int64, after *string) (*intgraphql.ListLabels_IssueLabels, error) {
	resp, err := c.gqlClient.ListLabels(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("labels query", err)
	}
	return &resp.IssueLabels, nil
}

// WorkflowState retrieves a single workflow state by ID.
//
// Returns:
//   - WorkflowState.ID: State UUID (always populated)
//   - WorkflowState.Name: State name (always populated)
//   - WorkflowState.Description: State description (may be empty)
//   - WorkflowState.Type: State type (unstarted/started/completed/canceled)
//   - WorkflowState.Color: State color hex code (always populated)
//   - WorkflowState.Position: Sort position (always populated)
//   - WorkflowState.CreatedAt: Creation timestamp (always populated)
//   - WorkflowState.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if state not found or query fails
//
// Permissions Required: Read
//
// Related: [WorkflowStates]
func (c *Client) WorkflowState(ctx context.Context, id string) (*intgraphql.GetWorkflowState_WorkflowState, error) {
	resp, err := c.gqlClient.GetWorkflowState(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("workflow state query", err)
	}
	return &resp.WorkflowState, nil
}

// WorkflowStates retrieves a paginated list of workflow states.
//
// Parameters:
//   - first: Number of states to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - WorkflowStates.Nodes: Array of states (may be empty)
//   - WorkflowStates.PageInfo.HasNextPage: true if more results available
//   - WorkflowStates.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [WorkflowState]
func (c *Client) WorkflowStates(ctx context.Context, first *int64, after *string) (*intgraphql.ListWorkflowStates_WorkflowStates, error) {
	resp, err := c.gqlClient.ListWorkflowStates(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("workflow states query", err)
	}
	return &resp.WorkflowStates, nil
}

// Cycle retrieves a single development cycle (sprint) by ID.
//
// Returns:
//   - Cycle.ID: Cycle UUID (always populated)
//   - Cycle.Number: Cycle number in sequence (always populated)
//   - Cycle.Name: Cycle name (may be empty)
//   - Cycle.Description: Cycle description (may be empty)
//   - Cycle.StartsAt: Start date timestamp (always populated)
//   - Cycle.EndsAt: End date timestamp (always populated)
//   - Cycle.CompletedAt: Completion timestamp (nil if not completed)
//   - Cycle.Progress: Completion progress 0.0-1.0 (always populated)
//   - Cycle.Team: Team with ID, Name, Key (always populated)
//   - Cycle.CreatedAt: Creation timestamp (always populated)
//   - Cycle.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if cycle not found or query fails
//
// Permissions Required: Read
//
// Related: [Cycles]
func (c *Client) Cycle(ctx context.Context, id string) (*intgraphql.GetCycle_Cycle, error) {
	resp, err := c.gqlClient.GetCycle(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("cycle query", err)
	}
	return &resp.Cycle, nil
}

// Cycles retrieves a paginated list of development cycles (sprints).
//
// Parameters:
//   - first: Number of cycles to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Cycles.Nodes: Array of cycles (may be empty)
//   - Cycles.PageInfo.HasNextPage: true if more results available
//   - Cycles.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Cycle]
func (c *Client) Cycles(ctx context.Context, first *int64, after *string) (*intgraphql.ListCycles_Cycles, error) {
	resp, err := c.gqlClient.ListCycles(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("cycles query", err)
	}
	return &resp.Cycles, nil
}

// CycleCreate creates a new development cycle (sprint).
//
// Parameters:
//   - input: Cycle creation parameters (name, description, start/end dates, team)
//
// Required fields:
//   - TeamID: UUID of the team (required)
//   - Name: Cycle name (required)
//   - StartsAt: Start date in ISO 8601 format (required)
//   - EndsAt: End date in ISO 8601 format (required)
//
// Optional fields:
//   - Description: Cycle description (nil = empty)
//
// Returns:
//   - Created cycle with ID, number, name, dates, and team
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	startsAt := "2025-01-06"
//	endsAt := "2025-01-17"
//	name := "Sprint 1"
//	cycle, err := client.CycleCreate(ctx, intgraphql.CycleCreateInput{
//	    TeamID:   teamID,
//	    Name:     &name,
//	    StartsAt: &startsAt,
//	    EndsAt:   &endsAt,
//	})
//
// Related: [CycleUpdate], [CycleArchive], [Cycles]
func (c *Client) CycleCreate(ctx context.Context, input intgraphql.CycleCreateInput) (*intgraphql.CreateCycle_CycleCreate_Cycle, error) {
	resp, err := c.gqlClient.CreateCycle(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("CycleCreate", err)
	}
	if !resp.CycleCreate.Success {
		return nil, errMutationFailed("CycleCreate")
	}
	return resp.CycleCreate.Cycle, nil
}

// CycleUpdate updates an existing development cycle.
//
// Parameters:
//   - id: Cycle UUID to update (required)
//   - input: Fields to update (all optional, nil = unchanged)
//
// Optional fields:
//   - Name: Update cycle name
//   - Description: Update description
//   - StartsAt: Update start date (ISO 8601 format)
//   - EndsAt: Update end date (ISO 8601 format)
//
// Returns:
//   - Updated cycle with new values
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	updatedName := "Sprint 1 (Extended)"
//	cycle, err := client.CycleUpdate(ctx, cycleID, intgraphql.CycleUpdateInput{
//	    Name: &updatedName,
//	})
//
// Related: [CycleCreate], [CycleArchive], [Cycle]
func (c *Client) CycleUpdate(ctx context.Context, id string, input intgraphql.CycleUpdateInput) (*intgraphql.UpdateCycle_CycleUpdate_Cycle, error) {
	resp, err := c.gqlClient.UpdateCycle(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("CycleUpdate", err)
	}
	if !resp.CycleUpdate.Success {
		return nil, errMutationFailed("CycleUpdate")
	}
	return resp.CycleUpdate.Cycle, nil
}

// CycleArchive archives (soft-deletes) a development cycle.
//
// Archiving a cycle:
//   - Removes it from active cycle lists
//   - Issues remain associated with the cycle
//   - Cycle data is preserved for historical tracking
//   - Can be unarchived via Linear UI
//
// Parameters:
//   - id: Cycle UUID to archive (required)
//
// Returns:
//   - error: Non-nil if archive fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.CycleArchive(ctx, cycleID)
//	if err != nil {
//	    return fmt.Errorf("failed to archive cycle: %w", err)
//	}
//
// Related: [CycleCreate], [CycleUpdate], [Cycle]
func (c *Client) CycleArchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ArchiveCycle(ctx, id)
	if err != nil {
		return wrapGraphQLError("CycleArchive", err)
	}
	if !resp.CycleArchive.Success {
		return errMutationFailed("CycleArchive")
	}
	return nil
}

// Roadmap retrieves a single product roadmap by ID.
//
// Returns:
//   - Roadmap.ID: Roadmap UUID (always populated)
//   - Roadmap.Name: Roadmap name (always populated)
//   - Roadmap.Description: Roadmap description (may be empty)
//   - Roadmap.SlugID: URL-friendly identifier (always populated)
//   - Roadmap.CreatedAt: Creation timestamp (always populated)
//   - Roadmap.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if roadmap not found or query fails
//
// Permissions Required: Read
//
// Related: [Roadmaps]
func (c *Client) Roadmap(ctx context.Context, id string) (*intgraphql.GetRoadmap_Roadmap, error) {
	resp, err := c.gqlClient.GetRoadmap(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("roadmap query", err)
	}
	return &resp.Roadmap, nil
}

// Roadmaps retrieves a paginated list of product roadmaps.
//
// Parameters:
//   - first: Number of roadmaps to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Roadmaps.Nodes: Array of roadmaps (may be empty)
//   - Roadmaps.PageInfo.HasNextPage: true if more results available
//   - Roadmaps.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Roadmap]
func (c *Client) Roadmaps(ctx context.Context, first *int64, after *string) (*intgraphql.ListRoadmaps_Roadmaps, error) {
	resp, err := c.gqlClient.ListRoadmaps(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("roadmaps query", err)
	}
	return &resp.Roadmaps, nil
}

// Attachment retrieves a single file attachment by ID.
//
// Returns:
//   - Attachment.ID: Attachment UUID (always populated)
//   - Attachment.Title: Attachment title (always populated)
//   - Attachment.Subtitle: Attachment subtitle (may be empty)
//   - Attachment.URL: File URL (always populated)
//   - Attachment.Source: External source (may be nil)
//   - Attachment.SourceType: Source type (may be nil)
//   - Attachment.Metadata: Additional metadata (always populated, may be empty object)
//   - Attachment.CreatedAt: Upload timestamp (always populated)
//   - Attachment.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if attachment not found or query fails
//
// Permissions Required: Read
//
// Related: [Attachments]
func (c *Client) Attachment(ctx context.Context, id string) (*intgraphql.GetAttachment_Attachment, error) {
	resp, err := c.gqlClient.GetAttachment(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("attachment query", err)
	}
	return &resp.Attachment, nil
}

// Attachments retrieves a paginated list of file attachments.
//
// Parameters:
//   - first: Number of attachments to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Attachments.Nodes: Array of attachments (may be empty)
//   - Attachments.PageInfo.HasNextPage: true if more results available
//   - Attachments.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Attachment]
func (c *Client) Attachments(ctx context.Context, first *int64, after *string) (*intgraphql.ListAttachments_Attachments, error) {
	resp, err := c.gqlClient.ListAttachments(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("attachments query", err)
	}
	return &resp.Attachments, nil
}

// AttachmentLinkURL attaches an external URL to an issue.
//
// Parameters:
//   - issueID: Issue UUID to attach to (required)
//   - url: URL to link (required)
//   - title: Display title for the attachment (nil = use URL as title)
//
// Returns:
//   - Created attachment with ID, title, and URL
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	title := "API Documentation"
//	attachment, err := client.AttachmentLinkURL(ctx, issueID,
//	    "https://docs.example.com/api",
//	    &title)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Attached: %s\n", attachment.URL)
//
// Related: [AttachmentLinkGitHubPR], [Attachments]
func (c *Client) AttachmentLinkURL(ctx context.Context, issueID, url string, title *string) (*intgraphql.AttachmentLinkURL_AttachmentLinkURL_Attachment, error) {
	resp, err := c.gqlClient.AttachmentLinkURL(ctx, issueID, url, title)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentLinkURL", err)
	}
	if !resp.AttachmentLinkURL.Success {
		return nil, errMutationFailed("AttachmentLinkURL")
	}
	return &resp.AttachmentLinkURL.Attachment, nil
}

// AttachmentLinkGitHubPR links a GitHub Pull Request to an issue.
//
// The attachment title is automatically extracted from the PR metadata.
// Requires GitHub integration to be configured in Linear workspace.
//
// Parameters:
//   - issueID: Issue UUID to attach to (required)
//   - url: GitHub PR URL (required, e.g., "https://github.com/owner/repo/pull/123")
//
// Returns:
//   - Created attachment with ID, title (from PR), and URL
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	attachment, err := client.AttachmentLinkGitHubPR(ctx, issueID,
//	    "https://github.com/myorg/myrepo/pull/456")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Linked PR: %s\n", attachment.Title)
//
// Related: [AttachmentLinkURL], [Attachments]
func (c *Client) AttachmentLinkGitHubPR(ctx context.Context, issueID, url string) (*intgraphql.AttachmentLinkGitHubPR_AttachmentLinkGitHubPr_Attachment, error) {
	resp, err := c.gqlClient.AttachmentLinkGitHubPr(ctx, issueID, url)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentLinkGitHubPR", err)
	}
	if !resp.AttachmentLinkGitHubPr.Success {
		return nil, errMutationFailed("AttachmentLinkGitHubPR")
	}
	return &resp.AttachmentLinkGitHubPr.Attachment, nil
}

// AttachmentCreate creates a custom attachment for an issue.
//
// This method creates attachments with custom metadata, useful for integrating
// external tools or displaying custom data cards in Linear.
//
// Parameters:
//   - input: Attachment creation parameters (title, issueId, url, metadata, etc.)
//
// Required fields:
//   - IssueID: UUID of the issue to attach to
//   - Title: Attachment title
//
// Optional fields:
//   - URL: External URL for the attachment
//   - Subtitle: Additional context
//   - IconURL: Custom icon (20x20px recommended)
//   - Metadata: JSON object with custom data
//
// Returns:
//   - Created attachment with ID, title, URL, and metadata
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	attachment, err := client.AttachmentCreate(ctx, intgraphql.AttachmentCreateInput{
//	    IssueID: issueID,
//	    Title:   "Build Status",
//	    URL:     stringPtr("https://ci.example.com/builds/123"),
//	    Subtitle: stringPtr("Build #123 - Passed"),
//	})
//
// Related: [AttachmentDelete], [AttachmentLinkURL]
func (c *Client) AttachmentCreate(ctx context.Context, input intgraphql.AttachmentCreateInput) (*intgraphql.AttachmentCreate_AttachmentCreate_Attachment, error) {
	resp, err := c.gqlClient.AttachmentCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentCreate", err)
	}
	if !resp.AttachmentCreate.Success {
		return nil, errMutationFailed("AttachmentCreate")
	}
	return &resp.AttachmentCreate.Attachment, nil
}

// AttachmentLinkSlack links a Slack thread to an issue.
//
// Requires Slack integration to be configured in Linear workspace.
//
// Parameters:
//   - issueID: Issue UUID to attach to (required)
//   - url: Slack message URL (required, e.g., "https://workspace.slack.com/archives/C123/p1234567890")
//
// Returns:
//   - Created attachment with ID, title, and URL
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	attachment, err := client.AttachmentLinkSlack(ctx, issueID,
//	    "https://myworkspace.slack.com/archives/C123ABC/p1234567890123456")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Linked Slack thread: %s\n", attachment.URL)
//
// Related: [AttachmentLinkURL], [Attachments]
func (c *Client) AttachmentLinkSlack(ctx context.Context, issueID, url string) (*intgraphql.AttachmentLinkSlack_AttachmentLinkSlack_Attachment, error) {
	resp, err := c.gqlClient.AttachmentLinkSlack(ctx, issueID, url)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentLinkSlack", err)
	}
	if !resp.AttachmentLinkSlack.Success {
		return nil, errMutationFailed("AttachmentLinkSlack")
	}
	return &resp.AttachmentLinkSlack.Attachment, nil
}

// AttachmentDelete permanently deletes an attachment.
//
// Parameters:
//   - id: Attachment UUID to delete (required)
//
// Returns:
//   - nil: Attachment successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.AttachmentDelete(ctx, attachmentID)
//	if err != nil {
//	    return fmt.Errorf("failed to delete attachment: %w", err)
//	}
//
// Related: [AttachmentCreate], [AttachmentLinkURL]
func (c *Client) AttachmentDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.AttachmentDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("AttachmentDelete", err)
	}
	if !resp.AttachmentDelete.Success {
		return errMutationFailed("AttachmentDelete")
	}
	return nil
}

// Initiative retrieves a single strategic initiative by ID.
//
// Returns:
//   - Initiative.ID: Initiative UUID (always populated)
//   - Initiative.Name: Initiative name (always populated)
//   - Initiative.Description: Initiative description (may be empty)
//   - Initiative.TargetDate: Target completion date (may be empty)
//   - Initiative.SortOrder: Sort position (always populated)
//   - Initiative.Icon: Initiative icon (may be empty)
//   - Initiative.Color: Initiative color hex code (may be empty)
//   - Initiative.CreatedAt: Creation timestamp (always populated)
//   - Initiative.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if initiative not found or query fails
//
// Permissions Required: Read
//
// Related: [Initiatives]
func (c *Client) Initiative(ctx context.Context, id string) (*intgraphql.GetInitiative_Initiative, error) {
	resp, err := c.gqlClient.GetInitiative(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("initiative query", err)
	}
	return &resp.Initiative, nil
}

// Initiatives retrieves a paginated list of strategic initiatives.
//
// Parameters:
//   - first: Number of initiatives to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Initiatives.Nodes: Array of initiatives (may be empty)
//   - Initiatives.PageInfo.HasNextPage: true if more results available
//   - Initiatives.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Initiative]
func (c *Client) Initiatives(ctx context.Context, first *int64, after *string) (*intgraphql.ListInitiatives_Initiatives, error) {
	resp, err := c.gqlClient.ListInitiatives(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("initiatives query", err)
	}
	return &resp.Initiatives, nil
}

// CommentCreate creates a new comment on an issue.
//
// Required Input Fields:
//   - IssueID: Issue to comment on (*string, get from Issues())
//   - Body: Comment text in markdown (*string)
//
// Returns:
//   - Comment.ID: Created comment UUID (always populated)
//   - Comment.Body: Comment text (always populated)
//   - Comment.CreatedAt: Creation timestamp (always populated)
//   - error: Non-nil if issue not found or mutation fails
//
// Permissions Required: Write (or comments:create)
//
// Related: [CommentUpdate], [CommentDelete], [Comments]
//
// Example:
//
//	issueID := "issue-uuid"
//	body := "This looks like a duplicate of #123"
//
//	comment, err := client.CommentCreate(ctx, CommentCreateInput{
//	    IssueID: &issueID,
//	    Body: &body,
//	})
func (c *Client) CommentCreate(ctx context.Context, input intgraphql.CommentCreateInput) (*intgraphql.CreateComment_CommentCreate_Comment, error) {
	resp, err := c.gqlClient.CreateComment(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("CommentCreate", err)
	}

	if !resp.CommentCreate.Success {
		return nil, errMutationFailed("CommentCreate")
	}

	return &resp.CommentCreate.Comment, nil
}

// CommentUpdate updates an existing comment.
func (c *Client) CommentUpdate(ctx context.Context, id string, input intgraphql.CommentUpdateInput) (*intgraphql.UpdateComment_CommentUpdate_Comment, error) {
	resp, err := c.gqlClient.UpdateComment(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("CommentUpdate", err)
	}

	if !resp.CommentUpdate.Success {
		return nil, errMutationFailed("CommentUpdate")
	}

	return &resp.CommentUpdate.Comment, nil
}

// CommentDelete deletes a comment by ID.
func (c *Client) CommentDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteComment(ctx, id)
	if err != nil {
		return wrapGraphQLError("CommentDelete", err)
	}

	if !resp.CommentDelete.Success {
		return errMutationFailed("CommentDelete")
	}

	return nil
}

// SearchIssues searches for issues matching the search term.
//
// The new searchIssues API replaces the deprecated issueSearch endpoint.
// It uses full-text search with optional structured filtering.
//
// Parameters:
//   - term: Search text (required). Searches across issue titles, descriptions, and optionally comments.
//   - first: Number of results per page (nil = default ~50, max: 250)
//   - after: Pagination cursor from previous PageInfo.EndCursor (nil = first page)
//   - filter: Optional structured filters (assignee, state, priority, team, etc.)
//   - includeArchived: Include archived issues in results (default: false)
//
// Returns:
//   - SearchIssues with nodes (matching issues) and pageInfo (pagination)
//   - totalCount: Total number of matches (useful for showing "X of Y results")
//   - Error on failure (network, auth, invalid filter)
//
// Example (simple text search):
//
//	term := "bug"
//	first := int64(50)
//	issues, err := client.SearchIssues(ctx, term, &first, nil, nil, nil)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Found %d results\n", len(issues.Nodes))
//	for _, issue := range issues.Nodes {
//	    fmt.Printf("[%.0f] %s\n", issue.Number, issue.Title)
//	}
//
// Example (with filter):
//
//	term := "login"
//	first := int64(50)
//	priorityHigh := float64(2)
//	filter := &intgraphql.IssueFilter{
//	    Priority: &intgraphql.NullableNumberComparator{Eq: &priorityHigh},
//	}
//	issues, err := client.SearchIssues(ctx, term, &first, nil, filter, nil)
//
// Related: [Issues], [IssueCreate], [IssueUpdate]
func (c *Client) SearchIssues(ctx context.Context, term string, first *int64, after *string, filter *intgraphql.IssueFilter, includeArchived *bool) (*intgraphql.SearchIssues_SearchIssues, error) {
	resp, err := c.gqlClient.SearchIssues(ctx, term, first, after, filter, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("issue search failed: %w", err)
	}
	return &resp.SearchIssues, nil
}

// Document retrieves a single knowledge base document by ID.
//
// Returns:
//   - Document.ID: Document UUID (always populated)
//   - Document.Title: Document title (always populated)
//   - Document.Content: Document content in markdown (may be empty)
//   - Document.SlugID: URL-friendly identifier (always populated)
//   - Document.Icon: Document icon (may be empty)
//   - Document.Color: Document color hex code (may be empty)
//   - Document.CreatedAt: Creation timestamp (always populated)
//   - Document.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if document not found or query fails
//
// Permissions Required: Read
//
// Related: [Documents]
func (c *Client) Document(ctx context.Context, id string) (*intgraphql.GetDocument_Document, error) {
	resp, err := c.gqlClient.GetDocument(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("document query", err)
	}
	return &resp.Document, nil
}

// Documents retrieves a paginated list of knowledge base documents.
//
// Parameters:
//   - first: Number of documents to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Documents.Nodes: Array of documents (may be empty)
//   - Documents.PageInfo.HasNextPage: true if more results available
//   - Documents.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Document]
func (c *Client) Documents(ctx context.Context, first *int64, after *string) (*intgraphql.ListDocuments_Documents, error) {
	resp, err := c.gqlClient.ListDocuments(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("documents query", err)
	}
	return &resp.Documents, nil
}

// Template retrieves a single issue template by ID.
//
// Returns:
//   - Template.ID: Template UUID (always populated)
//   - Template.Name: Template name (always populated)
//   - Template.Description: Template description (may be empty)
//   - Template.Type: Template type (always populated)
//   - Template.CreatedAt: Creation timestamp (always populated)
//   - Template.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if template not found or query fails
//
// Permissions Required: Read
//
// Related: [Templates]
func (c *Client) Template(ctx context.Context, id string) (*intgraphql.GetTemplate_Template, error) {
	resp, err := c.gqlClient.GetTemplate(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("template query", err)
	}
	return &resp.Template, nil
}

// Templates retrieves all templates.
func (c *Client) Templates(ctx context.Context) ([]*intgraphql.ListTemplates_Templates, error) {
	resp, err := c.gqlClient.ListTemplates(ctx)
	if err != nil {
		return nil, wrapGraphQLError("templates query", err)
	}
	return resp.Templates, nil
}

// IssueLabelCreate creates a new label.
func (c *Client) IssueLabelCreate(ctx context.Context, input intgraphql.IssueLabelCreateInput) (*intgraphql.CreateLabel_IssueLabelCreate_IssueLabel, error) {
	resp, err := c.gqlClient.CreateLabel(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueLabelCreate", err)
	}

	if !resp.IssueLabelCreate.Success {
		return nil, errMutationFailed("IssueLabelCreate")
	}

	return &resp.IssueLabelCreate.IssueLabel, nil
}

// IssueLabelUpdate updates an existing label.
func (c *Client) IssueLabelUpdate(ctx context.Context, id string, input intgraphql.IssueLabelUpdateInput) (*intgraphql.UpdateLabel_IssueLabelUpdate_IssueLabel, error) {
	resp, err := c.gqlClient.UpdateLabel(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueLabelUpdate", err)
	}

	if !resp.IssueLabelUpdate.Success {
		return nil, errMutationFailed("IssueLabelUpdate")
	}

	return &resp.IssueLabelUpdate.IssueLabel, nil
}

// IssueLabelDelete deletes a label by ID.
func (c *Client) IssueLabelDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteLabel(ctx, id)
	if err != nil {
		return wrapGraphQLError("IssueLabelDelete", err)
	}

	if !resp.IssueLabelDelete.Success {
		return errMutationFailed("IssueLabelDelete")
	}

	return nil
}

// IssueAddLabel adds a label to an issue.
//
// Simpler alternative to IssueUpdate with AddedLabelIds array.
// Does not require fetching existing labels first.
//
// Parameters:
//   - id: Issue UUID (required)
//   - labelID: Label UUID to add (required)
//
// Returns:
//   - Updated issue with labels collection
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	issue, err := client.IssueAddLabel(ctx, issueID, labelID)
//	if err != nil {
//	    return fmt.Errorf("failed to add label: %w", err)
//	}
//	fmt.Printf("Issue now has %d labels\n", len(issue.Labels.Nodes))
//
// Related: [IssueRemoveLabel], [IssueLabelCreate], [IssueUpdate]
func (c *Client) IssueAddLabel(ctx context.Context, id, labelID string) (*intgraphql.IssueAddLabel_IssueAddLabel_Issue, error) {
	resp, err := c.gqlClient.IssueAddLabel(ctx, id, labelID)
	if err != nil {
		return nil, wrapGraphQLError("IssueAddLabel", err)
	}
	if !resp.IssueAddLabel.Success {
		return nil, errMutationFailed("IssueAddLabel")
	}
	return resp.IssueAddLabel.Issue, nil
}

// IssueRemoveLabel removes a label from an issue.
//
// Simpler alternative to IssueUpdate with array manipulation.
// Does not require fetching and filtering existing labels.
//
// Parameters:
//   - id: Issue UUID (required)
//   - labelID: Label UUID to remove (required)
//
// Returns:
//   - Updated issue with labels collection
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	issue, err := client.IssueRemoveLabel(ctx, issueID, labelID)
//	if err != nil {
//	    return fmt.Errorf("failed to remove label: %w", err)
//	}
//	fmt.Printf("Issue now has %d labels\n", len(issue.Labels.Nodes))
//
// Related: [IssueAddLabel], [IssueLabelDelete], [IssueUpdate]
func (c *Client) IssueRemoveLabel(ctx context.Context, id, labelID string) (*intgraphql.IssueRemoveLabel_IssueRemoveLabel_Issue, error) {
	resp, err := c.gqlClient.IssueRemoveLabel(ctx, id, labelID)
	if err != nil {
		return nil, wrapGraphQLError("IssueRemoveLabel", err)
	}
	if !resp.IssueRemoveLabel.Success {
		return nil, errMutationFailed("IssueRemoveLabel")
	}
	return resp.IssueRemoveLabel.Issue, nil
}

// IssueRelationCreate creates a relationship between two issues.
//
// Relationship types:
//   - "blocks": This issue blocks another issue
//   - "blocked": This issue is blocked by another
//   - "duplicate": This issue is a duplicate of another
//   - "related": This issue is related to another
//
// Parameters:
//   - input: Relation parameters (issueId, relatedIssueId, type)
//
// Returns:
//   - Created relation with both issues and type
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	relation, err := client.IssueRelationCreate(ctx, intgraphql.IssueRelationCreateInput{
//	    IssueID:        &currentIssueID,
//	    RelatedIssueID: &blockerIssueID,
//	    Type:           "blocks",
//	})
//	fmt.Printf("Issue %s blocks %s\n", relation.Issue.Title, relation.RelatedIssue.Title)
//
// Related: [IssueRelationUpdate], [IssueRelationDelete]
func (c *Client) IssueRelationCreate(ctx context.Context, input intgraphql.IssueRelationCreateInput) (*intgraphql.IssueRelationCreate_IssueRelationCreate_IssueRelation, error) {
	resp, err := c.gqlClient.IssueRelationCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueRelationCreate", err)
	}
	if !resp.IssueRelationCreate.Success {
		return nil, errMutationFailed("IssueRelationCreate")
	}
	return &resp.IssueRelationCreate.IssueRelation, nil
}

// IssueRelationUpdate updates an existing relationship between issues.
//
// Use this to change the relationship type (e.g., from "related" to "blocks").
//
// Parameters:
//   - id: IssueRelation UUID to update (required)
//   - input: Fields to update (type is the main updatable field)
//
// Returns:
//   - Updated relation with new type
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	updatedType := "blocks"
//	relation, err := client.IssueRelationUpdate(ctx, relationID, intgraphql.IssueRelationUpdateInput{
//	    Type: &updatedType,
//	})
//
// Related: [IssueRelationCreate], [IssueRelationDelete]
func (c *Client) IssueRelationUpdate(ctx context.Context, id string, input intgraphql.IssueRelationUpdateInput) (*intgraphql.IssueRelationUpdate_IssueRelationUpdate_IssueRelation, error) {
	resp, err := c.gqlClient.IssueRelationUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueRelationUpdate", err)
	}
	if !resp.IssueRelationUpdate.Success {
		return nil, errMutationFailed("IssueRelationUpdate")
	}
	return &resp.IssueRelationUpdate.IssueRelation, nil
}

// IssueRelationDelete deletes a relationship between issues.
//
// Parameters:
//   - id: IssueRelation UUID to delete (required)
//
// Returns:
//   - nil: Relation successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.IssueRelationDelete(ctx, relationID)
//	if err != nil {
//	    return fmt.Errorf("failed to delete relation: %w", err)
//	}
//
// Related: [IssueRelationCreate], [IssueRelationUpdate]
func (c *Client) IssueRelationDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.IssueRelationDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("IssueRelationDelete", err)
	}
	if !resp.IssueRelationDelete.Success {
		return errMutationFailed("IssueRelationDelete")
	}
	return nil
}

// TeamCreate creates a new team.
func (c *Client) TeamCreate(ctx context.Context, input intgraphql.TeamCreateInput) (*intgraphql.CreateTeam_TeamCreate_Team, error) {
	resp, err := c.gqlClient.CreateTeam(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("TeamCreate", err)
	}

	if !resp.TeamCreate.Success {
		return nil, errMutationFailed("TeamCreate")
	}

	return resp.TeamCreate.Team, nil
}

// TeamUpdate updates an existing team.
func (c *Client) TeamUpdate(ctx context.Context, id string, input intgraphql.TeamUpdateInput) (*intgraphql.UpdateTeam_TeamUpdate_Team, error) {
	resp, err := c.gqlClient.UpdateTeam(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("TeamUpdate", err)
	}

	if !resp.TeamUpdate.Success {
		return nil, errMutationFailed("TeamUpdate")
	}

	return resp.TeamUpdate.Team, nil
}

// TeamDelete deletes a team by ID.
func (c *Client) TeamDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteTeam(ctx, id)
	if err != nil {
		return wrapGraphQLError("TeamDelete", err)
	}

	if !resp.TeamDelete.Success {
		return errMutationFailed("TeamDelete")
	}

	return nil
}

// ProjectCreate creates a new project.
func (c *Client) ProjectCreate(ctx context.Context, input intgraphql.ProjectCreateInput) (*intgraphql.CreateProject_ProjectCreate_Project, error) {
	resp, err := c.gqlClient.CreateProject(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectCreate", err)
	}

	if !resp.ProjectCreate.Success {
		return nil, errMutationFailed("ProjectCreate")
	}

	return resp.ProjectCreate.Project, nil
}

// ProjectUpdate updates an existing project.
func (c *Client) ProjectUpdate(ctx context.Context, id string, input intgraphql.ProjectUpdateInput) (*intgraphql.UpdateProject_ProjectUpdate_Project, error) {
	resp, err := c.gqlClient.UpdateProject(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectUpdate", err)
	}

	if !resp.ProjectUpdate.Success {
		return nil, errMutationFailed("ProjectUpdate")
	}

	return resp.ProjectUpdate.Project, nil
}

// ProjectDelete deletes a project by ID.
func (c *Client) ProjectDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteProject(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectDelete", err)
	}

	if !resp.ProjectDelete.Success {
		return errMutationFailed("ProjectDelete")
	}

	return nil
}

// ProjectMilestoneCreate creates a new milestone within a project.
//
// Milestones represent phases or stages within a project (e.g., "Q1 2025", "Beta Launch", "v1.0 Release").
//
// Parameters:
//   - input: Milestone creation parameters
//
// Required fields:
//   - ProjectID: UUID of the project (required)
//   - Name: Milestone name (required)
//
// Optional fields:
//   - Description: Milestone description
//   - TargetDate: Due date in ISO 8601 format (YYYY-MM-DD)
//   - SortOrder: Display order (lower numbers first)
//
// Returns:
//   - Created milestone with ID, name, dates, and project
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	name := "Beta Launch"
//	targetDate := "2025-03-01"
//	milestone, err := client.ProjectMilestoneCreate(ctx, intgraphql.ProjectMilestoneCreateInput{
//	    ProjectID:  &projectID,
//	    Name:       &name,
//	    TargetDate: &targetDate,
//	})
//
// Related: [ProjectMilestoneUpdate], [ProjectMilestoneDelete], [Projects]
func (c *Client) ProjectMilestoneCreate(ctx context.Context, input intgraphql.ProjectMilestoneCreateInput) (*intgraphql.ProjectMilestoneCreate_ProjectMilestoneCreate_ProjectMilestone, error) {
	resp, err := c.gqlClient.ProjectMilestoneCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectMilestoneCreate", err)
	}
	if !resp.ProjectMilestoneCreate.Success {
		return nil, errMutationFailed("ProjectMilestoneCreate")
	}
	return &resp.ProjectMilestoneCreate.ProjectMilestone, nil
}

// ProjectMilestoneUpdate updates an existing project milestone.
//
// Parameters:
//   - id: Milestone UUID to update (required)
//   - input: Fields to update (all optional, nil = unchanged)
//
// Optional fields:
//   - Name: Update milestone name
//   - Description: Update description
//   - TargetDate: Update target date (ISO 8601 format)
//   - SortOrder: Update display order
//
// Returns:
//   - Updated milestone with new values
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	updatedTargetDate := "2025-04-01"
//	milestone, err := client.ProjectMilestoneUpdate(ctx, milestoneID, intgraphql.ProjectMilestoneUpdateInput{
//	    TargetDate: &updatedTargetDate,
//	})
//
// Related: [ProjectMilestoneCreate], [ProjectMilestoneDelete]
func (c *Client) ProjectMilestoneUpdate(ctx context.Context, id string, input intgraphql.ProjectMilestoneUpdateInput) (*intgraphql.ProjectMilestoneUpdate_ProjectMilestoneUpdate_ProjectMilestone, error) {
	resp, err := c.gqlClient.ProjectMilestoneUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectMilestoneUpdate", err)
	}
	if !resp.ProjectMilestoneUpdate.Success {
		return nil, errMutationFailed("ProjectMilestoneUpdate")
	}
	return &resp.ProjectMilestoneUpdate.ProjectMilestone, nil
}

// ProjectMilestoneDelete deletes a project milestone.
//
// Parameters:
//   - id: Milestone UUID to delete (required)
//
// Returns:
//   - nil: Milestone successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Note: Issues associated with the milestone are not deleted.
//
// Permissions Required: Write
//
// Example:
//
//	err := client.ProjectMilestoneDelete(ctx, milestoneID)
//	if err != nil {
//	    return fmt.Errorf("failed to delete milestone: %w", err)
//	}
//
// Related: [ProjectMilestoneCreate], [ProjectMilestoneUpdate]
func (c *Client) ProjectMilestoneDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ProjectMilestoneDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectMilestoneDelete", err)
	}
	if !resp.ProjectMilestoneDelete.Success {
		return errMutationFailed("ProjectMilestoneDelete")
	}
	return nil
}

// FavoriteCreate stars an issue or project for quick access.
//
// Starred items appear in Linear's "Favorites" section for easy retrieval.
//
// Parameters:
//   - input: Favorite creation parameters (must include either IssueID or ProjectID)
//
// Returns:
//   - Created favorite with ID, type, and referenced resource
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example (star an issue):
//
//	favorite, err := client.FavoriteCreate(ctx, intgraphql.FavoriteCreateInput{
//	    IssueID: &issueID,
//	})
//	fmt.Printf("Starred issue: %s\n", favorite.Issue.Title)
//
// Example (star a project):
//
//	favorite, err := client.FavoriteCreate(ctx, intgraphql.FavoriteCreateInput{
//	    ProjectID: &projectID,
//	})
//	fmt.Printf("Starred project: %s\n", favorite.Project.Name)
//
// Related: [FavoriteDelete]
func (c *Client) FavoriteCreate(ctx context.Context, input intgraphql.FavoriteCreateInput) (*intgraphql.FavoriteCreate_FavoriteCreate_Favorite, error) {
	resp, err := c.gqlClient.FavoriteCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("FavoriteCreate", err)
	}
	if !resp.FavoriteCreate.Success {
		return nil, errMutationFailed("FavoriteCreate")
	}
	return &resp.FavoriteCreate.Favorite, nil
}

// FavoriteDelete removes a favorite (unstar).
//
// Parameters:
//   - id: Favorite UUID to delete (required)
//
// Returns:
//   - nil: Favorite successfully deleted (item unstarred)
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.FavoriteDelete(ctx, favoriteID)
//	if err != nil {
//	    return fmt.Errorf("failed to unstar: %w", err)
//	}
//
// Related: [FavoriteCreate]
func (c *Client) FavoriteDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.FavoriteDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("FavoriteDelete", err)
	}
	if !resp.FavoriteDelete.Success {
		return errMutationFailed("FavoriteDelete")
	}
	return nil
}

// ReactionCreate adds an emoji reaction to an issue or comment.
//
// Reactions provide quick feedback without creating a comment.
//
// Parameters:
//   - input: Reaction parameters (emoji, and either IssueID or CommentID)
//
// Returns:
//   - Created reaction with emoji and user
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example (react to issue):
//
//	reaction, err := client.ReactionCreate(ctx, intgraphql.ReactionCreateInput{
//	    Emoji:   "👍",
//	    IssueID: &issueID,
//	})
//	fmt.Printf("%s reacted with %s\n", reaction.User.Name, reaction.Emoji)
//
// Example (react to comment):
//
//	reaction, err := client.ReactionCreate(ctx, intgraphql.ReactionCreateInput{
//	    Emoji:     "❤️",
//	    CommentID: &commentID,
//	})
//
// Related: [ReactionDelete]
func (c *Client) ReactionCreate(ctx context.Context, input intgraphql.ReactionCreateInput) (*intgraphql.ReactionCreate_ReactionCreate_Reaction, error) {
	resp, err := c.gqlClient.ReactionCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ReactionCreate", err)
	}
	if !resp.ReactionCreate.Success {
		return nil, errMutationFailed("ReactionCreate")
	}
	return &resp.ReactionCreate.Reaction, nil
}

// ReactionDelete removes a reaction from an issue or comment.
//
// Parameters:
//   - id: Reaction UUID to delete (required)
//
// Returns:
//   - nil: Reaction successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.ReactionDelete(ctx, reactionID)
//	if err != nil {
//	    return fmt.Errorf("failed to remove reaction: %w", err)
//	}
//
// Related: [ReactionCreate]
func (c *Client) ReactionDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ReactionDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("ReactionDelete", err)
	}
	if !resp.ReactionDelete.Success {
		return errMutationFailed("ReactionDelete")
	}
	return nil
}

// NotificationUpdate marks a notification as read or archives it.
//
// Parameters:
//   - id: Notification UUID to update (required)
//   - input: Fields to update (readAt, archivedAt, snoozedUntilAt)
//
// Returns:
//   - Updated notification with new status
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example (mark as read):
//
//	now := time.Now()
//	notification, err := client.NotificationUpdate(ctx, notificationID, intgraphql.NotificationUpdateInput{
//	    ReadAt: &now,
//	})
//
// Related: [NotificationArchive]
func (c *Client) NotificationUpdate(ctx context.Context, id string, input intgraphql.NotificationUpdateInput) (*intgraphql.NotificationUpdate_NotificationUpdate_Notification, error) {
	resp, err := c.gqlClient.NotificationUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("NotificationUpdate", err)
	}
	if !resp.NotificationUpdate.Success {
		return nil, errMutationFailed("NotificationUpdate")
	}
	return &resp.NotificationUpdate.Notification, nil
}

// NotificationArchive archives a notification.
//
// Parameters:
//   - id: Notification UUID to archive (required)
//
// Returns:
//   - nil: Notification successfully archived
//   - error: Non-nil if archive fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.NotificationArchive(ctx, notificationID)
//
// Related: [NotificationUpdate], [NotificationSubscriptionCreate]
func (c *Client) NotificationArchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.NotificationArchive(ctx, id)
	if err != nil {
		return wrapGraphQLError("NotificationArchive", err)
	}
	if !resp.NotificationArchive.Success {
		return errMutationFailed("NotificationArchive")
	}
	return nil
}

// NotificationSubscriptionCreate subscribes to notifications for an issue or project.
//
// Parameters:
//   - input: Subscription parameters (issueId OR projectId)
//
// Returns:
//   - Created subscription with ID and active status
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	subscription, err := client.NotificationSubscriptionCreate(ctx, intgraphql.NotificationSubscriptionCreateInput{
//	    IssueID: &issueID,
//	})
//
// Related: [NotificationSubscriptionDelete]
func (c *Client) NotificationSubscriptionCreate(ctx context.Context, input intgraphql.NotificationSubscriptionCreateInput) (*intgraphql.NotificationSubscriptionCreate_NotificationSubscriptionCreate_NotificationSubscription, error) {
	resp, err := c.gqlClient.NotificationSubscriptionCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("NotificationSubscriptionCreate", err)
	}
	if !resp.NotificationSubscriptionCreate.Success {
		return nil, errMutationFailed("NotificationSubscriptionCreate")
	}
	return &resp.NotificationSubscriptionCreate.NotificationSubscription, nil
}

// NotificationSubscriptionDelete unsubscribes from notifications.
//
// Parameters:
//   - id: Subscription UUID to delete (required)
//
// Returns:
//   - nil: Subscription successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.NotificationSubscriptionDelete(ctx, subscriptionID)
//
// Related: [NotificationSubscriptionCreate]
func (c *Client) NotificationSubscriptionDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.NotificationSubscriptionDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("NotificationSubscriptionDelete", err)
	}
	if !resp.NotificationSubscriptionDelete.Success {
		return errMutationFailed("NotificationSubscriptionDelete")
	}
	return nil
}
