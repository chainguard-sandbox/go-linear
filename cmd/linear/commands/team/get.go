package team

import (
	"context"
	"fmt"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// NewGetCommand creates the team get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name|id>",
		Short: "Get a single team",
		Long: `Get detailed information about a specific team.

Accepts team name, key, or ID.

Examples:
  linear team get Engineering
  linear team get ENG
  linear team get <uuid> --output=json`,
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
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), team, true)
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
