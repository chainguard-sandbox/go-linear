package document

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewCreateCommand creates the document create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new document",
		Long: `Create knowledge base document. Safe operation.

Required: --title, and exactly ONE of: --project or --issue
Optional: --content

Example: go-linear document create --title="API Guide" --project=<uuid> --content="# API Documentation..."

Related: document_get, document_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.RejectRemovedFlags(cmd, "link the document to a project (--project) or an issue (--issue) instead", "initiative", "team"); err != nil {
				return err
			}

			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("title", "", "Document title (required)")
	_ = cmd.MarkFlagRequired("title")
	cmd.Flags().String("content", "", "Document content (markdown)")
	cmd.Flags().String("project", "", "Project name or UUID to link")
	cmd.Flags().String("initiative", "", "Initiative name or UUID to link")
	cmd.Flags().String("team", "", "Team name or key to link")
	cmd.Flags().String("issue", "", "Issue identifier or UUID to link")
	// Linear removed initiativeId and teamId from DocumentCreateInput in
	// @linear/sdk 103.0.0. Keep the flags, hidden, so they fail with an
	// explanation.
	cli.HideRemovedFlags(cmd, "initiative", "team")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	title, _ := cmd.Flags().GetString("title")
	input := intgraphql.DocumentCreateInput{
		Title: title, // Not a pointer - title is required in Linear API (String! in schema)
	}

	if content, _ := cmd.Flags().GetString("content"); content != "" {
		input.Content = &content
	}

	// Validate exactly one of project/issue is set
	project, _ := cmd.Flags().GetString("project")
	issue, _ := cmd.Flags().GetString("issue")

	setCount := 0
	if project != "" {
		setCount++
	}
	if issue != "" {
		setCount++
	}

	if setCount == 0 {
		return fmt.Errorf("must specify exactly one of: --project or --issue")
	}
	if setCount > 1 {
		return fmt.Errorf("must specify exactly one of: --project or --issue (not multiple)")
	}

	// Resolve and set the appropriate entity
	if project != "" {
		projectID, err := res.ResolveProject(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}
		input.ProjectID = &projectID
	}

	if issue != "" {
		issueID, err := res.ResolveIssue(ctx, issue)
		if err != nil {
			return fmt.Errorf("failed to resolve issue: %w", err)
		}
		input.IssueID = &issueID
	}

	result, err := client.DocumentCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
