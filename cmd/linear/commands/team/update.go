package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name|id>",
		Short: "Update an existing team",
		Long: `Update team. Modifies existing data.

Fields: --name, --description

Example: go-linear team update ENG --name="Platform Engineering" --description="Updated" --output=json

Related: team_get, team_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			res := resolver.New(client)

			teamID, err := res.ResolveTeam(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve team: %w", err)
			}

			input := intgraphql.TeamUpdateInput{}
			updated := false

			if name, _ := cmd.Flags().GetString("name"); name != "" {
				input.Name = &name
				updated = true
			}
			if desc, _ := cmd.Flags().GetString("description"); desc != "" {
				input.Description = &desc
				updated = true
			}

			if !updated {
				return fmt.Errorf("no fields to update specified")
			}

			result, err := client.TeamUpdate(ctx, teamID, input)
			if err != nil {
				return fmt.Errorf("failed to update team: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated team: %s\n", result.Name)
			return nil
		},
	}

	cmd.Flags().String("name", "", "New team name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}
