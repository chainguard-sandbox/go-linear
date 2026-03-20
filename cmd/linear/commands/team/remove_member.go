package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewRemoveMemberCommand creates the team remove-member command.
func NewRemoveMemberCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-member",
		Short: "Remove a user from a team",
		Long: `Remove user from team. Requires admin permissions.

Required: --team (name/key/UUID), --user (name/email/UUID)

Example: go-linear team remove-member --team=ENG --user=alice@example.com

Related: team_add-member, team_members`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRemoveMember(cmd, client)
		},
	}

	cmd.Flags().String("team", "", "Team name, key, or UUID (required)")
	_ = cmd.MarkFlagRequired("team")

	cmd.Flags().String("user", "", "User name, email, or UUID (required)")
	_ = cmd.MarkFlagRequired("user")

	return cmd
}

func runRemoveMember(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve team
	teamInput, _ := cmd.Flags().GetString("team")
	teamID, err := res.ResolveTeam(ctx, teamInput)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	// Resolve user
	userInput, _ := cmd.Flags().GetString("user")
	userID, err := res.ResolveUser(ctx, userInput)
	if err != nil {
		return fmt.Errorf("failed to resolve user: %w", err)
	}

	// Get team memberships to find the membership ID
	first := int64(250)
	teamWithMemberships, err := client.TeamMemberships(ctx, teamID, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to get team memberships: %w", err)
	}

	// Find the membership for this user
	var membershipID string
	var userName string
	for _, membership := range teamWithMemberships.Memberships.Nodes {
		if membership.User.ID == userID {
			membershipID = membership.ID
			userName = membership.User.Name
			break
		}
	}

	if membershipID == "" {
		return fmt.Errorf("user %s is not a member of team %s", userInput, teamWithMemberships.Name)
	}

	// Delete the membership
	err = client.TeamMembershipDelete(ctx, membershipID)
	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success":      true,
		"membershipId": membershipID,
		"userId":       userID,
		"teamId":       teamID,
		"teamName":     teamWithMemberships.Name,
		"userName":     userName,
	}, true)
}
