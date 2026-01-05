// Package template provides template commands for the Linear CLI.
package template

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

type ClientFactory func() (*linear.Client, error)

func NewTemplateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage Linear templates",
		Long:  "Commands for listing and viewing Linear issue templates.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))

	return cmd
}

func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all templates",
		Long: `List templates. Returns 4 default fields per template.

Example: go-linear template list --output=json

Related: template_get, issue_create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			templates, err := client.Templates(ctx)
			if err != nil {
				return fmt.Errorf("failed to list templates: %w", err)
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
				defaults := fieldfilter.GetDefaults("template.list", configOverrides)
				fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
				if err != nil {
					return fmt.Errorf("invalid --fields: %w", err)
				}
				return formatter.FormatJSONFiltered(cmd.OutOrStdout(), templates, true, fieldSelector)
			case "table":
				for _, tmpl := range templates {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", tmpl.Name)
				}
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,description,createdAt) | none | defaults,extra")
	return cmd
}

func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single template by ID",
		Long: `Get template by UUID. Returns 4 default fields.

Example: go-linear template get <uuid> --output=json

Related: template_list, issue_create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			template, err := client.Template(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get template: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				cfg, _ := config.Load()
				var configOverrides map[string]string
				if cfg != nil {
					configOverrides = cfg.FieldDefaults
				}
				fieldsSpec, _ := cmd.Flags().GetString("fields")
				defaults := fieldfilter.GetDefaults("template.get", configOverrides)
				fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
				if err != nil {
					return fmt.Errorf("invalid --fields: %w", err)
				}
				return formatter.FormatJSONFiltered(cmd.OutOrStdout(), template, true, fieldSelector)
			case "table":
				fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", template.Name)
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,description,createdAt) | none | defaults,extra")
	return cmd
}
