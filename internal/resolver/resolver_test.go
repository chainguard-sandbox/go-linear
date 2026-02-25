//go:build read

package resolver

import (
	"context"
	"os"
	"testing"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// testSetup creates a client, resolver, and cleanup function for tests.
func testSetup(t *testing.T) (*linear.Client, *Resolver, func()) {
	t.Helper()

	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("LINEAR_API_KEY not set")
	}

	client, err := linear.NewClient(apiKey)
	if err != nil {
		t.Fatal(err)
	}

	r := New(client)

	// Clear cache to avoid race conditions in parallel tests
	r.cache.Clear()

	cleanup := func() {
		r.cache.Clear()
		client.Close()
	}

	return client, r, cleanup
}

func TestLive_ResolveTeam(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get first team from workspace for dynamic testing
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil || len(teams.Nodes) == 0 {
		t.Skip("No teams available for testing")
	}

	firstTeam := teams.Nodes[0]

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "resolve by team name",
			input:   firstTeam.Name,
			wantErr: false,
		},
		{
			name:    "resolve by team key",
			input:   firstTeam.Key,
			wantErr: false,
		},
		{
			name:    "resolve by UUID passthrough",
			input:   firstTeam.ID,
			wantErr: false,
		},
		{
			name:    "nonexistent team",
			input:   "nonexistent-team-xyz-12345",
			wantErr: true,
		},
		{
			name:    "empty team name",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveTeam(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id == "" {
				t.Error("ResolveTeam() returned empty ID")
			}

			// Test cache hit on second call
			if !tt.wantErr {
				id2, err2 := r.ResolveTeam(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveTeam() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveTeam() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveUser(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get current user for dynamic testing
	viewer, err := client.Viewer(ctx)
	if err != nil {
		t.Fatalf("Failed to get viewer: %v", err)
	}

	// Get all users to find a test user
	first := int64(10)
	users, err := client.Users(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}

	var testUser struct {
		ID          string
		Name        string
		Email       string
		DisplayName string
	}
	if len(users.Nodes) > 0 {
		u := users.Nodes[0]
		testUser.ID = u.ID
		testUser.Name = u.Name
		testUser.Email = u.Email
		testUser.DisplayName = u.DisplayName
	}

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve 'me' special case",
			input:   "me",
			wantID:  viewer.ID,
			wantErr: false,
		},
		{
			name:    "resolve 'ME' case insensitive",
			input:   "ME",
			wantID:  viewer.ID,
			wantErr: false,
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
		},
		{
			name:    "empty user",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonexistent user",
			input:   "nonexistent-user-xyz-99999@example.invalid",
			wantErr: true,
		},
	}

	// Add user-based tests if we have a test user
	if testUser.Name != "" {
		tests = append(tests, struct {
			name    string
			input   string
			wantID  string
			wantErr bool
		}{
			name:    "resolve by name",
			input:   testUser.Name,
			wantID:  testUser.ID,
			wantErr: false,
		})
	}
	if testUser.Email != "" {
		tests = append(tests, struct {
			name    string
			input   string
			wantID  string
			wantErr bool
		}{
			name:    "resolve by email",
			input:   testUser.Email,
			wantID:  testUser.ID,
			wantErr: false,
		})
	}
	if testUser.DisplayName != "" {
		tests = append(tests, struct {
			name    string
			input   string
			wantID  string
			wantErr bool
		}{
			name:    "resolve by display name",
			input:   testUser.DisplayName,
			wantID:  testUser.ID,
			wantErr: false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveUser(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveUser() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveUser() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit (except for "me" which doesn't use cache)
			if !tt.wantErr && tt.input != "me" && tt.input != "ME" {
				id2, err2 := r.ResolveUser(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveUser() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveUser() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveState(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get workflow states for dynamic testing
	// Find a state with a unique name to avoid ambiguity across teams
	first := int64(200)
	states, err := client.WorkflowStates(ctx, &first, nil)
	if err != nil || len(states.Nodes) == 0 {
		t.Skip("No workflow states available for testing")
	}

	// Count occurrences of each state name
	nameCounts := make(map[string]int)
	for _, s := range states.Nodes {
		nameCounts[s.Name]++
	}

	// Find a state with a unique name (only appears once)
	var firstState *intgraphql.ListWorkflowStates_WorkflowStates_Nodes
	for _, s := range states.Nodes {
		if nameCounts[s.Name] == 1 {
			firstState = s
			break
		}
	}
	if firstState == nil {
		t.Skip("No unique state name found (all states are ambiguous across teams)")
	}

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve by state name",
			input:   firstState.Name,
			wantID:  firstState.ID,
			wantErr: false,
		},
		{
			name:    "resolve by state name case insensitive",
			input:   firstState.Name,
			wantErr: false,
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
		},
		{
			name:    "empty state",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonexistent state",
			input:   "NonexistentState99999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveState(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveState() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveState() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit
			if !tt.wantErr && !uuidRegex.MatchString(tt.input) {
				id2, err2 := r.ResolveState(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveState() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveState() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveLabel(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get labels for dynamic testing
	first := int64(50)
	labels, err := client.IssueLabels(ctx, &first, nil)
	if err != nil || len(labels.Nodes) == 0 {
		t.Skip("No labels available for testing")
	}

	firstLabel := labels.Nodes[0]

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve by label name",
			input:   firstLabel.Name,
			wantID:  firstLabel.ID,
			wantErr: false,
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
		},
		{
			name:    "empty label",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonexistent label",
			input:   "NonexistentLabel99999XYZ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveLabel(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveLabel() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveLabel() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit
			if !tt.wantErr && !uuidRegex.MatchString(tt.input) {
				id2, err2 := r.ResolveLabel(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveLabel() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveLabel() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveIssue(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get an issue for dynamic testing
	first := int64(1)
	issues, err := client.Issues(ctx, &first, nil)
	if err != nil || len(issues.Nodes) == 0 {
		t.Skip("No issues available for testing")
	}

	firstIssue := issues.Nodes[0]

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve by identifier",
			input:   firstIssue.Identifier,
			wantID:  firstIssue.ID,
			wantErr: false,
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
		},
		{
			name:    "empty issue",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonexistent issue",
			input:   "NONEXISTENT-99999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveIssue(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveIssue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveIssue() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveIssue() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit
			if !tt.wantErr && !uuidRegex.MatchString(tt.input) {
				id2, err2 := r.ResolveIssue(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveIssue() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveIssue() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveProject(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get projects for dynamic testing
	first := int64(10)
	projects, err := client.Projects(ctx, &first, nil)
	if err != nil || len(projects.Nodes) == 0 {
		t.Skip("No projects available for testing")
	}

	firstProject := projects.Nodes[0]

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve by project name",
			input:   firstProject.Name,
			wantID:  firstProject.ID,
			wantErr: false,
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
		},
		{
			name:    "empty project",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonexistent project",
			input:   "NonexistentProject99999XYZ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveProject(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveProject() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveProject() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit
			if !tt.wantErr && !uuidRegex.MatchString(tt.input) {
				id2, err2 := r.ResolveProject(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveProject() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveProject() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveCycle(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get cycles for dynamic testing
	first := int64(10)
	cycles, err := client.Cycles(ctx, &first, nil)
	if err != nil || len(cycles.Nodes) == 0 {
		t.Skip("No cycles available for testing")
	}

	firstCycle := cycles.Nodes[0]
	if firstCycle.Name == nil {
		t.Skip("First cycle has no name")
	}

	// Check if cycle name is unique to avoid ambiguous name errors
	cycleNameUnique := true
	for _, c := range cycles.Nodes {
		if c.ID != firstCycle.ID && c.Name != nil && *c.Name == *firstCycle.Name {
			cycleNameUnique = false
			break
		}
	}

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
		skip    bool
	}{
		{
			name:    "resolve by cycle name",
			input:   *firstCycle.Name,
			wantID:  firstCycle.ID,
			wantErr: false,
			skip:    !cycleNameUnique, // Skip if name is not unique
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
			skip:    false,
		},
		{
			name:    "empty cycle",
			input:   "",
			wantErr: true,
			skip:    false,
		},
		{
			name:    "nonexistent cycle",
			input:   "NonexistentCycle99999XYZ",
			wantErr: true,
			skip:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skipf("Skipping because cycle name is not unique")
			}
			id, err := r.ResolveCycle(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveCycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveCycle() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveCycle() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit
			if !tt.wantErr && !uuidRegex.MatchString(tt.input) {
				id2, err2 := r.ResolveCycle(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveCycle() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveCycle() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveInitiative(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get initiatives for dynamic testing
	first := int64(10)
	initiatives, err := client.Initiatives(ctx, &first, nil)
	if err != nil || len(initiatives.Nodes) == 0 {
		t.Skip("No initiatives available for testing")
	}

	firstInitiative := initiatives.Nodes[0]

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve by initiative name",
			input:   firstInitiative.Name,
			wantID:  firstInitiative.ID,
			wantErr: false,
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
		},
		{
			name:    "empty initiative",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonexistent initiative",
			input:   "NonexistentInitiative99999XYZ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveInitiative(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveInitiative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveInitiative() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveInitiative() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit
			if !tt.wantErr && !uuidRegex.MatchString(tt.input) {
				id2, err2 := r.ResolveInitiative(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveInitiative() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveInitiative() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

func TestLive_ResolveDocument(t *testing.T) {
	client, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Get documents for dynamic testing
	first := int64(10)
	documents, err := client.Documents(ctx, &first, nil)
	if err != nil || len(documents.Nodes) == 0 {
		t.Skip("No documents available for testing")
	}

	firstDocument := documents.Nodes[0]

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{
			name:    "resolve by document title",
			input:   firstDocument.Title,
			wantID:  firstDocument.ID,
			wantErr: false,
		},
		{
			name:    "uuid passthrough",
			input:   "12345678-1234-1234-1234-123456789abc",
			wantID:  "12345678-1234-1234-1234-123456789abc",
			wantErr: false,
		},
		{
			name:    "empty document",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonexistent document",
			input:   "NonexistentDocument99999XYZ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := r.ResolveDocument(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if id == "" {
					t.Error("ResolveDocument() returned empty ID")
				}
				if tt.wantID != "" && id != tt.wantID {
					t.Errorf("ResolveDocument() = %s, want %s", id, tt.wantID)
				}
			}

			// Test cache hit
			if !tt.wantErr && !uuidRegex.MatchString(tt.input) {
				id2, err2 := r.ResolveDocument(ctx, tt.input)
				if err2 != nil {
					t.Errorf("ResolveDocument() cached call error = %v", err2)
				}
				if id2 != id {
					t.Errorf("ResolveDocument() cached = %s, want %s", id2, id)
				}
			}
		})
	}
}

// TestResolver_CacheIntegration tests that resolver properly uses cache across methods.
func TestResolver_CacheIntegration(t *testing.T) {
	_, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	// Test that UUIDs are passed through without cache
	uuid := "12345678-1234-1234-1234-123456789abc"

	id, err := r.ResolveTeam(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveTeam(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveUser(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveUser(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveState(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveState(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveLabel(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveLabel(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveIssue(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveIssue(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveProject(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveProject(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveCycle(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveCycle(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveInitiative(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveInitiative(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}

	id, err = r.ResolveDocument(ctx, uuid)
	if err != nil || id != uuid {
		t.Errorf("ResolveDocument(uuid) = %s, %v; want %s, nil", id, err, uuid)
	}
}

// TestResolver_EmptyInputs tests all methods with empty inputs.
func TestResolver_EmptyInputs(t *testing.T) {
	_, r, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	tests := []struct {
		name   string
		method func(context.Context, string) (string, error)
	}{
		{"ResolveTeam", r.ResolveTeam},
		{"ResolveUser", r.ResolveUser},
		{"ResolveState", r.ResolveState},
		{"ResolveLabel", r.ResolveLabel},
		{"ResolveIssue", r.ResolveIssue},
		{"ResolveProject", r.ResolveProject},
		{"ResolveCycle", r.ResolveCycle},
		{"ResolveInitiative", r.ResolveInitiative},
		{"ResolveDocument", r.ResolveDocument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.method(ctx, "")
			if err == nil {
				t.Errorf("%s(\"\") should return error for empty input", tt.name)
			}
		})
	}
}
