//go:build read

package linear

import (
	"context"
	"os"
	"testing"
)

// TestLive_Users tests Users query against real Linear API.
func TestLive_Users(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(10)
	users, err := client.Users(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Users() error = %v", err)
	}

	if users == nil {
		t.Fatal("Users() returned nil")
	}

	t.Logf("Retrieved %d users", len(users.Nodes))
	for i, user := range users.Nodes {
		t.Logf("  [%d] %s (%s) active=%v", i+1, user.Name, user.Email, user.Active)
	}
}

// TestLive_User tests User query against real Linear API.
func TestLive_User(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a user ID from the list
	first := int64(1)
	users, err := client.Users(ctx, &first, nil)
	if err != nil || len(users.Nodes) == 0 {
		t.Skip("No users available for testing")
	}

	userID := users.Nodes[0].ID
	user, err := client.User(ctx, userID)
	if err != nil {
		t.Fatalf("User() error = %v", err)
	}

	if user == nil {
		t.Fatal("User() returned nil")
	}

	if user.ID != userID {
		t.Errorf("User().ID = %q, want %q", user.ID, userID)
	}

	t.Logf("Retrieved user: %s (%s)", user.Name, user.Email)
}

// TestLive_Organization tests Organization query against real Linear API.
func TestLive_Organization(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	org, err := client.Organization(ctx)
	if err != nil {
		t.Fatalf("Organization() error = %v", err)
	}

	if org == nil {
		t.Fatal("Organization() returned nil")
	}

	if org.ID == "" {
		t.Error("Organization().ID is empty")
	}

	if org.Name == "" {
		t.Error("Organization().Name is empty")
	}

	t.Logf("Organization: %s (key: %s)", org.Name, org.URLKey)
}

// TestLive_WorkflowStates tests WorkflowStates query against real Linear API.
func TestLive_WorkflowStates(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(50)
	states, err := client.WorkflowStates(ctx, &first, nil)
	if err != nil {
		t.Fatalf("WorkflowStates() error = %v", err)
	}

	if states == nil {
		t.Fatal("WorkflowStates() returned nil")
	}

	t.Logf("Retrieved %d workflow states", len(states.Nodes))
	for i, state := range states.Nodes {
		t.Logf("  [%d] %s (type: %s, color: %s)", i+1, state.Name, state.Type, state.Color)
	}
}

// TestLive_WorkflowState tests WorkflowState query against real Linear API.
func TestLive_WorkflowState(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a state ID from the list
	first := int64(1)
	states, err := client.WorkflowStates(ctx, &first, nil)
	if err != nil || len(states.Nodes) == 0 {
		t.Skip("No workflow states available for testing")
	}

	stateID := states.Nodes[0].ID
	state, err := client.WorkflowState(ctx, stateID)
	if err != nil {
		t.Fatalf("WorkflowState() error = %v", err)
	}

	if state == nil {
		t.Fatal("WorkflowState() returned nil")
	}

	if state.ID != stateID {
		t.Errorf("WorkflowState().ID = %q, want %q", state.ID, stateID)
	}

	t.Logf("Retrieved state: %s (type: %s)", state.Name, state.Type)
}

// TestLive_IssueLabels tests IssueLabels query against real Linear API.
func TestLive_IssueLabels(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(50)
	labels, err := client.IssueLabels(ctx, &first, nil)
	if err != nil {
		t.Fatalf("IssueLabels() error = %v", err)
	}

	if labels == nil {
		t.Fatal("IssueLabels() returned nil")
	}

	t.Logf("Retrieved %d labels", len(labels.Nodes))
	for i, label := range labels.Nodes {
		t.Logf("  [%d] %s (color: %s)", i+1, label.Name, label.Color)
	}
}

