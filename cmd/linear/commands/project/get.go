package project

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the project get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.OutputFlags{}

	cmd := &cobra.Command{
		Use:   "get <name-or-id>",
		Short: "Get a single project by name or ID",
		Long: `Get project by name or UUID. Returns 6 default fields.

Example: go-linear project get "Cloud cost optimization" --output=json
Example: go-linear project get <uuid> --output=json

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

			return runGet(cmd, client, projectID, flags)
		},
	}

	flags.Bind(cmd, "defaults (id,name,description,createdAt,color,state) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, projectID string, flags *cli.OutputFlags) error {
	ctx := cmd.Context()

	if err := flags.Validate(); err != nil {
		return err
	}

	project, err := client.Project(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	switch flags.Output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("project.get", configOverrides)
		fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), project, true, fieldSelector)
	case "table":
		// Identity
		fmt.Fprintf(cmd.OutOrStdout(), "ID:       %s\n", project.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Name:     %s\n", project.Name)

		// Status
		fmt.Fprintf(cmd.OutOrStdout(), "State:    %s\n", project.State)
		fmt.Fprintf(cmd.OutOrStdout(), "Progress: %.1f%%\n", project.Progress*100)
		if project.Health != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Health:   %s\n", *project.Health)
		}

		// Ownership
		if project.Lead != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Lead:     %s\n", project.Lead.Name)
		}
		if len(project.Teams.Nodes) > 0 {
			teamNames := make([]string, len(project.Teams.Nodes))
			for i, team := range project.Teams.Nodes {
				teamNames[i] = team.Key
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Teams:    %s\n", strings.Join(teamNames, ", "))
		}
		if len(project.Members.Nodes) > 0 {
			memberNames := make([]string, len(project.Members.Nodes))
			for i, member := range project.Members.Nodes {
				memberNames[i] = member.Name
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Members:  %s\n", strings.Join(memberNames, ", "))
		}

		// Timeline
		if project.TargetDate != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Target:   %s\n", *project.TargetDate)
		}
		if project.StartedAt != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Started:  %s\n", project.StartedAt.Format("2006-01-02"))
		}
		if project.CompletedAt != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Completed: %s\n", project.CompletedAt.Format("2006-01-02"))
		}
		if project.CanceledAt != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Canceled:  %s\n", project.CanceledAt.Format("2006-01-02"))
		}

		// Linked initiatives
		if len(project.Initiatives.Nodes) > 0 {
			initNames := make([]string, len(project.Initiatives.Nodes))
			for i, init := range project.Initiatives.Nodes {
				status := string(init.Status)
				if status != "" {
					initNames[i] = fmt.Sprintf("%s (%s)", init.Name, status)
				} else {
					initNames[i] = init.Name
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nInitiatives: %d linked\n", len(project.Initiatives.Nodes))
			for _, name := range initNames {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", name)
			}
		}

		// Additional info
		if project.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "\nDescription:\n%s\n", project.Description)
		}
		if project.URL != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "\nURL: %s\n", project.URL)
		}

		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", flags.Output)
	}
}
