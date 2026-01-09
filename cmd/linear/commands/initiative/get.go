package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.OutputFlags{}

	cmd := &cobra.Command{
		Use:   "get <name-or-id>",
		Short: "Get a single initiative by name or ID",
		Long: `Get initiative by name or UUID. Returns 4 default fields. Shows parent initiative and sub-initiative count in hierarchy.

Example: go-linear initiative get "Shrink Wolfi" --output=json
Example: go-linear initiative get <uuid> --output=json

Related: initiative_list, project_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			// Resolve initiative name to UUID
			res := resolver.New(client)
			initiativeID, err := res.ResolveInitiative(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve initiative: %w", err)
			}

			initiative, err := client.Initiative(ctx, initiativeID)
			if err != nil {
				return fmt.Errorf("failed to get initiative: %w", err)
			}

			switch flags.Output {
			case "json":
				cfg, _ := config.Load()
				var configOverrides map[string]string
				if cfg != nil {
					configOverrides = cfg.FieldDefaults
				}
				defaults := fieldfilter.GetDefaults("initiative.get", configOverrides)
				fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
				if err != nil {
					return fmt.Errorf("invalid --fields: %w", err)
				}
				return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiative, true, fieldSelector)
			case "table":
				fmt.Fprintf(cmd.OutOrStdout(), "ID:     %s\n", initiative.ID)
				fmt.Fprintf(cmd.OutOrStdout(), "Name:   %s\n", initiative.Name)
				if initiative.Status != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", initiative.Status)
				}

				// Health and target date
				if initiative.Health != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Health: %s\n", *initiative.Health)
				}
				if initiative.TargetDate != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Target: %s\n", *initiative.TargetDate)
				}

				// Ownership
				if initiative.Owner != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Owner:  %s\n", initiative.Owner.Name)
				}

				// Relationships
				if initiative.ParentInitiative != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Parent:   %s\n", initiative.ParentInitiative.Name)
				}

				// Linked projects
				if len(initiative.Projects.Nodes) > 0 {
					projectNames := make([]string, 0, len(initiative.Projects.Nodes))
					for _, proj := range initiative.Projects.Nodes {
						status := ""
						if proj.Progress > 0 {
							status = fmt.Sprintf(" (%.0f%%)", proj.Progress*100)
						}
						projectNames = append(projectNames, proj.Name+status)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Projects: %d linked\n", len(initiative.Projects.Nodes))
					for _, name := range projectNames {
						fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", name)
					}
				}

				// Additional info
				if initiative.Description != nil && *initiative.Description != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "\nDescription:\n%s\n", *initiative.Description)
				}
				if initiative.URL != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "\nURL: %s\n", initiative.URL)
				}
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", flags.Output)
			}
		},
	}

	flags.Bind(cmd, "defaults (id,name,description,createdAt) | none | defaults,extra")

	return cmd
}
