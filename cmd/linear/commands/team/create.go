package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the team create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new team",
		Long: `Create team. Safe operation.

Required: --name, --key (2-5 uppercase letters, used in issue IDs like PLT-123)
Optional: --description

Example: go-linear team create --name=Platform --key=PLT --description="Platform team" --output=json

Related: team_list, team_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("name", "", "Team name (required)")
	_ = cmd.MarkFlagRequired("name")

	cmd.Flags().String("key", "", "Team key/identifier (required)")
	_ = cmd.MarkFlagRequired("key")

	cmd.Flags().String("description", "", "Team description")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	name, _ := cmd.Flags().GetString("name")
	key, _ := cmd.Flags().GetString("key")

	input := intgraphql.TeamCreateInput{
		Name: name,
		Key:  &key,
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	result, err := client.TeamCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created team: %s (%s)\n", result.Name, result.Key)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
