package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the initiative create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new initiative",
		Long: `Create initiative. Safe operation.

Required: --name
Optional: --description

Example: go-linear initiative create --name="Security Policy" --description="Improve security" --output=json

Related: initiative_get, initiative_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("name", "", "Initiative name (required)")
	_ = cmd.MarkFlagRequired("name")

	cmd.Flags().String("description", "", "Initiative description (markdown)")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	name, _ := cmd.Flags().GetString("name")
	input := intgraphql.InitiativeCreateInput{
		Name: name,
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	result, err := client.InitiativeCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create initiative: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created initiative: %s\n", result.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
