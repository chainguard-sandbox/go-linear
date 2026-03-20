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

// NewStatusUpdateListCommand creates the project status-update-list command.
func NewStatusUpdateListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}
	fieldFlags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "status-update-list",
		Short: "List status updates for a project",
		Long: `List project status updates. Returns 5 default fields per update.

Required: --project (UUID or name)

Example: go-linear project status-update-list --project=<uuid>

Related: project_status-update-create, project_status-update-get, project_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusUpdateList(cmd, client, fieldFlags, paginationFlags)
		},
	}

	cmd.Flags().String("project", "", "Project name or UUID (required)")
	_ = cmd.MarkFlagRequired("project")

	paginationFlags.Bind(cmd, 50)
	fieldFlags.Bind(cmd, "defaults (id,body,health,createdAt,user.name) | none | defaults,extra")

	return cmd
}

func runStatusUpdateList(cmd *cobra.Command, client *linear.Client, fieldFlags *cli.FieldFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	res := resolver.New(client)

	// Resolve project
	projectInput, _ := cmd.Flags().GetString("project")
	projectID, err := res.ResolveProject(ctx, projectInput)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	first := paginationFlags.LimitPtr()

	updates, err := client.ListProjectUpdates(ctx, projectID, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list project status updates: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("project.status-update-list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), updates, true, fieldSelector)
}
