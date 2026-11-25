//go:build read

package linear

import (
	"context"
	"os"
	"testing"
)

// TestLive_Planning runs each Planning query against the real Linear API. It
// logs counts only, never names.
func TestLive_Planning(t *testing.T) {
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

	projects, err := client.PlanningProjects(ctx, &first, nil)
	if err != nil {
		t.Fatalf("PlanningProjects() error = %v", err)
	}
	t.Logf("Retrieved %d projects", len(projects.Nodes))

	initiatives, err := client.PlanningInitiatives(ctx, &first, nil)
	if err != nil {
		t.Fatalf("PlanningInitiatives() error = %v", err)
	}
	t.Logf("Retrieved %d initiatives", len(initiatives.Nodes))

	teams, err := client.PlanningTeams(ctx, &first, nil)
	if err != nil {
		t.Fatalf("PlanningTeams() error = %v", err)
	}
	t.Logf("Retrieved %d teams", len(teams.Nodes))

	relations, err := client.PlanningProjectRelations(ctx, &first, nil)
	if err != nil {
		t.Fatalf("PlanningProjectRelations() error = %v", err)
	}
	t.Logf("Retrieved %d project relations", len(relations.Nodes))

	milestones, err := client.PlanningProjectMilestones(ctx, &first, nil)
	if err != nil {
		t.Fatalf("PlanningProjectMilestones() error = %v", err)
	}
	t.Logf("Retrieved %d project milestones", len(milestones.Nodes))
}