// TestLive_IssueLabel tests IssueLabel query against real Linear API.
func TestLive_IssueLabel(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a label ID from the list
	first := int64(1)
	labels, err := client.IssueLabels(ctx, &first, nil)
	if err != nil || len(labels.Nodes) == 0 {
		t.Skip("No labels available for testing")
	}

	labelID := labels.Nodes[0].ID
	label, err := client.IssueLabel(ctx, labelID)
	if err != nil {
		t.Fatalf("IssueLabel() error = %v", err)
	}

	if label == nil {
		t.Fatal("IssueLabel() returned nil")
	}

	if label.ID != labelID {
		t.Errorf("IssueLabel().ID = %q, want %q", label.ID, labelID)
	}

	t.Logf("Retrieved label: %s (color: %s)", label.Name, label.Color)
}

// TestLive_Cycles tests Cycles query against real Linear API.
func TestLive_Cycles(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(10)
	cycles, err := client.Cycles(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Cycles() error = %v", err)
	}

	if cycles == nil {
		t.Fatal("Cycles() returned nil")
	}

	t.Logf("Retrieved %d cycles", len(cycles.Nodes))
	for i, cycle := range cycles.Nodes {
		name := "(unnamed)"
		if cycle.Name != nil {
			name = *cycle.Name
		}
		t.Logf("  [%d] %s (number: %.0f)", i+1, name, cycle.Number)
	}
}

// TestLive_Cycle tests Cycle query against real Linear API.
func TestLive_Cycle(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a cycle ID from the list
	first := int64(1)
	cycles, err := client.Cycles(ctx, &first, nil)
	if err != nil || len(cycles.Nodes) == 0 {
		t.Skip("No cycles available for testing")
	}

	cycleID := cycles.Nodes[0].ID
	cycle, err := client.Cycle(ctx, cycleID)
	if err != nil {
		t.Fatalf("Cycle() error = %v", err)
	}

	if cycle == nil {
		t.Fatal("Cycle() returned nil")
	}

	if cycle.ID != cycleID {
		t.Errorf("Cycle().ID = %q, want %q", cycle.ID, cycleID)
	}

	name := "(unnamed)"
	if cycle.Name != nil {
		name = *cycle.Name
	}
	t.Logf("Retrieved cycle: %s (number: %.0f)", name, cycle.Number)
}

// TestLive_Documents tests Documents query against real Linear API.
func TestLive_Documents(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(10)
	docs, err := client.Documents(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Documents() error = %v", err)
	}

	if docs == nil {
		t.Fatal("Documents() returned nil")
	}

	t.Logf("Retrieved %d documents", len(docs.Nodes))
	for i, doc := range docs.Nodes {
		t.Logf("  [%d] %s", i+1, doc.Title)
	}
}

// TestLive_Document tests Document query against real Linear API.
func TestLive_Document(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a document ID from the list
	first := int64(1)
	docs, err := client.Documents(ctx, &first, nil)
	if err != nil || len(docs.Nodes) == 0 {
		t.Skip("No documents available for testing")
	}

	docID := docs.Nodes[0].ID
	doc, err := client.Document(ctx, docID)
	if err != nil {
		t.Fatalf("Document() error = %v", err)
	}

	if doc == nil {
		t.Fatal("Document() returned nil")
	}

	if doc.ID != docID {
		t.Errorf("Document().ID = %q, want %q", doc.ID, docID)
	}

	t.Logf("Retrieved document: %s", doc.Title)
}

// TestLive_Roadmaps tests Roadmaps query against real Linear API.
func TestLive_Roadmaps(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(10)
	roadmaps, err := client.Roadmaps(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Roadmaps() error = %v", err)
	}

	if roadmaps == nil {
		t.Fatal("Roadmaps() returned nil")
	}

	t.Logf("Retrieved %d roadmaps", len(roadmaps.Nodes))
	for i, rm := range roadmaps.Nodes {
		t.Logf("  [%d] %s", i+1, rm.Name)
	}
}

// TestLive_Roadmap tests Roadmap query against real Linear API.
func TestLive_Roadmap(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a roadmap ID from the list
	first := int64(1)
	roadmaps, err := client.Roadmaps(ctx, &first, nil)
	if err != nil || len(roadmaps.Nodes) == 0 {
		t.Skip("No roadmaps available for testing")
	}

	rmID := roadmaps.Nodes[0].ID
	rm, err := client.Roadmap(ctx, rmID)
	if err != nil {
		t.Fatalf("Roadmap() error = %v", err)
	}

	if rm == nil {
		t.Fatal("Roadmap() returned nil")
	}

	if rm.ID != rmID {
		t.Errorf("Roadmap().ID = %q, want %q", rm.ID, rmID)
	}

	t.Logf("Retrieved roadmap: %s", rm.Name)
}

