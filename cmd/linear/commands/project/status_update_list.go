package project

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

// NewStatusUpdateListCommand creates the project status-update-list command.
func NewStatusUpdateListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}
	outputFlags := &cli.OutputFlags{}

	cmd := &cobra.Command{
		Use:   "status-update-list",
		Short: "List status updates for a project",
		Long: `List project status updates. Returns 5 default fields per update.

Required: --project (UUID or name)

Example: go-linear project status-update-list --project=<uuid> --output=json

Related: project_status-update-create, project_status-update-get, project_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusUpdateList(cmd, client, outputFlags, paginationFlags)
		},
	}

	cmd.Flags().String("project", "", "Project name or UUID (required)")
	_ = cmd.MarkFlagRequired("project")

	paginationFlags.Bind(cmd, 50)
	outputFlags.Bind(cmd, "defaults (id,body,health,createdAt,user.name) | none | defaults,extra")

	return cmd
}

func runStatusUpdateList(cmd *cobra.Command, client *linear.Client, outputFlags *cli.OutputFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

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

	switch outputFlags.Output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("project.status-update-list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(outputFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), updates, true, fieldSelector)
	case "table":
		if len(updates.ProjectUpdates.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No status updates found")
			return nil
		}
		for _, update := range updates.ProjectUpdates.Nodes {
			health := ""
			if update.Health != "" {
				health = fmt.Sprintf(" [%s]", update.Health)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", update.ID, health)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
