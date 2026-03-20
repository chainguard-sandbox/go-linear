package favorite

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewDeleteCommand creates the favorite delete command.
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <favorite-id>",
		Short: "Unstar a favorited item",
		Long: `Unstar item. Safe operation.

Example: go-linear favorite delete <favorite-uuid>

Related: favorite_create`,
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

func runDelete(cmd *cobra.Command, client *linear.Client, favoriteID string) error {
	ctx := cmd.Context()

	err := client.FavoriteDelete(ctx, favoriteID)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed favorite\n")
	return nil
}
