package cycle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

func NewArchiveCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a cycle",
		Long: `Archive a cycle in Linear.

⚠️ Caution: This archives the cycle, hiding it from default views.
Archived cycles can be unarchived later.

Examples:
  linear cycle archive <uuid>
  linear cycle archive <uuid> --output=json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			err = client.CycleArchive(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to archive cycle: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Archived cycle\n")
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	return cmd
}
