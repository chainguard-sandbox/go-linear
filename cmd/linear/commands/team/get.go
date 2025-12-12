package team

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the team get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name|id>",
		Short: "Get a single team",
		Long: `Get detailed information about a specific team.

Accepts team name, key, or UUID for flexible querying.

Parameters:
  <name|id>: Team name, key (e.g., "ENG"), or UUID (required)

Output (--output=json):
  Returns JSON with: id, name, key, description, color, private

Examples:
  # Get team by name
  linear team get Engineering

  # Get team by key
  linear team get ENG

  # Get with JSON output
  linear team get Engineering --output=json

TIP: Use 'linear team list' to discover team names and keys

Related Commands:
  - linear team list - List all teams
  - linear team members - List team members`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "Comma-separated fields for JSON output (e.g., 'id,name,key')")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, nameOrID string) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Resolve to team ID
	teamID, err := res.ResolveTeam(ctx, nameOrID)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	team, err := client.Team(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		fieldSelector, err := fieldfilter.New(fieldsSpec)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), team, true, fieldSelector)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", team.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Key:  %s\n", team.Key)
		if team.Description != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", *team.Description)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
