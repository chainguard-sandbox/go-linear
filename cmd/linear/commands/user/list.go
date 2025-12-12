package user

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the user list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all users",
		Long: `List all users in the Linear workspace.

Use this to discover user names and emails for assigning issues or filtering work.
Essential for finding assignees when creating or updating issues.

Output (--output=json):
  Returns JSON with:
  - nodes: Array of users
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each user contains:
  - id: User UUID
  - name: Full name
  - email: Email address (use for --assignee parameter)
  - displayName: Display name
  - active: Whether user is active in workspace

Examples:
  # List all users
  linear user list

  # JSON output for parameter discovery
  linear user list --output=json

TIP: Use email addresses or 'me' when assigning issues (e.g., --assignee=alice@company.com)

Related Commands:
  - linear user get - Get single user details
  - linear user completed - Get user's completed work
  - linear issue list --assignee=<email> - Filter issues by assignee`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 250, "Number of users to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "Comma-separated fields for JSON output (e.g., 'id,name,email')")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	users, err := client.Users(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		fieldSelector, err := fieldfilter.New(fieldsSpec)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), users, true, fieldSelector)
	case "table":
		return formatter.FormatUsersTable(cmd.OutOrStdout(), users.Nodes)
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
