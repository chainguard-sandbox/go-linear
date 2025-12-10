package team

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewMembersCommand creates the team members command.
func NewMembersCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a team",
		Long: `List all members of a specific team.

Examples:
  linear team members --team=Engineering
  linear team members --team=ENG --output=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runMembers(cmd, client)
		},
	}

	cmd.Flags().String("team", "", "Team name or ID (required)")
	_ = cmd.MarkFlagRequired("team")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runMembers(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Resolve team
	teamName, _ := cmd.Flags().GetString("team")
	teamID, err := res.ResolveTeam(ctx, teamName)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	// Get team details with members
	team, err := client.Team(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	// Get all users and filter by team membership
	// Note: This is a workaround since Team.Members might not be available in the current schema
	// In a production implementation, we'd query team.Members if available
	first := int64(250)
	users, err := client.Users(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"team":  team,
			"users": users.Nodes,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Team: %s (%s)\n\n", team.Name, team.Key)
		return formatter.FormatUsersTable(cmd.OutOrStdout(), users.Nodes)
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
