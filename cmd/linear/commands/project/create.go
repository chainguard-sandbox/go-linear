package project

import (
	"context"
	"fmt"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// NewCreateCommand creates the project create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create a new project in Linear.

This operation creates new project data and is safe to execute.

Examples:
  linear project create --name="Q1 Platform" --description="Platform improvements"
  linear project create --name="Mobile App Redesign" --description="Complete UI overhaul"

TIP: Projects help organize multiple related issues across teams`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("name", "", "Project name (required)")
	cmd.MarkFlagRequired("name")

	cmd.Flags().String("description", "", "Project description")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	name, _ := cmd.Flags().GetString("name")
	input := intgraphql.ProjectCreateInput{
		Name: name,
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	result, err := client.ProjectCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created project: %s\n", result.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
