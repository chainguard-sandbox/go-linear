package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewVelocityCommand creates the team velocity command.
func NewVelocityCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "velocity",
		Short: "Show team velocity metrics from recent cycles",
		Long: `Display velocity and capacity metrics for a team based on completed cycles.

Shows average points completed per cycle, issue throughput, and completion rates.
Uses the last 3-5 completed cycles by default.

Required: --team (name or key)
Optional: --cycles (number of recent cycles to analyze, default: 3)

Example: go-linear team velocity --team=ENG --output=table
Example: go-linear team velocity --team=ENG --cycles=5 --output=json

Related: cycle_get, cycle_list, team_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runVelocity(cmd, client)
		},
	}

	cmd.Flags().String("team", "", "Team name or key (required)")
	_ = cmd.MarkFlagRequired("team")
	cmd.Flags().Int("cycles", 3, "Number of recent completed cycles to analyze")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

type velocityMetrics struct {
	TeamID             string  `json:"teamId"`
	TeamName           string  `json:"teamName"`
	TeamKey            string  `json:"teamKey"`
	CyclesAnalyzed     int     `json:"cyclesAnalyzed"`
	AvgPointsCompleted float64 `json:"avgPointsCompleted"`
	AvgIssuesCompleted float64 `json:"avgIssuesCompleted"`
	AvgCompletionRate  float64 `json:"avgCompletionRate"`
	TotalPointsScope   float64 `json:"totalPointsScope"`
	TotalIssuesScope   float64 `json:"totalIssuesScope"`
}

func runVelocity(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve team
	teamInput, _ := cmd.Flags().GetString("team")
	teamID, err := res.ResolveTeam(ctx, teamInput)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	// Get team details
	team, err := client.Team(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	// Get number of cycles to analyze
	numCycles, _ := cmd.Flags().GetInt("cycles")
	numCycles = max(numCycles, 3)

	// Fetch recent cycles for this team (fetch extra to ensure we get enough completed ones)
	limit := int64(numCycles * 3)
	filter := &intgraphql.CycleFilter{
		Team: &intgraphql.TeamFilter{
			ID: &intgraphql.IDComparator{
				Eq: &teamID,
			},
		},
	}

	cycles, err := client.CyclesFiltered(ctx, &limit, nil, filter)
	if err != nil {
		return fmt.Errorf("failed to fetch cycles: %w", err)
	}

	// Filter to only completed cycles and limit to requested number
	completedCycles := make([]intgraphql.ListCyclesFiltered_Cycles_Nodes, 0, numCycles)
	for _, cycle := range cycles.Nodes {
		if cycle.CompletedAt != nil && len(completedCycles) < numCycles {
			completedCycles = append(completedCycles, *cycle)
		}
	}

	if len(completedCycles) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No completed cycles found for team %s\n", team.Name)
		return nil
	}

	// Calculate metrics
	var totalPointsCompleted, totalIssuesCompleted, totalPoints, totalIssues, totalCompletionRate float64

	for _, cycle := range completedCycles {
		if len(cycle.CompletedScopeHistory) > 0 {
			totalPointsCompleted += cycle.CompletedScopeHistory[len(cycle.CompletedScopeHistory)-1]
		}
		if len(cycle.ScopeHistory) > 0 {
			totalPoints += cycle.ScopeHistory[len(cycle.ScopeHistory)-1]
		}
		if len(cycle.CompletedIssueCountHistory) > 0 {
			totalIssuesCompleted += cycle.CompletedIssueCountHistory[len(cycle.CompletedIssueCountHistory)-1]
		}
		if len(cycle.IssueCountHistory) > 0 {
			totalIssues += cycle.IssueCountHistory[len(cycle.IssueCountHistory)-1]
		}
		totalCompletionRate += cycle.Progress
	}

	metrics := velocityMetrics{
		TeamID:             team.ID,
		TeamName:           team.Name,
		TeamKey:            team.Key,
		CyclesAnalyzed:     len(completedCycles),
		AvgPointsCompleted: totalPointsCompleted / float64(len(completedCycles)),
		AvgIssuesCompleted: totalIssuesCompleted / float64(len(completedCycles)),
		AvgCompletionRate:  totalCompletionRate / float64(len(completedCycles)),
		TotalPointsScope:   totalPoints,
		TotalIssuesScope:   totalIssues,
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), metrics, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Team Velocity Metrics: %s (%s)\n", metrics.TeamName, metrics.TeamKey)
		fmt.Fprintf(cmd.OutOrStdout(), "Cycles Analyzed: %d\n\n", metrics.CyclesAnalyzed)
		fmt.Fprintf(cmd.OutOrStdout(), "Average per Cycle:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Points Completed: %.1f\n", metrics.AvgPointsCompleted)
		fmt.Fprintf(cmd.OutOrStdout(), "  Issues Completed: %.1f\n", metrics.AvgIssuesCompleted)
		fmt.Fprintf(cmd.OutOrStdout(), "  Completion Rate:  %.1f%%\n", metrics.AvgCompletionRate*100)
		fmt.Fprintf(cmd.OutOrStdout(), "\nTotal Across %d Cycles:\n", metrics.CyclesAnalyzed)
		fmt.Fprintf(cmd.OutOrStdout(), "  Points Scope:     %.0f\n", metrics.TotalPointsScope)
		fmt.Fprintf(cmd.OutOrStdout(), "  Issues Scope:     %.0f\n", metrics.TotalIssuesScope)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
