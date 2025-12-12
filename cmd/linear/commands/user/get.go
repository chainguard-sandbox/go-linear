package user

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the user get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name|email|id>",
		Short: "Get a single user",
		Long: `Get detailed information about a specific user.

Accepts user name, email, UUID, or 'me' for the current authenticated user.

Parameters:
  <name|email|id>: User name, email, 'me', or UUID (required)

Output (--output=json):
  Returns JSON with: id, name, email, displayName, active, admin, avatarUrl

Examples:
  # Get current user
  linear user get me

  # Get user by email
  linear user get alice@company.com

  # Get with JSON output
  linear user get me --output=json

TIP: Use 'linear user list' to discover user names and emails

Related Commands:
  - linear user list - List all users
  - linear user completed - Get user's completed work`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "Comma-separated fields for JSON output (e.g., 'id,name,email')")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, nameOrEmailOrID string) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Resolve to user ID
	userID, err := res.ResolveUser(ctx, nameOrEmailOrID)
	if err != nil {
		return fmt.Errorf("failed to resolve user: %w", err)
	}

	user, err := client.User(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		fieldSelector, err := fieldfilter.New(fieldsSpec)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), user, true, fieldSelector)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name:   %s\n", user.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Email:  %s\n", user.Email)
		fmt.Fprintf(cmd.OutOrStdout(), "Active: %v\n", user.Active)
		if user.Admin {
			fmt.Fprintf(cmd.OutOrStdout(), "Admin:  Yes\n")
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
