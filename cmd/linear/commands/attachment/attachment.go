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
	cmd.AddCommand(NewLinkURLCommand(clientFactory))
	cmd.AddCommand(NewLinkGitHubCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))

	return cmd
}

func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all attachments",
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
	cmd.MarkFlagRequired("issue")
	cmd.Flags().String("url", "", "URL to link (required)")
	cmd.MarkFlagRequired("url")
	cmd.Flags().String("title", "", "Link title")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func NewLinkGitHubCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-github",
		Short: "Link a GitHub PR to an issue",
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
	cmd.MarkFlagRequired("issue")
	cmd.Flags().String("url", "", "GitHub PR URL (required)")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}
