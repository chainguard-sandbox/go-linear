// Package attachment provides attachment commands for the Linear CLI.
package attachment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

type ClientFactory func() (*linear.Client, error)

func NewAttachmentCommand(clientFactory ClientFactory) *cobra.Command {
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

func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all attachments",
		Long: `List attachments. Returns 5 default fields per attachment.

Example: go-linear-cli attachment list --limit=30 --output=json

Related: attachment_get, attachment_create, attachment_link-url`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			limit, _ := cmd.Flags().GetInt("limit")
			first := int64(limit)

			attachments, err := client.Attachments(ctx, &first, nil)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), attachments, true)
			}
			for _, att := range attachments.Nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", att.Title, att.URL)
			}
			return nil
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,title,url,source,createdAt) | none | defaults,extra")
	return cmd
}

func NewLinkURLCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-url",
		Short: "Link an external URL to an issue",
		Long: `Link URL to issue. Safe operation.

Required: --issue (ID from issue_list), --url
Optional: --title

Example: go-linear-cli attachment link-url --issue=ENG-123 --url=https://example.com/design --output=json

Related: attachment_link-github, attachment_link-slack, attachment_create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
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

func NewLinkGitHubCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-github",
		Short: "Link a GitHub PR to an issue",
		Long: `Link GitHub PR to issue. Safe operation.

Required: --issue (ID from issue_list), --url (GitHub PR URL like https://github.com/owner/repo/pull/123)

Example: go-linear-cli attachment link-github --issue=ENG-123 --url=https://github.com/owner/repo/pull/123 --output=json

Related: attachment_link-url, attachment_link-slack, issue_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
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
