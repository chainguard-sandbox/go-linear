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

// ResolveTeam resolves a team name or key to its ID.
// Accepts: team name, team key, or UUID.
func (r *Resolver) ResolveTeam(ctx context.Context, nameOrID string) (string, error) {
	if nameOrID == "" {
		return "", fmt.Errorf("team name/ID cannot be empty")
	}

	// Check if already a UUID
	if uuidRegex.MatchString(nameOrID) {
		return nameOrID, nil
	}

	// Check cache
	cacheKey := "team:" + strings.ToLower(nameOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Query API
	first := int64(100)
	teams, err := r.client.Teams(ctx, &first, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch teams: %w", err)
	}

	// Find by name or key (case-insensitive)
	var matches []*struct {
		ID   string
		Name string
		Key  string
	}

	// Case-insensitive comparison via EqualFold
	for _, team := range teams.Nodes {
		if strings.EqualFold(team.Name, nameOrID) || strings.EqualFold(team.Key, nameOrID) {
			matches = append(matches, &struct {
				ID   string
				Name string
				Key  string
			}{ID: team.ID, Name: team.Name, Key: team.Key})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("team not found: %s", nameOrID)
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = fmt.Sprintf("%s (%s)", m.Name, m.Key)
		}
		return "", fmt.Errorf("ambiguous team name %q, matches: %s", nameOrID, strings.Join(names, ", "))
	}

	// Cache and return
	r.cache.Set(cacheKey, matches[0].ID)
	return matches[0].ID, nil
}

// ResolveUser resolves a user name, email, or display name to their ID.
// Accepts: full name, email, display name, "me" (for current user), or UUID.
func (r *Resolver) ResolveUser(ctx context.Context, nameOrEmailOrID string) (string, error) {
	if nameOrEmailOrID == "" {
		return "", fmt.Errorf("user name/email/ID cannot be empty")
	}

	// Handle "me" special case
	if strings.EqualFold(nameOrEmailOrID, "me") {
		viewer, err := r.client.Viewer(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		return viewer.ID, nil
	}

	// Check if already a UUID
	if uuidRegex.MatchString(nameOrEmailOrID) {
		return nameOrEmailOrID, nil
	}

	// Check cache
	cacheKey := "user:" + strings.ToLower(nameOrEmailOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Query API
	first := int64(250)
	users, err := r.client.Users(ctx, &first, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch users: %w", err)
	}

	// Find by name, email, or display name (case-insensitive)
	var matches []*struct {
		ID    string
		Name  string
		Email string
	}

	// Case-insensitive comparison via EqualFold
	for _, user := range users.Nodes {
		if strings.EqualFold(user.Name, nameOrEmailOrID) ||
			strings.EqualFold(user.Email, nameOrEmailOrID) ||
			(user.DisplayName != "" && strings.EqualFold(user.DisplayName, nameOrEmailOrID)) {
			matches = append(matches, &struct {
				ID    string
				Name  string
				Email string
			}{ID: user.ID, Name: user.Name, Email: user.Email})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("user not found: %s", nameOrEmailOrID)
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = fmt.Sprintf("%s <%s>", m.Name, m.Email)
		}
		return "", fmt.Errorf("ambiguous user %q, matches: %s", nameOrEmailOrID, strings.Join(names, ", "))
	}

	// Cache and return
	r.cache.Set(cacheKey, matches[0].ID)
	return matches[0].ID, nil
}

// ResolveState resolves a workflow state name to its ID.
// Accepts: state name or UUID.
func (r *Resolver) ResolveState(ctx context.Context, nameOrID string) (string, error) {
	if nameOrID == "" {
		return "", fmt.Errorf("state name/ID cannot be empty")
	}

	// Check if already a UUID
	if uuidRegex.MatchString(nameOrID) {
		return nameOrID, nil
	}

	// Check cache
	cacheKey := "state:" + strings.ToLower(nameOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Query API
	first := int64(100)
	states, err := r.client.WorkflowStates(ctx, &first, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch workflow states: %w", err)
	}

	// Find by name (case-insensitive)
	var matches []*struct {
		ID   string
		Name string
	}

	// Case-insensitive comparison via EqualFold
	for _, state := range states.Nodes {
		if strings.EqualFold(state.Name, nameOrID) {
			matches = append(matches, &struct {
				ID   string
				Name string
			}{ID: state.ID, Name: state.Name})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("workflow state not found: %s", nameOrID)
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return "", fmt.Errorf("ambiguous state name %q, matches: %s", nameOrID, strings.Join(names, ", "))
	}

	// Cache and return
	r.cache.Set(cacheKey, matches[0].ID)
	return matches[0].ID, nil
}

// ResolveLabel resolves a label name to its ID.
// Accepts: label name or UUID.
func (r *Resolver) ResolveLabel(ctx context.Context, nameOrID string) (string, error) {
	if nameOrID == "" {
		return "", fmt.Errorf("label name/ID cannot be empty")
	}

	// Check if already a UUID
	if uuidRegex.MatchString(nameOrID) {
		return nameOrID, nil
	}

	// Check cache
	cacheKey := "label:" + strings.ToLower(nameOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Query API
	first := int64(250)
	labels, err := r.client.IssueLabels(ctx, &first, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch labels: %w", err)
	}

	// Find by name (case-insensitive)
	var matches []*struct {
		ID   string
		Name string
	}

	// Case-insensitive comparison via EqualFold
	for _, label := range labels.Nodes {
		if strings.EqualFold(label.Name, nameOrID) {
			matches = append(matches, &struct {
				ID   string
				Name string
			}{ID: label.ID, Name: label.Name})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("label not found: %s", nameOrID)
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return "", fmt.Errorf("ambiguous label name %q, matches: %s", nameOrID, strings.Join(names, ", "))
	}

	// Cache and return
	r.cache.Set(cacheKey, matches[0].ID)
	return matches[0].ID, nil
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
	if nameOrID == "" {
		return "", fmt.Errorf("project name/ID cannot be empty")
	}

	// Check if already a UUID
	if uuidRegex.MatchString(nameOrID) {
		return nameOrID, nil
	}

	// Check cache
	cacheKey := "project:" + strings.ToLower(nameOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Query API
	first := int64(100)
	projects, err := r.client.Projects(ctx, &first, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch projects: %w", err)
	}

	// Find by name (case-insensitive)
	var matches []*struct {
		ID   string
		Name string
	}

	// Case-insensitive comparison via EqualFold
	for _, project := range projects.Nodes {
		if strings.EqualFold(project.Name, nameOrID) {
			matches = append(matches, &struct {
				ID   string
				Name string
			}{ID: project.ID, Name: project.Name})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("project not found: %s", nameOrID)
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return "", fmt.Errorf("ambiguous project name %q, matches: %s", nameOrID, strings.Join(names, ", "))
	}

	// Cache and return
	r.cache.Set(cacheKey, matches[0].ID)
	return matches[0].ID, nil
}

// ResolveCycle resolves a cycle name to its ID.
// Accepts: cycle name or UUID.
func (r *Resolver) ResolveCycle(ctx context.Context, nameOrID string) (string, error) {
	if nameOrID == "" {
		return "", fmt.Errorf("cycle name/ID cannot be empty")
	}

	// Check if already a UUID
	if uuidRegex.MatchString(nameOrID) {
		return nameOrID, nil
	}

	// Check cache
	cacheKey := "cycle:" + strings.ToLower(nameOrID)
	if id, ok := r.cache.Get(cacheKey); ok {
		return id, nil
	}

	// Query API
	first := int64(100)
	cycles, err := r.client.Cycles(ctx, &first, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch cycles: %w", err)
	}

	// Find by name (case-insensitive)
	var matches []*struct {
		ID   string
		Name string
	}

	for _, cycle := range cycles.Nodes {
		if cycle.Name != nil && strings.EqualFold(*cycle.Name, nameOrID) {
			matches = append(matches, &struct {
				ID   string
				Name string
			}{ID: cycle.ID, Name: *cycle.Name})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("cycle not found: %s", nameOrID)
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return "", fmt.Errorf("ambiguous cycle name %q, matches: %s", nameOrID, strings.Join(names, ", "))
	}

	// Cache and return
	r.cache.Set(cacheKey, matches[0].ID)
	return matches[0].ID, nil
}
