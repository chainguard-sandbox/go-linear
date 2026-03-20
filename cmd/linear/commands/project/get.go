package project

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

// NewGetCommand creates the project get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	fieldFlags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <name-or-id>",
		Short: "Get a single project by name or ID",
		Long: `Get project by name or UUID. Returns 6 default fields.

Example: go-linear project get "Cloud cost optimization"
Example: go-linear project get <uuid>

Related: project_list, project_milestone-create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			// Resolve project name to UUID
			res := resolver.New(client)
			projectID, err := res.ResolveProject(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve project: %w", err)
			}

			return runGet(cmd, client, projectID, fieldFlags)
		},
	}

	fieldFlags.Bind(cmd, "defaults (id,name,description,createdAt,color,state) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, projectID string, fieldFlags *cli.FieldFlags) error {
	ctx := cmd.Context()

	project, err := client.Project(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("project.get", configOverrides)
	fieldSelector, err := fieldfilter.New(fieldFlags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), project, true, fieldSelector)
}
