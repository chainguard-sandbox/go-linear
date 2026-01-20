package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnarchiveCommand creates the initiative unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputOnlyFlags{}

	cmd := &cobra.Command{
		Use:   "unarchive <name-or-id>",
		Short: "Unarchive an initiative",
		Long: `Restore an archived initiative. Safe operation.

Example: go-linear initiative unarchive <uuid>
Example: go-linear initiative unarchive "Q1 Goals"

Related: initiative_archive, initiative_get`,
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

func runUnarchive(cmd *cobra.Command, client *linear.Client, initiativeID string, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Resolve initiative ID
	resolvedID, err := res.ResolveInitiative(ctx, initiativeID)
	if err != nil {
		return fmt.Errorf("failed to resolve initiative: %w", err)
	}

	err = client.InitiativeUnarchive(ctx, resolvedID)
	if err != nil {
		return fmt.Errorf("failed to unarchive initiative: %w", err)
	}

	if outputFlags.Output == "json" {
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success":      true,
			"initiativeId": initiativeID,
		}, true)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Initiative %s unarchived successfully\n", initiativeID)
	return nil
}
