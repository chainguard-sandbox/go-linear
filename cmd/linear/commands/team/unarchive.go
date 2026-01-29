package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnarchiveCommand creates the team unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputOnlyFlags{}

	cmd := &cobra.Command{
		Use:   "unarchive <name-or-id>",
		Short: "Unarchive a team",
		Long: `Restore an archived team. Safe operation.

Example: go-linear team unarchive <uuid>
Example: go-linear team unarchive "Engineering"

Related: team_delete, team_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUnarchive(cmd, client, args[0], outputFlags)
		},
	}

	outputFlags.Bind(cmd)

	return cmd
}

func runUnarchive(cmd *cobra.Command, client *linear.Client, teamID string, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Resolve team ID
	resolvedID, err := res.ResolveTeam(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	err = client.TeamUnarchive(ctx, resolvedID)
	if err != nil {
		return fmt.Errorf("failed to unarchive team: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success": true,
			"teamId":  teamID,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Team %s unarchived successfully\n", teamID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
