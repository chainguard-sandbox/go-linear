// Package template provides template commands for the Linear CLI.
package template

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()

			templates, err := client.Templates(ctx)
			if err != nil {
				return fmt.Errorf("failed to list templates: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				return formatter.FormatJSON(cmd.OutOrStdout(), templates, true)
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
	return cmd
}

func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single template by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			template, err := client.Template(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get template: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				return formatter.FormatJSON(cmd.OutOrStdout(), template, true)
			case "table":
				fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", template.Name)
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	return cmd
}
