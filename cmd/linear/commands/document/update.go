package document

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the document update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing document",
		Long: `Update knowledge base document. Modifies existing data.

Fields: --title, --content

Example: go-linear document update <uuid> --title="Updated API Guide" --output=json

Related: document_get, document_create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("title", "", "New document title")
	cmd.Flags().String("content", "", "New document content (markdown)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, documentID string) error {
	ctx := cmd.Context()

	input := intgraphql.DocumentUpdateInput{}
	updated := false

	if title, _ := cmd.Flags().GetString("title"); title != "" {
		input.Title = &title
		updated = true
	}

	if content, _ := cmd.Flags().GetString("content"); content != "" {
		input.Content = &content
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	result, err := client.DocumentUpdate(ctx, documentID, input)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated document: %s\n", result.Title)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
