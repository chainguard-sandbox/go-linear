package document

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnarchiveCommand creates the document unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputOnlyFlags{}

	cmd := &cobra.Command{
		Use:   "unarchive <id>",
		Short: "Unarchive a document",
		Long: `Restore a deleted document. Safe operation.

Example: go-linear document unarchive <uuid>

Related: document_delete, document_get`,
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

func runUnarchive(cmd *cobra.Command, client *linear.Client, documentID string, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	err := client.DocumentUnarchive(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to unarchive document: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success":    true,
			"documentId": documentID,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Document %s unarchived successfully\n", documentID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
