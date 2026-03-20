package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewCreateCommand creates the project create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create project. Safe operation.

Required: --name, --team (from team_list)
Optional: --description, --lead (user), --member (user, repeatable)

Example: go-linear project create --name="Q1 Platform" --team=ENG --lead=me --member=john@co.com

Related: project_list, project_get, team_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("name", "", "Project name (required)")
	_ = cmd.MarkFlagRequired("name")

	cmd.Flags().String("team", "", "Team name or ID (required)")
	_ = cmd.MarkFlagRequired("team")

	cmd.Flags().String("description", "", "Project description")
	cmd.Flags().String("lead", "", "Project lead (user name, email, or ID)")
	cmd.Flags().StringArray("member", []string{}, "Project members (user name, email, or ID - repeatable)")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	// Resolve team to UUID
	res := resolver.New(client)
	team, _ := cmd.Flags().GetString("team")
	teamID, err := res.ResolveTeam(ctx, team)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	name, _ := cmd.Flags().GetString("name")
	input := intgraphql.ProjectCreateInput{
		Name:    name,
		TeamIds: []string{teamID},
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	if lead, _ := cmd.Flags().GetString("lead"); lead != "" {
		leadID, err := res.ResolveUser(ctx, lead)
		if err != nil {
			return fmt.Errorf("failed to resolve lead: %w", err)
		}
		input.LeadID = &leadID
	}

	members, _ := cmd.Flags().GetStringArray("member")
	if len(members) > 0 {
		memberIDs := make([]string, 0, len(members))
		for _, member := range members {
			memberID, err := res.ResolveUser(ctx, member)
			if err != nil {
				return fmt.Errorf("failed to resolve member %q: %w", member, err)
			}
			memberIDs = append(memberIDs, memberID)
		}
		input.MemberIds = memberIDs
	}

	result, err := client.ProjectCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
