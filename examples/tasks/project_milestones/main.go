// Package main demonstrates how to manage project milestones.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating a test project
// 2. Creating milestones (e.g., "Q1 2025", "Beta", "v1.0")
// 3. Updating milestone target dates
// 4. Deleting milestones
// 5. Cleaning up test data
//
// Milestones help break projects into phases:
// - Timeline-based: "Q1 2025", "Q2 2025"
// - Phase-based: "Alpha", "Beta", "GA"
// - Version-based: "v0.9", "v1.0", "v2.0"
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/project_milestones
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create client
	client, err := linear.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Step 1: Get team
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}
	teamID := teams.Nodes[0].ID

	// Step 2: Create test project
	fmt.Println("Creating test project...")
	projectName := "Test Project - Milestone Demo"
	projectDesc := "Demonstrates project milestone management"

	project, err := client.ProjectCreate(ctx, intgraphql.ProjectCreateInput{
		TeamIds:     []string{teamID},
		Name:        projectName,
		Description: &projectDesc,
	})
	if err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}
	fmt.Printf("Created project: %s (ID: %s)\n", project.Name, project.ID)
	fmt.Println()

	// Step 3: Create milestone 1 (Q1 2025)
	fmt.Println("Creating milestones...")
	m1Name := "Q1 2025"
	m1Desc := "First quarter deliverables"
	m1Date := "2025-03-31"

	milestone1, err := client.ProjectMilestoneCreate(ctx, intgraphql.ProjectMilestoneCreateInput{
		ProjectID:   project.ID,
		Name:        m1Name,
		Description: &m1Desc,
		TargetDate:  &m1Date,
	})
	if err != nil {
		log.Fatalf("Failed to create milestone 1: %v", err)
	}
	fmt.Printf("Created: %s (Target: %s)\n", milestone1.Name, *milestone1.TargetDate)

	// Step 4: Create milestone 2 (Beta Launch)
	m2Name := "Beta Launch"
	m2Date := "2025-06-15"

	milestone2, err := client.ProjectMilestoneCreate(ctx, intgraphql.ProjectMilestoneCreateInput{
		ProjectID:  project.ID,
		Name:       m2Name,
		TargetDate: &m2Date,
	})
	if err != nil {
		log.Fatalf("Failed to create milestone 2: %v", err)
	}
	fmt.Printf("Created: %s (Target: %s)\n", milestone2.Name, *milestone2.TargetDate)
	fmt.Println()

	// Step 5: Update milestone target date
	fmt.Printf("Updating '%s' target date...\n", milestone1.Name)
	updatedTargetDate := "2025-04-15"
	updated, err := client.ProjectMilestoneUpdate(ctx, milestone1.ID, intgraphql.ProjectMilestoneUpdateInput{
		TargetDate: &updatedTargetDate,
	})
	if err != nil {
		log.Fatalf("Failed to update milestone: %v", err)
	}
	if updated.TargetDate != nil {
		fmt.Printf("✓ Updated! New target: %s\n", *updated.TargetDate)
	}
	fmt.Println()

	// Step 6: Clean up test data
	fmt.Println("Cleaning up test data...")
	if err := client.ProjectMilestoneDelete(ctx, milestone1.ID); err != nil {
		log.Printf("Warning: Failed to delete milestone 1: %v", err)
	}
	if err := client.ProjectMilestoneDelete(ctx, milestone2.ID); err != nil {
		log.Printf("Warning: Failed to delete milestone 2: %v", err)
	}
	if err := client.ProjectDelete(ctx, project.ID); err != nil {
		log.Printf("Warning: Failed to delete project: %v", err)
	}
	fmt.Println("✓ Test data deleted")
	fmt.Println()

	fmt.Println("Project Milestone Features:")
	fmt.Println("  - Create milestones with target dates")
	fmt.Println("  - Update names, descriptions, dates")
	fmt.Println("  - Delete milestones (issues preserved)")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Timeline planning: Q1, Q2, Q3, Q4")
	fmt.Println("  - Release phases: Alpha, Beta, GA")
	fmt.Println("  - Version milestones: v0.9, v1.0, v2.0")
	fmt.Println("  - Project stages: Design, Development, Testing, Launch")
}