// TestLive_Templates tests Templates query against real Linear API.
func TestLive_Templates(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	templates, err := client.Templates(ctx)
	if err != nil {
		t.Fatalf("Templates() error = %v", err)
	}

	t.Logf("Retrieved %d templates", len(templates))
	for i, tmpl := range templates {
		t.Logf("  [%d] %s (type: %s)", i+1, tmpl.Name, tmpl.Type)
	}
}

// TestLive_Template tests Template query against real Linear API.
func TestLive_Template(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a template ID from the list
	templates, err := client.Templates(ctx)
	if err != nil || len(templates) == 0 {
		t.Skip("No templates available for testing")
	}

	tmplID := templates[0].ID
	tmpl, err := client.Template(ctx, tmplID)
	if err != nil {
		t.Fatalf("Template() error = %v", err)
	}

	if tmpl == nil {
		t.Fatal("Template() returned nil")
	}

	if tmpl.ID != tmplID {
		t.Errorf("Template().ID = %q, want %q", tmpl.ID, tmplID)
	}

	t.Logf("Retrieved template: %s (type: %s)", tmpl.Name, tmpl.Type)
}

// TestLive_Comments tests Comments query against real Linear API.
func TestLive_Comments(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(10)
	comments, err := client.Comments(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Comments() error = %v", err)
	}

	if comments == nil {
		t.Fatal("Comments() returned nil")
	}

	t.Logf("Retrieved %d comments", len(comments.Nodes))
	for i, c := range comments.Nodes {
		body := c.Body
		if len(body) > 50 {
			body = body[:50] + "..."
		}
		t.Logf("  [%d] %s", i+1, body)
	}
}

// TestLive_Comment tests Comment query against real Linear API.
func TestLive_Comment(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a comment ID from the list
	first := int64(1)
	comments, err := client.Comments(ctx, &first, nil)
	if err != nil || len(comments.Nodes) == 0 {
		t.Skip("No comments available for testing")
	}

	commentID := comments.Nodes[0].ID
	comment, err := client.Comment(ctx, commentID)
	if err != nil {
		t.Fatalf("Comment() error = %v", err)
	}

	if comment == nil {
		t.Fatal("Comment() returned nil")
	}

	if comment.ID != commentID {
		t.Errorf("Comment().ID = %q, want %q", comment.ID, commentID)
	}

	t.Logf("Retrieved comment: %s...", comment.Body[:min(50, len(comment.Body))])
}

// TestLive_Team tests Team query against real Linear API.
func TestLive_Team(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a team ID from the list
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil || len(teams.Nodes) == 0 {
		t.Skip("No teams available for testing")
	}

	teamID := teams.Nodes[0].ID
	team, err := client.Team(ctx, teamID)
	if err != nil {
		t.Fatalf("Team() error = %v", err)
	}

	if team == nil {
		t.Fatal("Team() returned nil")
	}

	if team.ID != teamID {
		t.Errorf("Team().ID = %q, want %q", team.ID, teamID)
	}

	t.Logf("Retrieved team: %s (%s)", team.Name, team.Key)
}

// TestLive_Project tests Project query against real Linear API.
func TestLive_Project(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First get a project ID from the list
	first := int64(1)
	projects, err := client.Projects(ctx, &first, nil)
	if err != nil || len(projects.Nodes) == 0 {
		t.Skip("No projects available for testing")
	}

	projectID := projects.Nodes[0].ID
	project, err := client.Project(ctx, projectID)
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}

	if project == nil {
		t.Fatal("Project() returned nil")
	}

	if project.ID != projectID {
		t.Errorf("Project().ID = %q, want %q", project.ID, projectID)
	}

	t.Logf("Retrieved project: %s (state: %s)", project.Name, project.State)
}
