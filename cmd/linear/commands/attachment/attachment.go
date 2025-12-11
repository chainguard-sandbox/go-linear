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
		Long: `List all file attachments and linked resources from Linear.

Use this to:
- Browse attachments across all issues
- Find linked GitHub PRs, Slack threads, or external URLs
- Audit attached resources for compliance or documentation

Attachments can be files, external links, GitHub PRs, or Slack conversations.

Output (--output=json):
  Returns JSON with:
  - nodes: Array of attachments
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each attachment contains:
  - id: Attachment UUID
  - title: Attachment title
  - url: Attachment URL
  - source: Source type (uploaded, url, github, slack)
  - issue: Associated issue reference

Examples:
  # List all attachments
  linear attachment list

  # List with limit
  linear attachment list --limit=30

  # JSON output for parsing
  linear attachment list --output=json

Related Commands:
  - linear attachment get - Get single attachment details
  - linear attachment create - Create custom attachment
  - linear attachment link-url - Link external URL`,
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
	return cmd
}

func NewLinkURLCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-url",
		Short: "Link an external URL to an issue",
		Long: `Link an external URL as an attachment to an issue.

This operation creates new attachment data and is safe to execute.
Use for linking design files, documentation, monitoring dashboards, or any external resource.

Parameters:
  --issue: Issue identifier (e.g., ENG-123) or UUID (required)
  --url: URL to link (required)
  --title: Link title (optional, defaults to page title if available)

Examples:
  # Link URL to issue
  linear attachment link-url --issue=ENG-123 --url=https://example.com/design

  # Link with custom title
  linear attachment link-url --issue=ENG-123 --url=https://example.com --title="Design Mockups"

  # Link with JSON output
  linear attachment link-url --issue=ENG-123 --url=https://example.com --output=json

TIP: URLs are automatically fetched to extract title and metadata

Related Commands:
  - linear attachment link-github - Link GitHub PR
  - linear attachment link-slack - Link Slack thread
  - linear attachment create - Create custom attachment with metadata`,
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
		Long: `Link a GitHub pull request as an attachment to an issue.

This operation creates new attachment data and is safe to execute.
Creates automatic sync between Linear issues and GitHub PRs for development tracking.

Parameters:
  --issue: Issue identifier (e.g., ENG-123) or UUID (required)
  --url: GitHub PR URL (required, e.g., https://github.com/owner/repo/pull/123)

Examples:
  # Link GitHub PR to issue
  linear attachment link-github --issue=ENG-123 --url=https://github.com/owner/repo/pull/123

  # Link PR with JSON output
  linear attachment link-github --issue=ENG-123 --url=<gh-pr-url> --output=json

TIP: GitHub integration automatically updates PR status in Linear

Common Errors:
  - "invalid URL": Must be a valid GitHub pull request URL
  - "integration not configured": Enable GitHub integration in Linear settings

Related Commands:
  - linear attachment link-url - Link generic external URL
  - linear attachment link-slack - Link Slack thread
  - linear attachment create - Create custom attachment`,
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
