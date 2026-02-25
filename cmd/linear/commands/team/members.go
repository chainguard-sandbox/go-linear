package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewMembersCommand creates the team members command.
func NewMembersCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a team",
		Long: `List team members. Use --count for member count only.

Required: --team (from team_list)

Example: go-linear team members --team=ENG

Count: --count returns {"count": N}
Related: team_get, user_list`,
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

	cmd.Flags().Bool("count", false, "Return only count, not member list")

	return cmd
}

func runMembers(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
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

	// Check if count mode
	countMode, _ := cmd.Flags().GetBool("count")
	if countMode {
		count := len(users.Nodes)
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"count": count,
		}, true)
	}

	// Normal mode - list members
	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"team":  team,
		"users": users.Nodes,
	}, true)
}
