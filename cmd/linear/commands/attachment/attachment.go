// Package attachment provides attachment commands for the Linear CLI.
package attachment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	attachmentfilter "github.com/chainguard-sandbox/go-linear/internal/filter/attachment"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func NewAttachmentCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Manage Linear attachments",
		Long:  "Commands for listing attachments and linking external resources.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewLinkURLCommand(clientFactory))
	cmd.AddCommand(NewLinkGitHubCommand(clientFactory))
	cmd.AddCommand(NewLinkSlackCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))

	return cmd
}

func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all attachments",
		Long: `List attachments with filtering. Returns 5 default fields per attachment.

Filters: --title, --url, --source-type, --creator
Date filters: --created-after, --created-before, --updated-after, --updated-before

Example: go-linear attachment list --source-type=github --limit=30 --output=json

Related: attachment_get, attachment_create, attachment_link-url`,
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
	cmd.Flags().String("id", "", "Attachment UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("source-type", "", "Source type: uploaded, url, github, slack")

	// Text filters
	cmd.Flags().String("title", "", "Title contains (case-insensitive)")
	cmd.Flags().String("subtitle", "", "Subtitle contains (case-insensitive)")
	cmd.Flags().String("url", "", "URL contains (case-insensitive)")

	// Output
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,title,url,source,createdAt) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := attachmentfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	attFilter := filterBuilder.Build()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	// Use filtered or unfiltered query based on whether filters were set
	if attFilter != nil {
		attachments, err := client.AttachmentsFiltered(ctx, &first, nil, attFilter)
		if err != nil {
			return fmt.Errorf("failed to list attachments: %w", err)
		}

		switch output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("attachment.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), attachments, true, fieldSelector)
		case "table":
			if len(attachments.Nodes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No attachments found")
				return nil
			}
			for _, att := range attachments.Nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", att.Title, att.URL)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output format: %s", output)
		}
	}

	// No filters: use regular query
	attachments, err := client.Attachments(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list attachments: %w", err)
	}

	switch output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("attachment.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), attachments, true, fieldSelector)
	case "table":
		if len(attachments.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No attachments found")
			return nil
		}
		for _, att := range attachments.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", att.Title, att.URL)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}

func NewLinkURLCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-url",
		Short: "Link an external URL to an issue",
		Long: `Link URL to issue. Safe operation.

Required: --issue (ID from issue_list), --url
Optional: --title

Example: go-linear attachment link-url --issue=ENG-123 --url=https://example.com/design --output=json

Related: attachment_link-github, attachment_link-slack, attachment_create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			issueID, _ := cmd.Flags().GetString("issue")
			url, _ := cmd.Flags().GetString("url")
			title, _ := cmd.Flags().GetString("title")

			var titlePtr *string
			if title != "" {
				titlePtr = &title
			}

			result, err := client.AttachmentLinkURL(ctx, issueID, url, titlePtr)
			if err != nil {
				return fmt.Errorf("failed to link URL: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Linked URL to issue\n")
			return nil
		},
	}

	cmd.Flags().String("issue", "", "Issue ID (required)")
	_ = cmd.MarkFlagRequired("issue")
	cmd.Flags().String("url", "", "URL to link (required)")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().String("title", "", "Link title")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func NewLinkGitHubCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-github",
		Short: "Link a GitHub PR to an issue",
		Long: `Link GitHub PR to issue. Safe operation.

Required: --issue (ID from issue_list), --url (GitHub PR URL like https://github.com/owner/repo/pull/123)

Example: go-linear attachment link-github --issue=ENG-123 --url=https://github.com/owner/repo/pull/123 --output=json

Related: attachment_link-url, attachment_link-slack, issue_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			issueID, _ := cmd.Flags().GetString("issue")
			url, _ := cmd.Flags().GetString("url")

			result, err := client.AttachmentLinkGitHubPR(ctx, issueID, url)
			if err != nil {
				return fmt.Errorf("failed to link GitHub PR: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Linked GitHub PR to issue\n")
			return nil
		},
	}

	cmd.Flags().String("issue", "", "Issue ID (required)")
	_ = cmd.MarkFlagRequired("issue")
	cmd.Flags().String("url", "", "GitHub PR URL (required)")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}
