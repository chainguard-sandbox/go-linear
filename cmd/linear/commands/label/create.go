package label

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the label create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue label",
		Long: `Create label. Safe operation.

Required: --name, --color (hex like #ff0000)
Optional: --description

Example: go-linear label create --name=bug --color="#ff0000" --description="Bug reports" --output=json

Related: label_list, label_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("name", "", "Label name (required)")
	_ = cmd.MarkFlagRequired("name")

	cmd.Flags().String("color", "", "Label color hex code (e.g., #ff0000) (required)")
	_ = cmd.MarkFlagRequired("color")

	cmd.Flags().String("description", "", "Label description")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	name, _ := cmd.Flags().GetString("name")
	color, _ := cmd.Flags().GetString("color")

	input := intgraphql.IssueLabelCreateInput{
		Name:  name,
		Color: &color,
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	result, err := client.IssueLabelCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created label: %s (%s)\n", result.Name, result.Color)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
