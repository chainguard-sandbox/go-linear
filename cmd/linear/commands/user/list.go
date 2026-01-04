package user

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	userfilter "github.com/chainguard-sandbox/go-linear/internal/filter/user"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the user list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all users",
		Long: `List users with filtering. Returns 6 default fields per user. Use for discovering emails for --assignee.

Filters: --name, --email, --display-name, --active, --admin, --is-me
Date filters: --created-after, --created-before, --updated-after, --updated-before

Example: go-linear user list --active --output=json

Returns: {nodes: [{6 user fields}...], pageInfo: {hasNextPage, endCursor}}
Related: user_get, user_completed, issue_list`,
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
	cmd.Flags().IntP("limit", "l", 250, "Number of users to return")

	// Date filters
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")

	// Entity filters
	cmd.Flags().String("id", "", "User UUID")

	// Text filters
	cmd.Flags().String("name", "", "Name contains (case-insensitive)")
	cmd.Flags().String("email", "", "Email contains (case-insensitive)")
	cmd.Flags().String("display-name", "", "Display name contains (case-insensitive)")

	// Boolean filters
	cmd.Flags().Bool("active", false, "Filter by active status")
	cmd.Flags().Bool("admin", false, "Filter by admin status")
	cmd.Flags().Bool("is-me", false, "Filter for current user")

	// Output
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,displayName,email,active,avatarUrl) | none | defaults,extra | id,email,...")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	// Build filter from flags
	filterBuilder := userfilter.NewFilterBuilder(nil)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	usrFilter := filterBuilder.Build()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	// Use filtered or unfiltered query based on whether filters were set
	if usrFilter != nil {
		users, err := client.UsersFiltered(ctx, &first, nil, usrFilter)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		switch output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("user.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), users, true, fieldSelector)
		case "table":
			return formatter.FormatUsersTableFiltered(cmd.OutOrStdout(), users.Nodes)
		default:
			return fmt.Errorf("unsupported output format: %s", output)
		}
	}

	// No filters: use regular query
	users, err := client.Users(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	switch output {
	case "json":
		// Load config for field defaults
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}

		// Get command defaults
		defaults := fieldfilter.GetDefaults("user.list", configOverrides)

		// Parse field selector with defaults (list command preserves nodes/pageInfo)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
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
