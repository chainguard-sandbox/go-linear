package customview

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewUpdateCommand creates the custom-view update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a custom view",
		Long: `Update an existing custom view. Modifies existing data.

Fields: --name, --description, --icon, --color, --shared, --filter-data, --team

Example: go-linear custom-view update <uuid> --name="Updated View"
         go-linear custom-view update <uuid> --shared=true --filter-data='{"priority":{"lte":1}}'

Related: custom-view_get, custom-view_create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("name", "", "New custom view name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("icon", "", "New icon")
	cmd.Flags().String("color", "", "New icon color")
	cmd.Flags().String("team", "", "Team name, key, or UUID")
	cmd.Flags().Bool("shared", false, "Share with organization")
	cmd.Flags().String("filter-data", "", "Issue filter JSON string or @filename")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, id string) error {
	ctx := cmd.Context()

	input := intgraphql.CustomViewUpdateInput{}
	updated := false

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		input.Name = &name
		updated = true
	}

	if description, _ := cmd.Flags().GetString("description"); description != "" {
		input.Description = &description
		updated = true
	}

	if icon, _ := cmd.Flags().GetString("icon"); icon != "" {
		input.Icon = &icon
		updated = true
	}

	if color, _ := cmd.Flags().GetString("color"); color != "" {
		input.Color = &color
		updated = true
	}

	if cmd.Flags().Changed("shared") {
		shared, _ := cmd.Flags().GetBool("shared")
		input.Shared = &shared
		updated = true
	}

	if team, _ := cmd.Flags().GetString("team"); team != "" {
		res := resolver.New(client)
		teamID, err := res.ResolveTeam(ctx, team)
		if err != nil {
			return fmt.Errorf("failed to resolve team: %w", err)
		}
		input.TeamID = &teamID
		updated = true
	}

	if filterData, _ := cmd.Flags().GetString("filter-data"); filterData != "" {
		filter, err := parseFilterData(filterData)
		if err != nil {
			return fmt.Errorf("invalid --filter-data: %w", err)
		}
		input.FilterData = filter
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	result, err := client.CustomViewUpdate(ctx, id, input)
	if err != nil {
		return fmt.Errorf("failed to update custom view: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
