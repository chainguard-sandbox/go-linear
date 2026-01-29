package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the project update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project",
		Long: `Update project. Modifies existing data.

Fields: --name, --description

Example: go-linear project update <uuid> --name="New Name" --description="Updated" --output=json

Related: project_get, project_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("name", "", "New project name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("lead", "", "New project lead (user name, email, or ID)")
	cmd.Flags().StringArray("member", []string{}, "Set project members (replaces all current members)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, projectID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	input := intgraphql.ProjectUpdateInput{}
	updated := false

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		input.Name = &name
		updated = true
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
		updated = true
	}

	if lead, _ := cmd.Flags().GetString("lead"); lead != "" {
		leadID, err := res.ResolveUser(ctx, lead)
		if err != nil {
			return fmt.Errorf("failed to resolve lead: %w", err)
		}
		input.LeadID = &leadID
		updated = true
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
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	result, err := client.ProjectUpdate(ctx, projectID, input)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated project: %s\n", result.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
