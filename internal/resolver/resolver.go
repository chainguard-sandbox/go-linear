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
//
// Returns ResolutionError for user-facing errors with suggestions.
func resolve[T any](r *Resolver, ctx context.Context, nameOrID string, matcher entityMatcher[T]) (string, error) {
	// Empty check
	if nameOrID == "" {
		return "", &ResolutionError{
			EntityType: matcher.entityName,
			Input:      nameOrID,
			Reason:     "empty input",
			Internal:   fmt.Errorf("empty %s name/ID", matcher.entityName),
		}
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

	// Fetch entities - wrap errors without exposing internal details
	entities, err := matcher.fetch(ctx)
	if err != nil {
		return "", newFetchError(matcher.entityName, err)
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
		return "", newNotFoundError(matcher.entityName, nameOrID, nil)
	}

	// Handle ambiguity - collect names for internal logging only
	if len(matchedIDs) > 1 {
		names := make([]string, len(matchedEntities))
		for i, entity := range matchedEntities {
			names[i] = matcher.formatName(entity)
		}
		return "", newAmbiguousError(matcher.entityName, nameOrID, names)
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
			return "", newFetchError("user", err)
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

// stateTypeAliases maps common user-friendly names to workflow state types.
// State types are: triage, backlog, unstarted, started, completed, canceled
// Each alias maps to a single type to avoid ambiguity.
var stateTypeAliases = map[string]string{
	"todo":        "backlog", // Most common "todo" interpretation
	"to do":       "backlog",
	"to-do":       "backlog",
	"done":        "completed",
	"complete":    "completed",
	"finished":    "completed",
	"closed":      "completed",
	"canceled":    "canceled",
	"in progress": "started",
	"in-progress": "started",
	"wip":         "started",
	"active":      "started",
}

// ResolveState resolves a workflow state name to its ID.
// Accepts: state name, state type, common aliases (todo, done, etc.), or UUID.
// For aliases/types, only resolves if there's exactly one matching state.
// In multi-team workspaces with multiple matching states, returns an error
// with suggestions for disambiguation.
func (r *Resolver) ResolveState(ctx context.Context, nameOrID string) (string, error) {
	if nameOrID == "" {
		return "", &ResolutionError{
			EntityType: "workflow state",
			Input:      nameOrID,
			Reason:     "empty input",
			Internal:   fmt.Errorf("empty workflow state name/ID"),
		}
	}

	// UUID passthrough
	if uuidRegex.MatchString(nameOrID) {
		return nameOrID, nil
	}

	// Check cache
	cacheKey := "state:" + strings.ToLower(nameOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Fetch all states
	first := int64(100)
	resp, err := r.client.WorkflowStates(ctx, &first, nil)
	if err != nil {
		return "", newFetchError("workflow state", err)
	}

	queryLower := strings.ToLower(nameOrID)

	// Check for exact name match first (case-insensitive)
	// Collect all matches in case of duplicates across teams
	var nameMatches []*intgraphql.ListWorkflowStates_WorkflowStates_Nodes
	for _, state := range resp.Nodes {
		if strings.EqualFold(state.Name, nameOrID) {
			nameMatches = append(nameMatches, state)
		}
	}

	if len(nameMatches) == 1 {
		r.cache.Set(cacheKey, nameMatches[0].ID)
		return nameMatches[0].ID, nil
	}
	if len(nameMatches) > 1 {
		// Multiple states with same name across teams
		return "", &ResolutionError{
			EntityType: "workflow state",
			Input:      nameOrID,
			Reason:     fmt.Sprintf("ambiguous (%d teams have state '%s')", len(nameMatches), nameOrID),
			Suggestions: []string{
				"Use the state UUID directly",
				"Specify --team to disambiguate",
			},
			Internal: fmt.Errorf("state name %q exists in %d teams", nameOrID, len(nameMatches)),
		}
	}

	// Determine target type for alias/type-based lookup
	targetType := ""
	if t, ok := stateTypeAliases[queryLower]; ok {
		targetType = t
	} else {
		// Check if query is a valid state type directly
		validTypes := []string{"triage", "backlog", "unstarted", "started", "completed", "canceled"}
		for _, vt := range validTypes {
			if strings.EqualFold(nameOrID, vt) {
				targetType = vt
				break
			}
		}
	}

	// If we have a target type, find all matching states
	if targetType != "" {
		var typeMatches []*intgraphql.ListWorkflowStates_WorkflowStates_Nodes
		for _, state := range resp.Nodes {
			if strings.EqualFold(state.Type, targetType) {
				typeMatches = append(typeMatches, state)
			}
		}

		if len(typeMatches) == 1 {
			// Unambiguous - single team or unique state
			r.cache.Set(cacheKey, typeMatches[0].ID)
			return typeMatches[0].ID, nil
		}
		if len(typeMatches) > 1 {
			// Multiple states of same type - collect unique names for suggestions
			stateNames := make([]string, 0, len(typeMatches))
			seen := make(map[string]bool)
			for _, s := range typeMatches {
				if !seen[s.Name] {
					stateNames = append(stateNames, fmt.Sprintf("'%s'", s.Name))
					seen[s.Name] = true
				}
			}
			suggestions := []string{
				fmt.Sprintf("Use exact state name: %s", strings.Join(stateNames, ", ")),
				"Use the state UUID directly",
			}
			return "", &ResolutionError{
				EntityType:  "workflow state",
				Input:       nameOrID,
				Reason:      fmt.Sprintf("ambiguous (%d states of type '%s')", len(typeMatches), targetType),
				Suggestions: suggestions,
				Internal:    fmt.Errorf("alias %q matches %d states", nameOrID, len(typeMatches)),
			}
		}
	}

	return "", newNotFoundError("workflow state", nameOrID, nil)
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
		return "", &ResolutionError{
			EntityType: "issue",
			Input:      identifierOrID,
			Reason:     "empty input",
			Internal:   fmt.Errorf("empty issue identifier/ID"),
		}
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

	// Use direct issue lookup - Linear's issue(id:) query accepts both UUIDs and identifiers
	result, err := r.client.Issue(ctx, identifierOrID)
	if err != nil {
		return "", newFetchError("issue", err)
	}

	if result == nil {
		return "", newNotFoundError("issue", identifierOrID, nil)
	}

	// Cache and return
	r.cache.Set(cacheKey, result.ID)
	return result.ID, nil
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

// ResolveMilestone resolves a milestone ID.
// Currently only accepts UUIDs. Name-based resolution requires knowing the project.
// Use: go-linear project milestone-list <project> to find milestone UUIDs.
func (r *Resolver) ResolveMilestone(ctx context.Context, idOrName string) (string, error) {
	if idOrName == "" {
		return "", &ResolutionError{
			EntityType: "milestone",
			Input:      idOrName,
			Reason:     "empty input",
			Internal:   fmt.Errorf("empty milestone ID"),
		}
	}

	// UUID passthrough
	if uuidRegex.MatchString(idOrName) {
		return idOrName, nil
	}

	// Name-based resolution not yet supported
	return "", &ResolutionError{
		EntityType: "milestone",
		Input:      idOrName,
		Reason:     "name resolution not supported",
		Internal:   fmt.Errorf("milestone name resolution requires project context"),
	}
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

// ResolveProjectStatus resolves a project status name to its ID.
// Accepts: status name (e.g., "Backlog", "In Progress", "Completed") or UUID.
func (r *Resolver) ResolveProjectStatus(ctx context.Context, nameOrID string) (string, error) {
	return resolve(r, ctx, nameOrID, entityMatcher[*intgraphql.ListProjectStatuses_Organization_ProjectStatuses]{
		cachePrefix: "project_status:",
		entityName:  "project status",
		fetch: func(ctx context.Context) ([]*intgraphql.ListProjectStatuses_Organization_ProjectStatuses, error) {
			statuses, err := r.client.ProjectStatuses(ctx)
			if err != nil {
				return nil, err
			}
			return statuses, nil
		},
		matches: func(status *intgraphql.ListProjectStatuses_Organization_ProjectStatuses, query string) bool {
			return strings.EqualFold(status.Name, query)
		},
		getID:      func(status *intgraphql.ListProjectStatuses_Organization_ProjectStatuses) string { return status.ID },
		formatName: func(status *intgraphql.ListProjectStatuses_Organization_ProjectStatuses) string { return status.Name },
	})
}
