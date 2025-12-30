package initiative

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the initiative update command.
func NewUpdateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing initiative",
		Long: `Update initiative. Modifies existing data.

Fields: --name, --description

Example: go-linear initiative update <uuid> --name="Updated Security Policy" --output=json

Related: initiative_get, initiative_create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("description", "", "New description (markdown)")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, initiativeID string) error {
	ctx := context.Background()

	input := intgraphql.InitiativeUpdateInput{}
	updated := false

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		input.Name = &name
		updated = true
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	result, err := client.InitiativeUpdate(ctx, initiativeID, input)
	if err != nil {
		return fmt.Errorf("failed to update initiative: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated initiative: %s\n", result.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
