// Package document provides document commands for the Linear CLI.
package document

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	docfilter "github.com/chainguard-sandbox/go-linear/internal/filter/document"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
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
		Short: "List documents with filtering",
		Long: `List documents with filtering. Returns 4 default fields per document.

Filters: --title, --creator, --project, --initiative, --issue, --slug-id
Date filters: --created-after, --updated-after, etc. (date formats: see issue_list)

Example: go-linear document list --created-after=30d --output=json

Related: document_get, issue_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	// Pagination
	cmd.Flags().IntP("limit", "l", 50, "Number to return")

	// Date filters
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")

	// Entity filters
	cmd.Flags().String("id", "", "Document UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("initiative", "", "Initiative name or UUID")
	cmd.Flags().String("project", "", "Project name or UUID")
	cmd.Flags().String("issue", "", "Issue identifier or UUID")

	// Text filters
	cmd.Flags().String("title", "", "Title contains (case-insensitive)")
	cmd.Flags().String("slug-id", "", "Document slug ID (exact match)")

	// Output
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,title,content,createdAt) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := docfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	docFilter := filterBuilder.Build()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	// Use filtered or unfiltered query based on whether filters were set
	if docFilter != nil {
		documents, err := client.DocumentsFiltered(ctx, &first, nil, docFilter)
		if err != nil {
			return fmt.Errorf("failed to list documents: %w", err)
		}

		switch output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("document.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), documents, true, fieldSelector)
		case "table":
			if len(documents.Nodes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No documents found")
				return nil
			}
			for _, doc := range documents.Nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", doc.Title)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output format: %s", output)
		}
	}

	// No filters: use regular query
	documents, err := client.Documents(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}

	switch output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("document.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), documents, true, fieldSelector)
	case "table":
		if len(documents.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No documents found")
			return nil
		}
		for _, doc := range documents.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", doc.Title)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}

func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single document by ID",
		Long: `Get document by UUID. Returns 4 default fields.

Example: go-linear document get <uuid> --output=json

Related: document_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
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
