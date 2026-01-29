package document

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewDeleteCommand creates the document delete command.
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	outputFlags := &cli.OutputOnlyFlags{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a document",
		Long: `⚠️ Delete knowledge base document. Cannot be undone. Prompts unless --yes.

Example: go-linear document delete <uuid>

Related: document_list, document_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runDelete(cmd, client, args[0], confirmFlags, outputFlags)
		},
	}

	outputFlags.Bind(cmd)
	confirmFlags.Bind(cmd)

	return cmd
}

func runDelete(cmd *cobra.Command, client *linear.Client, documentID string, confirmFlags *cli.ConfirmationFlags, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Confirmation
	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Delete document %s? This cannot be undone.\n", documentID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(response), "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	err := client.DocumentDelete(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Document deleted\n")
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
