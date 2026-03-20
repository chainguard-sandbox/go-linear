package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewGetCommand creates the user get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <name|email|id>",
		Short: "Get a single user",
		Long: `Get user by name, email, 'me', or UUID. Returns 7 default fields.

Example: go-linear user get me

Related: user_list, user_completed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0], flags)
		},
	}

	flags.Bind(cmd, "defaults (id,name,displayName,email,active,avatarUrl,admin) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, nameOrEmailOrID string, flags *cli.FieldFlags) error {
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

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("user.get", configOverrides)
	fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), user, true, fieldSelector)
}
