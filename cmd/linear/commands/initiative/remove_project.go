package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewRemoveProjectCommand creates the initiative remove-project command.
func NewRemoveProjectCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-project",
		Short: "Unlink a project from an initiative",
		Long: `Unlink project from initiative. Modifies existing data.

Removes the InitiativeToProject association. The project and initiative
themselves are not deleted, only the link between them.

Required: --initiative (UUID or name), --project (UUID or name)

Example: go-linear initiative remove-project --initiative=<uuid> --project=<uuid>

Related: initiative_add-project, project_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRemoveProject(cmd, client)
		},
	}

	cmd.Flags().String("initiative", "", "Initiative name or UUID (required)")
	_ = cmd.MarkFlagRequired("initiative")
	cmd.Flags().String("project", "", "Project name or UUID (required)")
	_ = cmd.MarkFlagRequired("project")

	return cmd
}

func runRemoveProject(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve initiative
	initiativeInput, _ := cmd.Flags().GetString("initiative")
	initiativeID, err := res.ResolveInitiative(ctx, initiativeInput)
	if err != nil {
		return fmt.Errorf("failed to resolve initiative: %w", err)
	}

	// Resolve project
	projectInput, _ := cmd.Flags().GetString("project")
	projectID, err := res.ResolveProject(ctx, projectInput)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	// Find the link by searching through all links with pagination
	// Note: Linear API doesn't support filtering initiativeToProjects, so we must search
	const searchBatchSize = 100 // Links to fetch per page
	var linkID string
	var linkFound bool
	limit := int64(searchBatchSize)
	cursor := (*string)(nil)

	for {
		links, err := client.ListInitiativeToProjects(ctx, &limit, cursor)
		if err != nil {
			return fmt.Errorf("failed to list initiative-project links: %w", err)
		}

		// Search current page for matching link
		for _, link := range links.Nodes {
			if link.Initiative.ID == initiativeID && link.Project.ID == projectID {
				linkID = link.ID
				linkFound = true
				break
			}
		}

		// Exit if found or no more pages
		if linkFound || !links.PageInfo.HasNextPage {
			break
		}

		cursor = links.PageInfo.EndCursor
	}

	if !linkFound {
		return fmt.Errorf("no link found between initiative %s and project %s", initiativeInput, projectInput)
	}

	// Delete the link
	err = client.InitiativeToProjectDelete(ctx, linkID)
	if err != nil {
		return fmt.Errorf("failed to unlink project from initiative: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
}
