// Package document provides document commands for the Linear CLI.
package document

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

type ClientFactory func() (*linear.Client, error)

func NewDocumentCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "document",
		Aliases: []string{"doc"},
		Short:   "Manage Linear documents",
		Long:    "Commands for listing and viewing Linear knowledge base documents.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))

	return cmd
}

func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all documents",
		Long: `List knowledge base documents from Linear.

Use this to:
- Browse available documentation and guides
- Find reference materials for your team
- Discover documents by title for linking

Output (--output=json):
  Returns JSON with:
  - nodes: Array of documents
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each document contains:
  - id: Document UUID
  - title: Document title
  - content: Markdown content
  - createdAt: Creation timestamp

Examples:
  # List all documents
  linear document list

  # List with limit
  linear document list --limit=20

  # JSON output for programmatic access
  linear document list --output=json

Related Commands:
  - linear document get - Get full document details including content`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			limit, _ := cmd.Flags().GetInt("limit")
			first := int64(limit)

			documents, err := client.Documents(ctx, &first, nil)
			if err != nil {
				return fmt.Errorf("failed to list documents: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				return formatter.FormatJSON(cmd.OutOrStdout(), documents, true)
			case "table":
				for _, doc := range documents.Nodes {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", doc.Title)
				}
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	return cmd
}

func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single document by ID",
		Long: `Get detailed information about a specific knowledge base document.

Retrieve full document details including title, content (markdown), and metadata.
Use this to access documentation, guides, or reference materials.

Parameters:
  <id>: Document UUID (required)

Output (--output=json):
  Returns JSON with: id, title, content, createdAt, updatedAt

Examples:
  # Get document by UUID
  linear document get <document-uuid>

  # Get with JSON output for parsing
  linear document get <document-uuid> --output=json

TIP: Use 'linear document list' to discover document IDs

Related Commands:
  - linear document list - List all documents`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			document, err := client.Document(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get document: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				return formatter.FormatJSON(cmd.OutOrStdout(), document, true)
			case "table":
				fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", document.Title)
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	return cmd
}
