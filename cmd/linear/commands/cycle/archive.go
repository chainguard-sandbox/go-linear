package cycle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

func NewArchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	cmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a cycle",
		Long: `Archive cycle. Hides from default views. Can be unarchived.

Example: go-linear cycle archive <uuid> --output=json

Related: cycle_list, cycle_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
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
	confirmFlags.Bind(cmd)
	return cmd
}
