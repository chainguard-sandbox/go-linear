package issue

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewDeleteCommand creates the issue delete command.
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	outputFlags := &cli.OutputOnlyFlags{}
	var permanent bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an issue",
		Long: `Delete issue. Cannot be undone. Prompts unless --yes.

By default, moves to trash (30-day grace period).
Use --permanent to permanently delete (no grace period).

Example: go-linear issue delete ENG-123
Example: go-linear issue delete ENG-123 --permanent --yes

Related: issue_archive, issue_unarchive, issue_list, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runDelete(cmd, client, args[0], permanent, confirmFlags, outputFlags)
		},
	}

	outputFlags.Bind(cmd)
	confirmFlags.Bind(cmd)
	cmd.Flags().BoolVar(&permanent, "permanent", false, "Permanently delete (cannot be undone, no grace period)")
	return cmd
}

func runDelete(cmd *cobra.Command, client *linear.Client, issueID string, permanent bool, confirmFlags *cli.ConfirmationFlags, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Confirmation prompt unless --yes
	if !confirmFlags.Yes {
		var warning string
		if permanent {
			warning = fmt.Sprintf("Are you sure you want to PERMANENTLY delete issue %s? This cannot be undone.", issueID)
		} else {
			warning = fmt.Sprintf("Are you sure you want to delete issue %s? It will be moved to trash (30-day grace period).", issueID)
		}
		fmt.Fprintf(cmd.OutOrStderr(), "%s\n", warning)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if !strings.EqualFold(response, "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	// Delete issue
	var permanentPtr *bool
	if permanent {
		permanentPtr = &permanent
	}
	err := client.IssueDelete(ctx, issueID, permanentPtr)
	if err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}

	// Format output
	action := "moved to trash"
	if permanent {
		action = "permanently deleted"
	}

	switch outputFlags.Output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success":   true,
			"issueId":   issueID,
			"permanent": permanent,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Issue %s %s successfully\n", issueID, action)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
