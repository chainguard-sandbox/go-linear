package linear

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Yamashou/gqlgenc/clientv2"
	intgraphql "github.com/eslerm/go-linear/internal/graphql"
)

// Client manages communication with the Linear API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
	gqlClient  intgraphql.LinearGraphQLClient
}

// NewClient creates a new Linear API client.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		baseURL:   "https://api.linear.app/graphql",
		apiKey:    apiKey,
		userAgent: "go-linear/0.1.0",
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
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

// Viewer returns the currently authenticated user.
func (c *Client) Viewer(ctx context.Context) (*intgraphql.Viewer_Viewer, error) {
	resp, err := c.gqlClient.Viewer(ctx)
	if err != nil {
		return nil, fmt.Errorf("viewer query failed: %w", err)
	}
	return &resp.Viewer, nil
}

// Issue retrieves a single issue by ID.
func (c *Client) Issue(ctx context.Context, id string) (*intgraphql.GetIssue_Issue, error) {
	resp, err := c.gqlClient.GetIssue(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("issue query failed: %w", err)
	}
	return &resp.Issue, nil
}

// Issues retrieves a paginated list of issues.
func (c *Client) Issues(ctx context.Context, first *int64, after *string) (*intgraphql.ListIssues_Issues, error) {
	resp, err := c.gqlClient.ListIssues(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("issues query failed: %w", err)
	}
	return &resp.Issues, nil
}

// Team retrieves a single team by ID.
func (c *Client) Team(ctx context.Context, id string) (*intgraphql.GetTeam_Team, error) {
	resp, err := c.gqlClient.GetTeam(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("team query failed: %w", err)
	}
	return &resp.Team, nil
}

// Teams retrieves a paginated list of teams.
func (c *Client) Teams(ctx context.Context, first *int64, after *string) (*intgraphql.ListTeams_Teams, error) {
	resp, err := c.gqlClient.ListTeams(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("teams query failed: %w", err)
	}
	return &resp.Teams, nil
}

// Project retrieves a single project by ID.
func (c *Client) Project(ctx context.Context, id string) (*intgraphql.GetProject_Project, error) {
	resp, err := c.gqlClient.GetProject(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("project query failed: %w", err)
	}
	return &resp.Project, nil
}

// Projects retrieves a paginated list of projects.
func (c *Client) Projects(ctx context.Context, first *int64, after *string) (*intgraphql.ListProjects_Projects, error) {
	resp, err := c.gqlClient.ListProjects(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("projects query failed: %w", err)
	}
	return &resp.Projects, nil
}

// User retrieves a single user by ID.
func (c *Client) User(ctx context.Context, id string) (*intgraphql.GetUser_User, error) {
	resp, err := c.gqlClient.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user query failed: %w", err)
	}
	return &resp.User, nil
}

// Users retrieves a paginated list of users.
func (c *Client) Users(ctx context.Context, first *int64, after *string) (*intgraphql.ListUsers_Users, error) {
	resp, err := c.gqlClient.ListUsers(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("users query failed: %w", err)
	}
	return &resp.Users, nil
}

// Comment retrieves a single comment by ID.
func (c *Client) Comment(ctx context.Context, id string) (*intgraphql.GetComment_Comment, error) {
	resp, err := c.gqlClient.GetComment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("comment query failed: %w", err)
	}
	return &resp.Comment, nil
}

// Comments retrieves a paginated list of comments.
func (c *Client) Comments(ctx context.Context, first *int64, after *string) (*intgraphql.ListComments_Comments, error) {
	resp, err := c.gqlClient.ListComments(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("comments query failed: %w", err)
	}
	return &resp.Comments, nil
}

// IssueCreate creates a new issue.
func (c *Client) IssueCreate(ctx context.Context, input intgraphql.IssueCreateInput) (*intgraphql.CreateIssue_IssueCreate_Issue, error) {
	resp, err := c.gqlClient.CreateIssue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("issue create failed: %w", err)
	}

	if !resp.IssueCreate.Success {
		return nil, fmt.Errorf("issue create failed: success=false")
	}

	return resp.IssueCreate.Issue, nil
}

