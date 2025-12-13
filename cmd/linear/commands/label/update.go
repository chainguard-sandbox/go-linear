package label

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

func NewUpdateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name|id>",
		Short: "Update an existing label",
		Long: `Update label. Modifies existing data.

Fields: --name, --color (hex), --description

Example: go-linear-cli label update bug --color="#ff0000" --description="Critical bugs" --output=json

Related: label_get, label_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			res := resolver.New(client)

			labelID, err := res.ResolveLabel(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve label: %w", err)
			}

			input := intgraphql.IssueLabelUpdateInput{}
			updated := false

			if name, _ := cmd.Flags().GetString("name"); name != "" {
				input.Name = &name
				updated = true
			}
			if color, _ := cmd.Flags().GetString("color"); color != "" {
				input.Color = &color
				updated = true
			}
			if desc, _ := cmd.Flags().GetString("description"); desc != "" {
				input.Description = &desc
				updated = true
			}

			if !updated {
				return fmt.Errorf("no fields to update specified")
			}

			result, err := client.IssueLabelUpdate(ctx, labelID, input)
			if err != nil {
				return fmt.Errorf("failed to update label: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated label: %s\n", result.Name)
			return nil
		},
	}

	cmd.Flags().String("name", "", "New label name")
	cmd.Flags().String("color", "", "New color hex code")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}
