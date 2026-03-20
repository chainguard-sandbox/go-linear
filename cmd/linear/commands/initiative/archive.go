package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewArchiveCommand creates the initiative archive command.
func NewArchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
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

			return runArchive(cmd, client, args[0])
		},
	}

	return cmd
}

func runArchive(cmd *cobra.Command, client *linear.Client, initiativeID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve initiative ID
	resolvedID, err := res.ResolveInitiative(ctx, initiativeID)
	if err != nil {
		return fmt.Errorf("failed to resolve initiative: %w", err)
	}

	err = client.InitiativeArchive(ctx, resolvedID)
	if err != nil {
		return fmt.Errorf("failed to archive initiative: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success":      true,
		"initiativeId": initiativeID,
	}, true)
}