// IssueUpdate updates an existing issue.
func (c *Client) IssueUpdate(ctx context.Context, id string, input intgraphql.IssueUpdateInput) (*intgraphql.UpdateIssue_IssueUpdate_Issue, error) {
	resp, err := c.gqlClient.UpdateIssue(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("issue update failed: %w", err)
	}

	if !resp.IssueUpdate.Success {
		return nil, fmt.Errorf("issue update failed: success=false")
	}

	return resp.IssueUpdate.Issue, nil
}

// IssueDelete deletes an issue by ID.
func (c *Client) IssueDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteIssue(ctx, id)
	if err != nil {
		return fmt.Errorf("issue delete failed: %w", err)
	}

	if !resp.IssueDelete.Success {
		return fmt.Errorf("issue delete failed: success=false")
	}

	return nil
}

// Organization retrieves the organization information.
func (c *Client) Organization(ctx context.Context) (*intgraphql.GetOrganization_Organization, error) {
	resp, err := c.gqlClient.GetOrganization(ctx)
	if err != nil {
		return nil, fmt.Errorf("organization query failed: %w", err)
	}
	return &resp.Organization, nil
}

// IssueLabel retrieves a single issue label by ID.
func (c *Client) IssueLabel(ctx context.Context, id string) (*intgraphql.GetLabel_IssueLabel, error) {
	resp, err := c.gqlClient.GetLabel(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("label query failed: %w", err)
	}
	return &resp.IssueLabel, nil
}

// IssueLabels retrieves a paginated list of issue labels.
func (c *Client) IssueLabels(ctx context.Context, first *int64, after *string) (*intgraphql.ListLabels_IssueLabels, error) {
	resp, err := c.gqlClient.ListLabels(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("labels query failed: %w", err)
	}
	return &resp.IssueLabels, nil
}

// WorkflowState retrieves a single workflow state by ID.
func (c *Client) WorkflowState(ctx context.Context, id string) (*intgraphql.GetWorkflowState_WorkflowState, error) {
	resp, err := c.gqlClient.GetWorkflowState(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("workflow state query failed: %w", err)
	}
	return &resp.WorkflowState, nil
}

// WorkflowStates retrieves a paginated list of workflow states.
func (c *Client) WorkflowStates(ctx context.Context, first *int64, after *string) (*intgraphql.ListWorkflowStates_WorkflowStates, error) {
	resp, err := c.gqlClient.ListWorkflowStates(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("workflow states query failed: %w", err)
	}
	return &resp.WorkflowStates, nil
}

// Cycle retrieves a single cycle by ID.
func (c *Client) Cycle(ctx context.Context, id string) (*intgraphql.GetCycle_Cycle, error) {
	resp, err := c.gqlClient.GetCycle(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cycle query failed: %w", err)
	}
	return &resp.Cycle, nil
}

// Cycles retrieves a paginated list of cycles.
func (c *Client) Cycles(ctx context.Context, first *int64, after *string) (*intgraphql.ListCycles_Cycles, error) {
	resp, err := c.gqlClient.ListCycles(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("cycles query failed: %w", err)
	}
	return &resp.Cycles, nil
}

// Roadmap retrieves a single roadmap by ID.
func (c *Client) Roadmap(ctx context.Context, id string) (*intgraphql.GetRoadmap_Roadmap, error) {
	resp, err := c.gqlClient.GetRoadmap(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("roadmap query failed: %w", err)
	}
	return &resp.Roadmap, nil
}

// Roadmaps retrieves a paginated list of roadmaps.
func (c *Client) Roadmaps(ctx context.Context, first *int64, after *string) (*intgraphql.ListRoadmaps_Roadmaps, error) {
	resp, err := c.gqlClient.ListRoadmaps(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("roadmaps query failed: %w", err)
	}
	return &resp.Roadmaps, nil
}

// Attachment retrieves a single attachment by ID.
func (c *Client) Attachment(ctx context.Context, id string) (*intgraphql.GetAttachment_Attachment, error) {
	resp, err := c.gqlClient.GetAttachment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("attachment query failed: %w", err)
	}
	return &resp.Attachment, nil
}

