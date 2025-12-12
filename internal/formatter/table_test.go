package formatter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"long text", "this is a very long string", 10, "this is..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatPriority(t *testing.T) {
	tests := []struct {
		priority float64
		want     string
	}{
		{0, "None"},
		{1, "Urgent"},
		{2, "High"},
		{3, "Normal"},
		{4, "Low"},
		{5, "5"},
	}

	for _, tt := range tests {
		got := formatPriority(tt.priority)
		if got != tt.want {
			t.Errorf("formatPriority(%f) = %s, want %s", tt.priority, got, tt.want)
		}
	}
}

func TestFormatTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 min ago", now.Add(-1 * time.Minute), "1 min ago"},
		{"5 mins ago", now.Add(-5 * time.Minute), "5 mins ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1 hour ago"},
		{"3 hours ago", now.Add(-3 * time.Hour), "3 hours ago"},
		{"1 day ago", now.Add(-24 * time.Hour), "1 day ago"},
		{"3 days ago", now.Add(-72 * time.Hour), "3 days ago"},
		{"1 week ago", now.Add(-7 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.time)
			if got != tt.want {
				t.Errorf("formatTime() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFormatAssignee(t *testing.T) {
	tests := []struct {
		name     string
		assignee *intgraphql.ListIssues_Issues_Nodes_Assignee
		want     string
	}{
		{"nil assignee", nil, "-"},
		{"with assignee", &intgraphql.ListIssues_Issues_Nodes_Assignee{Name: "Alice"}, "Alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAssignee(tt.assignee)
			if got != tt.want {
				t.Errorf("formatAssignee() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFormatCSV(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"ID", "Title", "Priority"}
	rows := [][]string{
		{"1", "Test Issue", "High"},
		{"2", "Bug with, comma", "Low"},
		{"3", "Quote \"test\"", "Normal"},
	}

	err := FormatCSV(&buf, headers, rows)
	if err != nil {
		t.Fatalf("FormatCSV() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("FormatCSV() output lines = %d, want 4", len(lines))
	}

	// Check header
	if lines[0] != "ID,Title,Priority" {
		t.Errorf("FormatCSV() header = %s, want ID,Title,Priority", lines[0])
	}

	// Check quoted field with comma
	if !strings.Contains(lines[2], "\"Bug with, comma\"") {
		t.Errorf("FormatCSV() should quote cells with commas, got %s", lines[2])
	}
}

func TestFormatIssuesTable(t *testing.T) {
	var buf bytes.Buffer

	t.Run("empty issues", func(t *testing.T) {
		buf.Reset()
		err := FormatIssuesTable(&buf, nil)
		if err != nil {
			t.Fatalf("FormatIssuesTable() error = %v", err)
		}
		if !strings.Contains(buf.String(), "No issues found") {
			t.Errorf("FormatIssuesTable() should show 'No issues found'")
		}
	})

	t.Run("with issues", func(t *testing.T) {
		buf.Reset()
		issues := []*intgraphql.ListIssues_Issues_Nodes{
			{
				Identifier: "HEX-1",
				Title:      "Test Issue",
				State:      intgraphql.ListIssues_Issues_Nodes_State{Name: "Todo"},
				Priority:   1,
				UpdatedAt:  time.Now().Add(-1 * time.Hour),
			},
		}

		err := FormatIssuesTable(&buf, issues)
		if err != nil {
			t.Fatalf("FormatIssuesTable() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "HEX-1") {
			t.Errorf("FormatIssuesTable() should contain issue identifier")
		}
		if !strings.Contains(output, "Test Issue") {
			t.Errorf("FormatIssuesTable() should contain title")
		}
	})
}
