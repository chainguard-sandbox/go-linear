// Package document provides document commands for the Linear CLI.
package document

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
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
		Long: `List documents. Returns 4 default fields per document.

Example: go-linear-cli document list --output=json

Related: document_get`,
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
				cfg, _ := config.Load()
				var configOverrides map[string]string
				if cfg != nil {
					configOverrides = cfg.FieldDefaults
				}
				defaults := fieldfilter.GetDefaults("document.list", configOverrides)
				fieldsSpec, _ := cmd.Flags().GetString("fields")
				fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
				if err != nil {
					return fmt.Errorf("invalid --fields: %w", err)
				}
				return formatter.FormatJSONFiltered(cmd.OutOrStdout(), documents, true, fieldSelector)
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
	cmd.Flags().String("fields", "", "defaults (id,title,content,createdAt) | none | defaults,extra")
	return cmd
}

func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single document by ID",
		Long: `Get document by UUID. Returns 4 default fields.

Example: go-linear-cli document get <uuid> --output=json

Related: document_list`,
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
			fieldsSpec, _ := cmd.Flags().GetString("fields")

			switch output {
			case "json":
				cfg, _ := config.Load()
				var configOverrides map[string]string
				if cfg != nil {
					configOverrides = cfg.FieldDefaults
				}
				defaults := fieldfilter.GetDefaults("document.get", configOverrides)
				fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
				if err != nil {
					return fmt.Errorf("invalid --fields: %w", err)
				}
				return formatter.FormatJSONFiltered(cmd.OutOrStdout(), document, true, fieldSelector)
			case "table":
				fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", document.Title)
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,title,content,createdAt) | none | defaults,extra")
	return cmd
}
