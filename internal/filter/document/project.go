package document

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyProject handles --project flag.
func ApplyProject(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	project, _ := cmd.Flags().GetString("project")
	if project == "" {
		return nil
	}

	projectID, err := b.Resolver().ResolveProject(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	b.Filter().Project = &intgraphql.ProjectFilter{
		ID: &intgraphql.IDComparator{Eq: &projectID},
	}
	return nil
}
