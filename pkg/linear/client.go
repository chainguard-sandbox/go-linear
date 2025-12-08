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
