package reaction

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewDeleteCommand creates the reaction delete command.
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <reaction-id>",
		Short: "Remove an emoji reaction",
		Long: `Remove emoji reaction. Safe operation.

Example: go-linear reaction delete <reaction-uuid>

Related: reaction_create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runDelete(cmd, client, args[0])
		},
	}

	return cmd
}

func runDelete(cmd *cobra.Command, client *linear.Client, reactionID string) error {
	ctx := cmd.Context()

	err := client.ReactionDelete(ctx, reactionID)
	if err != nil {
		return fmt.Errorf("failed to delete reaction: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed reaction\n")
	return nil
}
