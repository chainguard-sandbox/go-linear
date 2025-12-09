//go:build read

package linear

import (
	"context"
	"os"
	"testing"
)

// TestLive_Teams tests Teams query against real Linear API.
func TestLive_Teams(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}

	if teams == nil {
		t.Fatal("Teams() returned nil")
	}

	t.Logf("Retrieved %d teams", len(teams.Nodes))
	for i, team := range teams.Nodes {
		t.Logf("  [%d] %s (%s)", i+1, team.Name, team.Key)
	}
}

// TestLive_Projects tests Projects query against real Linear API.
func TestLive_Projects(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	first := int64(10)
	projects, err := client.Projects(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Projects() error = %v", err)
	}

	if projects == nil {
		t.Fatal("Projects() returned nil")
	}

	t.Logf("Retrieved %d projects", len(projects.Nodes))
	for i, project := range projects.Nodes {
		t.Logf("  [%d] %s (state: %s)", i+1, project.Name, project.State)
	}
}
