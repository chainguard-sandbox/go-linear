package customview

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewCreateCommand creates the custom-view create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new custom view",
		Long: `Create a custom view. Safe operation.

Required: --name
Optional: --description, --icon, --color, --team, --shared, --filter-data

--filter-data accepts a JSON string or @filename for IssueFilter JSON.

Example: go-linear custom-view create --name="My Bugs" --filter-data='{"priority":{"lte":2}}'
         go-linear custom-view create --name="Team View" --team=ENG --shared

Related: custom-view_get, custom-view_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("name", "", "Custom view name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("description", "", "Custom view description")
	cmd.Flags().String("icon", "", "Custom view icon")
	cmd.Flags().String("color", "", "Icon color")
	cmd.Flags().String("team", "", "Team name, key, or UUID")
	cmd.Flags().Bool("shared", false, "Share with organization")
	cmd.Flags().String("filter-data", "", "Issue filter JSON string or @filename")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	name, _ := cmd.Flags().GetString("name")
	input := intgraphql.CustomViewCreateInput{
		Name: name,
	}

	if description, _ := cmd.Flags().GetString("description"); description != "" {
		input.Description = &description
	}

	if icon, _ := cmd.Flags().GetString("icon"); icon != "" {
		input.Icon = &icon
	}

	if color, _ := cmd.Flags().GetString("color"); color != "" {
		input.Color = &color
	}

	if shared, _ := cmd.Flags().GetBool("shared"); cmd.Flags().Changed("shared") {
		input.Shared = &shared
	}

	if team, _ := cmd.Flags().GetString("team"); team != "" {
		res := resolver.New(client)
		teamID, err := res.ResolveTeam(ctx, team)
		if err != nil {
			return fmt.Errorf("failed to resolve team: %w", err)
		}
		input.TeamID = &teamID
	}

	if filterData, _ := cmd.Flags().GetString("filter-data"); filterData != "" {
		filter, err := parseFilterData(filterData)
		if err != nil {
			return fmt.Errorf("invalid --filter-data: %w", err)
		}
		input.FilterData = filter
	}

	result, err := client.CustomViewCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create custom view: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}

// parseFilterData parses a JSON string or @filename into an IssueFilter.
func parseFilterData(value string) (*intgraphql.IssueFilter, error) {
	var jsonData string

	if strings.HasPrefix(value, "@") {
		filename := strings.TrimPrefix(value, "@")
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", filename, err)
		}
		jsonData = string(data)
	} else {
		jsonData = value
	}

	var filter intgraphql.IssueFilter
	if err := json.Unmarshal([]byte(jsonData), &filter); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &filter, nil
}
