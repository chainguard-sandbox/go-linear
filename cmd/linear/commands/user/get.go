package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
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
		Long: `Get user by name, email, 'me', or UUID. Returns 7 default fields.

Example: go-linear user get me --output=json

Related: user_list, user_completed`,
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
	cmd.Flags().String("fields", "", "defaults (id,name,displayName,email,active,avatarUrl,admin) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, nameOrEmailOrID string) error {
	ctx := cmd.Context()
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
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("user.get", configOverrides)
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
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
