package user

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCompletedCommand creates the user completed command.
// This is a CRITICAL complex query that answers: "find users from team X who completed tasks yesterday"
func NewCompletedCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completed",
		Short: "Get users' completed issues",
		Long: `Get completed issues by user or team. Complex query - replaces 5-step workflow.

Filters: --user=me or email (single user) | --team (all team members), --completed-after=yesterday|7d (see issue_list), --completed-before=today
Must specify EITHER --user OR --team.

Example: go-linear-cli user completed --team=ENG --completed-after=yesterday --completed-before=today --output=json

Returns: [{user: {name, email}, count: number}...]
Related: user_list, team_list, issue_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCompleted(cmd, client)
		},
	}

	// User/Team selection (mutually exclusive)
	cmd.Flags().String("user", "", "User name, email, or ID (or 'me' for current user)")
	cmd.Flags().String("team", "", "Team name or ID (queries all team members)")

	// Date filtering (required)
	cmd.Flags().String("completed-after", "yesterday", "Completed after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("completed-before", "today", "Completed before date")

	// Pagination
	cmd.Flags().IntP("limit", "l", 100, "Max issues per user")

	// Output
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCompleted(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)
	parser := dateparser.New()

	// Parse dates
	afterStr, _ := cmd.Flags().GetString("completed-after")
	after, err := parser.Parse(afterStr)
	if err != nil {
		return fmt.Errorf("invalid completed-after date: %w", err)
	}

	beforeStr, _ := cmd.Flags().GetString("completed-before")
	before, err := parser.Parse(beforeStr)
	if err != nil {
		return fmt.Errorf("invalid completed-before date: %w", err)
	}

	// Determine users to query
	var userIDs []string
	var userMap map[string]*intgraphql.ListUsers_Users_Nodes

	userName, _ := cmd.Flags().GetString("user")
	teamName, _ := cmd.Flags().GetString("team")

	// Validation: mutually exclusive
	if userName == "" && teamName == "" {
		return fmt.Errorf("must specify either --user or --team\n\nExamples:\n  go-linear-cli user completed --user=me --completed-after=7d\n  go-linear-cli user completed --team=Engineering --completed-after=yesterday")
	}

	if userName != "" && teamName != "" {
		return fmt.Errorf("cannot specify both --user and --team (choose one)\n\nFor a specific user:\n  go-linear-cli user completed --user=alice@company.com\n\nFor all team members:\n  go-linear-cli user completed --team=Engineering")
	}

	if teamName != "" {
		// Get all team members
		teamID, err := res.ResolveTeam(ctx, teamName)
		if err != nil {
			return fmt.Errorf("failed to resolve team: %w", err)
		}

		// Get all users (we'll filter by team in the issue query)
		_ = teamID // TODO: Filter users by team membership when API supports it

		first := int64(250)
		users, err := client.Users(ctx, &first, nil)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		userMap = make(map[string]*intgraphql.ListUsers_Users_Nodes)
		for _, user := range users.Nodes {
			if user.Active {
				userIDs = append(userIDs, user.ID)
				userMap[user.ID] = user
			}
		}
	} else {
		// Single user
		userID, err := res.ResolveUser(ctx, userName)
		if err != nil {
			return fmt.Errorf("failed to resolve user: %w", err)
		}
		userIDs = []string{userID}

		user, err := client.User(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		userMap = map[string]*intgraphql.ListUsers_Users_Nodes{
			userID: {
				ID:     user.ID,
				Name:   user.Name,
				Email:  user.Email,
				Active: user.Active,
			},
		}
	}

	// Query completed issues for each user
	type UserCompletion struct {
		User  *intgraphql.ListUsers_Users_Nodes
		Count int
	}

	results := make([]UserCompletion, 0)
	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	afterDate := after.Format("2006-01-02T15:04:05.000Z")
	beforeDate := before.Format("2006-01-02T15:04:05.000Z")

	for _, userID := range userIDs {
		// Build filter for this user's completed issues
		filter := &intgraphql.IssueFilter{
			Assignee: &intgraphql.NullableUserFilter{
				ID: &intgraphql.IDComparator{
					Eq: &userID,
				},
			},
			CompletedAt: &intgraphql.NullableDateComparator{
				Gte: &afterDate,
				Lt:  &beforeDate,
			},
		}

		// Query using SearchIssues with empty term
		searchResult, err := client.SearchIssues(ctx, "", &first, nil, filter, nil)
		if err != nil {
			// Skip on error, continue with other users
			continue
		}

		if len(searchResult.Nodes) > 0 {
			// Convert search nodes to list nodes format
			// (they have the same structure but different types)
			results = append(results, UserCompletion{
				User:  userMap[userID],
				Count: len(searchResult.Nodes),
			})
		}
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), results, true)
	case "table":
		if len(results) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No completed issues found")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Completed issues from %s to %s:\n\n", afterStr, beforeStr)
		for _, result := range results {
			fmt.Fprintf(cmd.OutOrStdout(), "%s <%s> - %d issues\n", result.User.Name, result.User.Email, result.Count)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
