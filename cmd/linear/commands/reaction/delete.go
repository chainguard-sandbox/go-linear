package reaction

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewDeleteCommand creates the reaction delete command.
func NewDeleteCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <reaction-id>",
		Short: "Remove an emoji reaction",
		Long: `Remove an emoji reaction from an issue or comment.

This operation is safe and can be reversed by re-adding the reaction.

Parameters:
  <reaction-id>: Reaction UUID to delete (required)

Examples:
  # Remove a reaction
  linear reaction delete <reaction-uuid>

TIP: View reactions in Linear UI or via API to find reaction IDs

Related Commands:
  - linear reaction create - Add an emoji reaction`,
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
	ctx := context.Background()

	err := client.ReactionDelete(ctx, reactionID)
	if err != nil {
		return fmt.Errorf("failed to delete reaction: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed reaction\n")
	return nil
}
