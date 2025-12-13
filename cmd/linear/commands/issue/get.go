package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the issue get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single issue by ID",
		Long: `Get issue by identifier (e.g., ENG-123) or UUID. Returns 10 default fields.

Examples:
  go-linear-cli issue get ENG-123 --output=json
  go-linear-cli issue get ENG-123 --fields=defaults,estimate --output=json
  go-linear-cli issue get ENG-123 --fields=id,title,url --output=json

Related: linear issue list (discover IDs), linear issue update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,identifier,title,url,state.name,team.key,priority,createdAt,description,assignee.name) | none | defaults,extra | id,title,...")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, issueID string) error {
	ctx := context.Background()

	issue, err := client.Issue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		// Load config for field defaults
		cfg, _ := config.Load() // Ignore error - config is optional
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}

		// Get command defaults
		defaults := fieldfilter.GetDefaults("issue.get", configOverrides)

		// Parse field selector with defaults
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}

		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), issue, true, fieldSelector)
	case "table":
		// Simple table output for single issue
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", issue.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Title:       %s\n", issue.Title)
		fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", issue.State.Name)
		if issue.Assignee != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Assignee:    %s\n", issue.Assignee.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Priority:    %.0f\n", issue.Priority)
		fmt.Fprintf(cmd.OutOrStdout(), "Updated:     %s\n", issue.UpdatedAt)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
