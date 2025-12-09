package linear

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	intgraphql "github.com/eslerm/go-linear/internal/graphql"
)

// Client manages communication with the Linear API.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	apiKey      string
	userAgent   string
	gqlClient   intgraphql.LinearGraphQLClient
	logger      *slog.Logger
	onRateLimit func(*RateLimitInfo)

	// Transport configuration
	baseTransport    *http.Transport
	maxRetries       int
	initialBackoff   time.Duration
	maxBackoff       time.Duration
	maxRetryDuration time.Duration
	metricsEnabled   bool
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
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}

	baseTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 3, // Match Linear's ~2 req/sec rate limit
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: baseTransport,
		},
		baseURL:          "https://api.linear.app/graphql",
		apiKey:           apiKey,
		userAgent:        "go-linear/0.1.0",
		baseTransport:    baseTransport,
		maxRetries:       3,
		initialBackoff:   1 * time.Second,
		maxBackoff:       30 * time.Second,
		maxRetryDuration: 90 * time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Wrap transport with retry/rate-limit handling if configured
	if c.maxRetries > 0 || c.logger != nil || c.onRateLimit != nil || c.metricsEnabled {
		transport := &Transport{
			Base:             c.httpClient.Transport,
			Logger:           c.logger,
			MaxRetries:       c.maxRetries,
			InitialBackoff:   c.initialBackoff,
			MaxBackoff:       c.maxBackoff,
			MaxRetryDuration: c.maxRetryDuration,
			OnRateLimit:      c.onRateLimit,
			MetricsEnabled:   c.metricsEnabled,
		}
		c.httpClient.Transport = transport
	}

	// Create gqlgenc client with auth interceptor
	authInterceptor := func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res any, next clientv2.RequestInterceptorFunc) error {
		// Set Authorization header
		authValue := c.apiKey
		if len(authValue) > 7 && authValue[:7] != "Bearer " {
			if len(authValue) > 8 && authValue[:8] != "lin_api_" {
				authValue = "Bearer " + authValue
			}
		}
		req.Header.Set("Authorization", authValue)
		req.Header.Set("User-Agent", c.userAgent)
		return next(ctx, req, gqlInfo, res)
	}

	c.gqlClient = intgraphql.NewClient(c.httpClient, c.baseURL, nil, authInterceptor)

	return c, nil
}

// Close closes idle connections and cleans up resources.
// It's safe to call Close multiple times.
func (c *Client) Close() error {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
	return nil
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
	resp, err := c.gqlClient.Viewer(ctx)
	if err != nil {
		return nil, fmt.Errorf("viewer query failed: %w", err)
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
		return nil, fmt.Errorf("issue query failed: %w", err)
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
		return nil, fmt.Errorf("issues query failed: %w", err)
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
		return nil, fmt.Errorf("team query failed: %w", err)
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
		return nil, fmt.Errorf("teams query failed: %w", err)
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
		return nil, fmt.Errorf("project query failed: %w", err)
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
		return nil, fmt.Errorf("projects query failed: %w", err)
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
		return nil, fmt.Errorf("user query failed: %w", err)
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
		return nil, fmt.Errorf("users query failed: %w", err)
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
		return nil, fmt.Errorf("comment query failed: %w", err)
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
		return nil, fmt.Errorf("comments query failed: %w", err)
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
//	priority := 1 // Urgent
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
//	newTitle := "Updated: Fix critical bug"
//	priority := 1 // Urgent
//
//	updated, err := client.IssueUpdate(ctx, issueID, IssueUpdateInput{
//	    Title: &newTitle,
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
		return nil, fmt.Errorf("organization query failed: %w", err)
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
		return nil, fmt.Errorf("label query failed: %w", err)
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
		return nil, fmt.Errorf("labels query failed: %w", err)
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
		return nil, fmt.Errorf("workflow state query failed: %w", err)
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
		return nil, fmt.Errorf("workflow states query failed: %w", err)
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
		return nil, fmt.Errorf("cycle query failed: %w", err)
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
		return nil, fmt.Errorf("cycles query failed: %w", err)
	}
	return &resp.Cycles, nil
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
		return nil, fmt.Errorf("roadmap query failed: %w", err)
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
		return nil, fmt.Errorf("roadmaps query failed: %w", err)
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
		return nil, fmt.Errorf("attachment query failed: %w", err)
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
		return nil, fmt.Errorf("attachments query failed: %w", err)
	}
	return &resp.Attachments, nil
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
		return nil, fmt.Errorf("initiative query failed: %w", err)
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
		return nil, fmt.Errorf("initiatives query failed: %w", err)
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

// IssueSearch searches for issues matching a query string.
func (c *Client) IssueSearch(ctx context.Context, query string, first *int64, after *string) (*intgraphql.SearchIssues_IssueSearch, error) {
	resp, err := c.gqlClient.SearchIssues(ctx, query, first, after)
	if err != nil {
		return nil, fmt.Errorf("issue search failed: %w", err)
	}
	return &resp.IssueSearch, nil
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
		return nil, fmt.Errorf("document query failed: %w", err)
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
		return nil, fmt.Errorf("documents query failed: %w", err)
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
		return nil, fmt.Errorf("template query failed: %w", err)
	}
	return &resp.Template, nil
}

// Templates retrieves all templates.
func (c *Client) Templates(ctx context.Context) ([]*intgraphql.ListTemplates_Templates, error) {
	resp, err := c.gqlClient.ListTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("templates query failed: %w", err)
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
