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

// NewAddMemberCommand creates the team add-member command.
func NewAddMemberCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-member",
		Short: "Add a user to a team",
		Long: `Add user to team. Requires admin permissions.

Required: --team (name/key/UUID), --user (name/email/UUID)

Example: go-linear team add-member --team=ENG --user=alice@company.com --output=json

Related: team_remove-member, team_members`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runAddMember(cmd, client)
		},
	}

	cmd.Flags().String("team", "", "Team name, key, or UUID (required)")
	_ = cmd.MarkFlagRequired("team")

	cmd.Flags().String("user", "", "User name, email, or UUID (required)")
	_ = cmd.MarkFlagRequired("user")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runAddMember(cmd *cobra.Command, client *linear.Client) error {
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

	input := intgraphql.TeamMembershipCreateInput{
		TeamID: teamID,
		UserID: userID,
	}

	result, err := client.TeamMembershipCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Added %s to team %s\n",
			result.User.Name, result.Team.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
