// Package resolver provides name-to-ID resolution for Linear resources.
//
// Resolves friendly names (e.g., "Engineering", "alice@company.com") to UUIDs
// with caching to improve performance.
package resolver

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Resolver resolves names to IDs for Linear resources.
type Resolver struct {
	client *linear.Client
	cache  *Cache
}

// New creates a new resolver with a 5-minute cache TTL.
func New(client *linear.Client) *Resolver {
	return &Resolver{
		client: client,
		cache:  NewCache(5 * time.Minute),
	}
}

// entityMatcher defines how to fetch, match, and format entities for resolution.
type entityMatcher[T any] struct {
	cachePrefix string
	entityName  string
	fetch       func(ctx context.Context) ([]T, error)
	matches     func(entity T, query string) bool
	getID       func(entity T) string
	formatName  func(entity T) string
}

// resolve is a generic helper that implements the common resolution pattern:
// 1. Empty check
// 2. UUID passthrough
// 3. Cache lookup
// 4. Fetch entities
// 5. Find matches
// 6. Handle ambiguity
// 7. Cache and return
func resolve[T any](r *Resolver, ctx context.Context, nameOrID string, matcher entityMatcher[T]) (string, error) {
	// Empty check
	if nameOrID == "" {
		return "", fmt.Errorf("%s name/ID cannot be empty", matcher.entityName)
	}

	// UUID passthrough
	if uuidRegex.MatchString(nameOrID) {
		return nameOrID, nil
	}

	// Cache lookup
	cacheKey := matcher.cachePrefix + strings.ToLower(nameOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Fetch entities
	entities, err := matcher.fetch(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %ss: %w", matcher.entityName, err)
	}

	// Find matches
	var matchedIDs []string
	var matchedEntities []T
	for _, entity := range entities {
		if matcher.matches(entity, nameOrID) {
			matchedIDs = append(matchedIDs, matcher.getID(entity))
			matchedEntities = append(matchedEntities, entity)
		}
	}

	// Handle no matches
	if len(matchedIDs) == 0 {
		return "", fmt.Errorf("%s not found: %s", matcher.entityName, nameOrID)
	}

	// Handle ambiguity
	if len(matchedIDs) > 1 {
		names := make([]string, len(matchedEntities))
		for i, entity := range matchedEntities {
			names[i] = matcher.formatName(entity)
		}
		return "", fmt.Errorf("ambiguous %s name %q, matches: %s", matcher.entityName, nameOrID, strings.Join(names, ", "))
	}

	// Cache and return
	r.cache.Set(cacheKey, matchedIDs[0])
	return matchedIDs[0], nil
}

// ResolveTeam resolves a team name or key to its ID.
// Accepts: team name, team key, or UUID.
func (r *Resolver) ResolveTeam(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListTeams_Teams_Nodes]{
		cachePrefix: "team:",
		entityName:  "team",
		fetch: func(ctx context.Context) ([]*intgraphql.ListTeams_Teams_Nodes, error) {
			first := int64(100)
			resp, err := r.client.Teams(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(team *intgraphql.ListTeams_Teams_Nodes, query string) bool {
			return strings.EqualFold(team.Name, query) || strings.EqualFold(team.Key, query)
		},
		getID: func(team *intgraphql.ListTeams_Teams_Nodes) string { return team.ID },
		formatName: func(team *intgraphql.ListTeams_Teams_Nodes) string {
			return fmt.Sprintf("%s (%s)", team.Name, team.Key)
		},
	})
}

// ResolveUser resolves a user name, email, or display name to their ID.
// Accepts: full name, email, display name, "me" (for current user), or UUID.
func (r *Resolver) ResolveUser(ctx context.Context, nameOrEmailOrID string) (string, error) {
	// Handle "me" special case before generic resolution
	if strings.EqualFold(nameOrEmailOrID, "me") {
		viewer, err := r.client.Viewer(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		return viewer.ID, nil
	}

	return resolve(r, ctx, nameOrEmailOrID, entityMatcher[*intgraphql.ListUsers_Users_Nodes]{
		cachePrefix: "user:",
		entityName:  "user",
		fetch: func(ctx context.Context) ([]*intgraphql.ListUsers_Users_Nodes, error) {
			first := int64(250)
			resp, err := r.client.Users(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(user *intgraphql.ListUsers_Users_Nodes, query string) bool {
			return strings.EqualFold(user.Name, query) ||
				strings.EqualFold(user.Email, query) ||
				(user.DisplayName != "" && strings.EqualFold(user.DisplayName, query))
		},
		getID: func(user *intgraphql.ListUsers_Users_Nodes) string { return user.ID },
		formatName: func(user *intgraphql.ListUsers_Users_Nodes) string {
			return fmt.Sprintf("%s <%s>", user.Name, user.Email)
		},
	})
}

// ResolveState resolves a workflow state name to its ID.
// Accepts: state name or UUID.
func (r *Resolver) ResolveState(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListWorkflowStates_WorkflowStates_Nodes]{
		cachePrefix: "state:",
		entityName:  "workflow state",
		fetch: func(ctx context.Context) ([]*intgraphql.ListWorkflowStates_WorkflowStates_Nodes, error) {
			first := int64(100)
			resp, err := r.client.WorkflowStates(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(state *intgraphql.ListWorkflowStates_WorkflowStates_Nodes, query string) bool {
			return strings.EqualFold(state.Name, query)
		},
		getID:      func(state *intgraphql.ListWorkflowStates_WorkflowStates_Nodes) string { return state.ID },
		formatName: func(state *intgraphql.ListWorkflowStates_WorkflowStates_Nodes) string { return state.Name },
	})
}

// ResolveLabel resolves a label name to its ID.
// Accepts: label name or UUID.
func (r *Resolver) ResolveLabel(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListLabels_IssueLabels_Nodes]{
		cachePrefix: "label:",
		entityName:  "label",
		fetch: func(ctx context.Context) ([]*intgraphql.ListLabels_IssueLabels_Nodes, error) {
			first := int64(250)
			resp, err := r.client.IssueLabels(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(label *intgraphql.ListLabels_IssueLabels_Nodes, query string) bool {
			return strings.EqualFold(label.Name, query)
		},
		getID:      func(label *intgraphql.ListLabels_IssueLabels_Nodes) string { return label.ID },
		formatName: func(label *intgraphql.ListLabels_IssueLabels_Nodes) string { return label.Name },
	})
}

// ResolveIssue resolves an issue identifier or ID to its UUID.
// Accepts: issue identifier (e.g., "ENG-123"), issue number, or UUID.
func (r *Resolver) ResolveIssue(ctx context.Context, identifierOrID string) (string, error) {
	if identifierOrID == "" {
		return "", fmt.Errorf("issue identifier/ID cannot be empty")
	}

	// Check if already a UUID
	if uuidRegex.MatchString(identifierOrID) {
		return identifierOrID, nil
	}

	// Check cache
	cacheKey := "issue:" + strings.ToLower(identifierOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Search by identifier (ENG-123) or number
	first := int64(1)
	result, err := r.client.SearchIssues(ctx, identifierOrID, &first, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to search for issue %q: %w", identifierOrID, err)
	}

	if len(result.Nodes) == 0 {
		return "", fmt.Errorf("issue not found: %s", identifierOrID)
	}

	// Cache and return
	r.cache.Set(cacheKey, result.Nodes[0].ID)
	return result.Nodes[0].ID, nil
}

// ResolveProject resolves a project name to its ID.
// Accepts: project name or UUID.
func (r *Resolver) ResolveProject(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListProjects_Projects_Nodes]{
		cachePrefix: "project:",
		entityName:  "project",
		fetch: func(ctx context.Context) ([]*intgraphql.ListProjects_Projects_Nodes, error) {
			first := int64(100)
			resp, err := r.client.Projects(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(project *intgraphql.ListProjects_Projects_Nodes, query string) bool {
			return strings.EqualFold(project.Name, query)
		},
		getID:      func(project *intgraphql.ListProjects_Projects_Nodes) string { return project.ID },
		formatName: func(project *intgraphql.ListProjects_Projects_Nodes) string { return project.Name },
	})
}

// ResolveCycle resolves a cycle name to its ID.
// Accepts: cycle name or UUID.
func (r *Resolver) ResolveCycle(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListCycles_Cycles_Nodes]{
		cachePrefix: "cycle:",
		entityName:  "cycle",
		fetch: func(ctx context.Context) ([]*intgraphql.ListCycles_Cycles_Nodes, error) {
			first := int64(100)
			resp, err := r.client.Cycles(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(cycle *intgraphql.ListCycles_Cycles_Nodes, query string) bool {
			return cycle.Name != nil && strings.EqualFold(*cycle.Name, query)
		},
		getID: func(cycle *intgraphql.ListCycles_Cycles_Nodes) string { return cycle.ID },
		formatName: func(cycle *intgraphql.ListCycles_Cycles_Nodes) string {
			if cycle.Name != nil {
				return *cycle.Name
			}
			return cycle.ID
		},
	})
}

// ResolveInitiative resolves an initiative name to its ID.
// Accepts: initiative name or UUID.
func (r *Resolver) ResolveInitiative(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListInitiatives_Initiatives_Nodes]{
		cachePrefix: "initiative:",
		entityName:  "initiative",
		fetch: func(ctx context.Context) ([]*intgraphql.ListInitiatives_Initiatives_Nodes, error) {
			first := int64(100)
			resp, err := r.client.Initiatives(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(initiative *intgraphql.ListInitiatives_Initiatives_Nodes, query string) bool {
			return strings.EqualFold(initiative.Name, query)
		},
		getID:      func(initiative *intgraphql.ListInitiatives_Initiatives_Nodes) string { return initiative.ID },
		formatName: func(initiative *intgraphql.ListInitiatives_Initiatives_Nodes) string { return initiative.Name },
	})
}

// ResolveDocument resolves a document title to its ID.
// Accepts: document title or UUID.
func (r *Resolver) ResolveDocument(ctx context.Context, titleOrID string) (string, error) {
	return resolve(r, ctx, titleOrID, entityMatcher[*intgraphql.ListDocuments_Documents_Nodes]{
		cachePrefix: "document:",
		entityName:  "document",
		fetch: func(ctx context.Context) ([]*intgraphql.ListDocuments_Documents_Nodes, error) {
			first := int64(100)
			resp, err := r.client.Documents(ctx, &first, nil)
			if err != nil {
				return nil, err
			}
			return resp.Nodes, nil
		},
		matches: func(doc *intgraphql.ListDocuments_Documents_Nodes, query string) bool {
			return strings.EqualFold(doc.Title, query)
		},
		getID:      func(doc *intgraphql.ListDocuments_Documents_Nodes) string { return doc.ID },
		formatName: func(doc *intgraphql.ListDocuments_Documents_Nodes) string { return doc.Title },
	})
}

// ResolveTemplate resolves a template name to its ID.
// Accepts: template name or UUID.
func (r *Resolver) ResolveTemplate(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListTemplates_Templates]{
		cachePrefix: "template:",
		entityName:  "template",
		fetch: func(ctx context.Context) ([]*intgraphql.ListTemplates_Templates, error) {
			templates, err := r.client.Templates(ctx)
			if err != nil {
				return nil, err
			}
			return templates, nil
		},
		matches: func(template *intgraphql.ListTemplates_Templates, query string) bool {
			return strings.EqualFold(template.Name, query)
		},
		getID:      func(template *intgraphql.ListTemplates_Templates) string { return template.ID },
		formatName: func(template *intgraphql.ListTemplates_Templates) string { return template.Name },
	})
}
