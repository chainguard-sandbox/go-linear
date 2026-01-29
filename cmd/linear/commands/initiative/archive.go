package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewArchiveCommand creates the initiative archive command.
func NewArchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputOnlyFlags{}

	cmd := &cobra.Command{
		Use:   "archive <name-or-id>",
		Short: "Archive an initiative",
		Long: `Archive initiative. Hides from default views. Can be unarchived.

Example: go-linear initiative archive <uuid>
Example: go-linear initiative archive "Q1 Goals"

Related: initiative_unarchive, initiative_delete, initiative_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runArchive(cmd, client, args[0], outputFlags)
		},
	}

	outputFlags.Bind(cmd)

	return cmd
}

func runArchive(cmd *cobra.Command, client *linear.Client, initiativeID string, outputFlags *cli.OutputOnlyFlags) error {
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

	err = client.InitiativeArchive(ctx, resolvedID)
	if err != nil {
		return fmt.Errorf("failed to archive initiative: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success":      true,
			"initiativeId": initiativeID,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Initiative %s archived successfully\n", initiativeID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
