package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the team get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.OutputFlags{}

	cmd := &cobra.Command{
		Use:   "get <name|id>",
		Short: "Get a single team",
		Long: `Get team by name, key (e.g., ENG), or UUID. Returns 9 default fields (includes issueCount).

Example: go-linear team get ENG --output=json

Related: team_list, team_members`,
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

	flags.Bind(cmd, "defaults (id,name,key,description,icon,createdAt,color,private) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, nameOrID string, flags *cli.OutputFlags) error {
	ctx := cmd.Context()

	if err := flags.Validate(); err != nil {
		return err
	}

	res := resolver.New(client)

	// Resolve to team ID
	teamID, err := res.ResolveTeam(ctx, nameOrID)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	team, err := client.Team(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	switch flags.Output {
	case "json":
		// Load config for field defaults
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}

		// Get command defaults
		defaults := fieldfilter.GetDefaults("team.get", configOverrides)

		// Parse field selector with defaults
		fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}

		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), team, true, fieldSelector)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", team.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Key:  %s\n", team.Key)
		if team.Description != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", *team.Description)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", flags.Output)
	}
}
