package initiative

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all initiatives",
		Long: `List strategic initiatives from Linear.

Use this to:
- Browse company-wide strategic initiatives
- Track high-level goals and objectives
- Discover initiatives for organizational planning

Initiatives represent large strategic efforts that span multiple projects and teams.

Output (--output=json):
  Returns JSON with:
  - nodes: Array of initiatives
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each initiative contains:
  - id: Initiative UUID
  - name: Initiative name
  - description: Initiative description
  - targetDate: Target completion date

Examples:
  # List all initiatives
  linear initiative list

  # List with limit
  linear initiative list --limit=10

  # JSON output for parsing
  linear initiative list --output=json

Related Commands:
  - linear initiative get - Get single initiative details
  - linear project list - List projects (initiatives contain projects)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			limit, _ := cmd.Flags().GetInt("limit")
			first := int64(limit)

			initiatives, err := client.Initiatives(ctx, &first, nil)
			if err != nil {
				return fmt.Errorf("failed to list initiatives: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			fieldsSpec, _ := cmd.Flags().GetString("fields")

			switch output {
			case "json":
				fieldSelector, err := fieldfilter.New(fieldsSpec)
				if err != nil {
					return fmt.Errorf("invalid --fields: %w", err)
				}
				return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiatives, true, fieldSelector)
			case "table":
				for _, init := range initiatives.Nodes {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", init.Name)
				}
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "Comma-separated fields for JSON output (e.g., 'id,name')")

	return cmd
}
