package formatter

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// FormatIssuesTable writes issues as a formatted table to the writer.
func FormatIssuesTable(w io.Writer, issues []*intgraphql.ListIssues_Issues_Nodes) error {
	if len(issues) == 0 {
		fmt.Fprintln(w, "No issues found")
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.Header("ID", "Title", "State", "Assignee", "Priority", "Updated")

	for _, issue := range issues {
		_ = table.Append(
			issue.Identifier,
			truncate(issue.Title, 50),
			issue.State.Name,
			formatAssignee(issue.Assignee),
			formatPriority(issue.Priority),
			formatTime(issue.UpdatedAt),
		)
	}

	return table.Render()
}

// FormatTeamsTable writes teams as a formatted table to the writer.
func FormatTeamsTable(w io.Writer, teams []*intgraphql.ListTeams_Teams_Nodes) error {
	if len(teams) == 0 {
		fmt.Fprintln(w, "No teams found")
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.Header("Key", "Name", "Description")

	for _, team := range teams {
		desc := ""
		if team.Description != nil {
			desc = truncate(*team.Description, 60)
		}
		_ = table.Append(team.Key, team.Name, desc)
	}

	return table.Render()
}

// FormatUsersTable writes users as a formatted table to the writer.
func FormatUsersTable(w io.Writer, users []*intgraphql.ListUsers_Users_Nodes) error {
	if len(users) == 0 {
		fmt.Fprintln(w, "No users found")
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.Header("Name", "Email", "Active")

	for _, user := range users {
		active := "Yes"
		if !user.Active {
			active = "No"
		}
		_ = table.Append(user.Name, user.Email, active)
	}

	return table.Render()
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatAssignee(assignee *intgraphql.ListIssues_Issues_Nodes_Assignee) string {
	if assignee == nil {
		return "-"
	}
	return assignee.Name
}

func formatPriority(priority float64) string {
	switch int(priority) {
	case 0:
		return "None"
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Normal"
	case 4:
		return "Low"
	default:
		return fmt.Sprintf("%d", int(priority))
	}
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	return t.Format("Jan 2, 2006")
}

// FormatCSV writes data as CSV to the writer.
func FormatCSV(w io.Writer, headers []string, rows [][]string) error {
	// Write header
	fmt.Fprintln(w, strings.Join(headers, ","))

	// Write rows
	for _, row := range rows {
		quoted := make([]string, len(row))
		for i, cell := range row {
			// Quote cells that contain commas or quotes
			if strings.Contains(cell, ",") || strings.Contains(cell, "\"") {
				quoted[i] = fmt.Sprintf("%q", cell)
			} else {
				quoted[i] = cell
			}
		}
		fmt.Fprintln(w, strings.Join(quoted, ","))
	}

	return nil
}
