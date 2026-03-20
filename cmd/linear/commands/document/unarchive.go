package document

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewUnarchiveCommand creates the document unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
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

			return runUnarchive(cmd, client, args[0])
		},
	}

	return cmd
}

func runUnarchive(cmd *cobra.Command, client *linear.Client, documentID string) error {
	ctx := cmd.Context()

	err := client.DocumentUnarchive(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to unarchive document: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success":    true,
		"documentId": documentID,
	}, true)
}