// Attachments retrieves a paginated list of attachments.
func (c *Client) Attachments(ctx context.Context, first *int64, after *string) (*intgraphql.ListAttachments_Attachments, error) {
	resp, err := c.gqlClient.ListAttachments(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("attachments query failed: %w", err)
	}
	return &resp.Attachments, nil
}

// Initiative retrieves a single initiative by ID.
func (c *Client) Initiative(ctx context.Context, id string) (*intgraphql.GetInitiative_Initiative, error) {
	resp, err := c.gqlClient.GetInitiative(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("initiative query failed: %w", err)
	}
	return &resp.Initiative, nil
}

// Initiatives retrieves a paginated list of initiatives.
func (c *Client) Initiatives(ctx context.Context, first *int64, after *string) (*intgraphql.ListInitiatives_Initiatives, error) {
	resp, err := c.gqlClient.ListInitiatives(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("initiatives query failed: %w", err)
	}
	return &resp.Initiatives, nil
}

// CommentCreate creates a new comment.
func (c *Client) CommentCreate(ctx context.Context, input intgraphql.CommentCreateInput) (*intgraphql.CreateComment_CommentCreate_Comment, error) {
	resp, err := c.gqlClient.CreateComment(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("comment create failed: %w", err)
	}

	if !resp.CommentCreate.Success {
		return nil, fmt.Errorf("comment create failed: success=false")
	}

	return &resp.CommentCreate.Comment, nil
}

// CommentUpdate updates an existing comment.
func (c *Client) CommentUpdate(ctx context.Context, id string, input intgraphql.CommentUpdateInput) (*intgraphql.UpdateComment_CommentUpdate_Comment, error) {
	resp, err := c.gqlClient.UpdateComment(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("comment update failed: %w", err)
	}

	if !resp.CommentUpdate.Success {
		return nil, fmt.Errorf("comment update failed: success=false")
	}

	return &resp.CommentUpdate.Comment, nil
}

// CommentDelete deletes a comment by ID.
func (c *Client) CommentDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteComment(ctx, id)
	if err != nil {
		return fmt.Errorf("comment delete failed: %w", err)
	}

	if !resp.CommentDelete.Success {
		return fmt.Errorf("comment delete failed: success=false")
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

// Document retrieves a single document by ID.
func (c *Client) Document(ctx context.Context, id string) (*intgraphql.GetDocument_Document, error) {
	resp, err := c.gqlClient.GetDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("document query failed: %w", err)
	}
	return &resp.Document, nil
}

// Documents retrieves a paginated list of documents.
func (c *Client) Documents(ctx context.Context, first *int64, after *string) (*intgraphql.ListDocuments_Documents, error) {
	resp, err := c.gqlClient.ListDocuments(ctx, first, after)
	if err != nil {
		return nil, fmt.Errorf("documents query failed: %w", err)
	}
	return &resp.Documents, nil
}

// Template retrieves a single template by ID.
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
		return nil, fmt.Errorf("label create failed: %w", err)
	}

	if !resp.IssueLabelCreate.Success {
		return nil, fmt.Errorf("label create failed: success=false")
	}

	return &resp.IssueLabelCreate.IssueLabel, nil
}

// IssueLabelUpdate updates an existing label.
func (c *Client) IssueLabelUpdate(ctx context.Context, id string, input intgraphql.IssueLabelUpdateInput) (*intgraphql.UpdateLabel_IssueLabelUpdate_IssueLabel, error) {
	resp, err := c.gqlClient.UpdateLabel(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("label update failed: %w", err)
	}

	if !resp.IssueLabelUpdate.Success {
		return nil, fmt.Errorf("label update failed: success=false")
	}

	return &resp.IssueLabelUpdate.IssueLabel, nil
}

// IssueLabelDelete deletes a label by ID.
func (c *Client) IssueLabelDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteLabel(ctx, id)
	if err != nil {
		return fmt.Errorf("label delete failed: %w", err)
	}

	if !resp.IssueLabelDelete.Success {
		return fmt.Errorf("label delete failed: success=false")
	}

	return nil
}
