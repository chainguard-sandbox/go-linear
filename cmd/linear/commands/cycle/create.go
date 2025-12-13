package cycle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cycle",
		Long: `Create cycle (sprint). Safe operation.

Required: --team (from team_list), --starts-at, --ends-at (date formats: see issue_list)
Optional: --name

Example: go-linear cycle create --team=ENG --starts-at=2025-12-16 --ends-at=14d --name="Sprint 42" --output=json

Related: cycle_list, team_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			res := resolver.New(client)
			parser := dateparser.New()

			teamName, _ := cmd.Flags().GetString("team")
			teamID, err := res.ResolveTeam(ctx, teamName)
			if err != nil {
				return fmt.Errorf("failed to resolve team: %w", err)
			}

			startsStr, _ := cmd.Flags().GetString("starts-at")
			starts, err := parser.Parse(startsStr)
			if err != nil {
				return fmt.Errorf("invalid starts-at date: %w", err)
			}

			endsStr, _ := cmd.Flags().GetString("ends-at")
			ends, err := parser.Parse(endsStr)
			if err != nil {
				return fmt.Errorf("invalid ends-at date: %w", err)
			}

			input := intgraphql.CycleCreateInput{
				TeamID:   teamID,
				StartsAt: starts,
				EndsAt:   ends,
			}

			if name, _ := cmd.Flags().GetString("name"); name != "" {
				input.Name = &name
			}

			result, err := client.CycleCreate(ctx, input)
			if err != nil {
				return fmt.Errorf("failed to create cycle: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
			}
			name := "cycle"
			if result.Name != nil {
				name = *result.Name
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Created cycle: %s\n", name)
			return nil
		},
	}

	cmd.Flags().String("team", "", "Team name or ID (required)")
	_ = cmd.MarkFlagRequired("team")
	cmd.Flags().String("starts-at", "", "Start date (required)")
	_ = cmd.MarkFlagRequired("starts-at")
	cmd.Flags().String("ends-at", "", "End date (required)")
	_ = cmd.MarkFlagRequired("ends-at")
	cmd.Flags().String("name", "", "Cycle name")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}
