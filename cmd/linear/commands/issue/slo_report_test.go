package issue

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// sloMockServer serves teams and labels for the resolver and two pages of issues.
// It records the filter variable of every issues request.
type sloMockServer struct {
	*httptest.Server
	mu      sync.Mutex
	filters []map[string]any
	pages   int
}

const sloPage1 = `{"data": {"issues": {
	"nodes": [
		{"id": "i1", "identifier": "ENG-1", "title": "p1 fixed fast", "priority": 1, "url": "u", "number": 1,
		 "state": {"id": "s", "name": "Done"}, "team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
		 "assignee": {"id": "u1", "name": "Alice", "displayName": "alice", "email": "a@x"},
		 "createdAt": "2026-02-01T00:00:00Z", "updatedAt": "2026-02-02T00:00:00Z", "completedAt": "2026-02-02T00:00:00Z",
		 "labels": {"nodes": [{"id": "l1", "name": "P1"}]}},
		{"id": "i2", "identifier": "ENG-2", "title": "p1 and p2, late", "priority": 1, "url": "u", "number": 2,
		 "state": {"id": "s", "name": "Done"}, "team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
		 "createdAt": "2026-02-01T00:00:00Z", "updatedAt": "2026-02-06T00:00:00Z", "completedAt": "2026-02-06T00:00:00Z",
		 "labels": {"nodes": [{"id": "l1", "name": "p1"}, {"id": "l2", "name": "p2"}]}}
	],
	"pageInfo": {"hasNextPage": true, "endCursor": "cursor-1"}
}}}`

const sloPage2 = `{"data": {"issues": {
	"nodes": [
		{"id": "i3", "identifier": "ENG-3", "title": "p2 open, at risk", "priority": 2, "url": "u", "number": 3,
		 "state": {"id": "s", "name": "In Progress"}, "team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
		 "createdAt": "2026-02-17T00:00:00Z", "updatedAt": "2026-02-17T00:00:00Z", "completedAt": null,
		 "labels": {"nodes": [{"id": "l2", "name": "p2"}]}},
		{"id": "i4", "identifier": "ENG-4", "title": "no policy label", "priority": 3, "url": "u", "number": 4,
		 "state": {"id": "s", "name": "Todo"}, "team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
		 "createdAt": "2026-02-20T00:00:00Z", "updatedAt": "2026-02-20T00:00:00Z", "completedAt": null,
		 "labels": {"nodes": []}}
	],
	"pageInfo": {"hasNextPage": false, "endCursor": null}
}}}`

func newSLOMockServer(t *testing.T) *sloMockServer {
	t.Helper()
	s := &sloMockServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		switch {
		case strings.Contains(req.Query, "issueLabels"):
			_, _ = w.Write([]byte(mockLabelsResponse))
		case strings.Contains(req.Query, "teams"):
			_, _ = w.Write([]byte(mockTeamsResponse))
		case strings.Contains(req.Query, "issues"):
			s.mu.Lock()
			f, _ := req.Variables["filter"].(map[string]any)
			s.filters = append(s.filters, f)
			s.pages++
			s.mu.Unlock()
			if after, _ := req.Variables["after"].(string); after == "cursor-1" {
				_, _ = w.Write([]byte(sloPage2))
			} else {
				_, _ = w.Write([]byte(sloPage1))
			}
		default:
			http.Error(w, "unexpected query", http.StatusBadRequest)
		}
	}))
	t.Cleanup(s.Close)
	return s
}

// sloTestNow is 2026-03-01: ENG-3 (p2, created 2026-02-17) is 12 of 14 days in.
var sloTestNow = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

func runSLOCommand(t *testing.T, srv *sloMockServer, args ...string) (string, error) {
	t.Helper()
	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(srv.URL))
	}
	cmd := newSLOReportCommand(factory, func() time.Time { return sloTestNow })
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestSLOReportCommand_JSON(t *testing.T) {
	srv := newSLOMockServer(t)
	out, err := runSLOCommand(t, srv, "--policy", "p1=2d,p2=14d", "--team", "ENG", "--group-by", "label")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var report sloReport
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}

	if srv.pages != 2 {
		t.Errorf("fetched %d pages, want 2", srv.pages)
	}
	if report.Fetched != 4 || report.Unmatched != 1 {
		t.Errorf("fetched=%d unmatched=%d, want 4 and 1", report.Fetched, report.Unmatched)
	}
	if !report.GeneratedAt.Equal(sloTestNow) {
		t.Errorf("generatedAt = %v, want %v", report.GeneratedAt, sloTestNow)
	}

	// ENG-1: P1 (case-insensitive) met in 1d. ENG-2: p1 governs (strictest), missed at 5d.
	// ENG-3: p2 open 12d of 14d -> at risk.
	if got := strings.Join(groupKeys(report.Groups), ","); got != "p1,p2" {
		t.Fatalf("group keys = %s, want p1,p2", got)
	}
	p1, p2 := report.Groups[0], report.Groups[1]
	if p1.Total != 2 || p1.Met != 1 || p1.Missed != 1 || p1.MedianElapsed != "3d" || p1.MaxElapsed != "5d" {
		t.Errorf("p1 = %+v", p1)
	}
	if p2.Total != 1 || p2.AtRisk != 1 || p2.MaxElapsedSeconds != int64(12*24*time.Hour/time.Second) {
		t.Errorf("p2 = %+v", p2)
	}
	o := report.Overall
	if o.Total != 3 || o.Met != 1 || o.Missed != 1 || o.AtRisk != 1 || o.OpenWithinTarget != 0 {
		t.Errorf("overall = %+v", o)
	}

	// Server-side filter: team, canceled excluded, policy labels.
	f := srv.filters[0]
	if f["team"] == nil {
		t.Error("filter missing team")
	}
	canceled, _ := f["canceledAt"].(map[string]any)
	if canceled["null"] != true {
		t.Errorf("canceledAt filter = %v, want {null: true}", f["canceledAt"])
	}
	labelsJSON, _ := json.Marshal(f["labels"])
	if !strings.Contains(string(labelsJSON), `"eqIgnoreCase":"p1"`) || !strings.Contains(string(labelsJSON), `"eqIgnoreCase":"p2"`) {
		t.Errorf("labels filter = %s, want policy labels", labelsJSON)
	}
}

func TestSLOReportCommand_Table(t *testing.T) {
	srv := newSLOMockServer(t)
	out, err := runSLOCommand(t, srv, "--policy", "p1=2d,p2=14d", "--group-by", "assignee", "-o", "table")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 4 {
		t.Fatalf("unexpected table:\n%s", out)
	}
	if fields := strings.Fields(lines[0]); strings.Join(fields, " ") != "GROUP TOTAL MET MISSED AT-RISK OPEN-OK MEDIAN MAX" {
		t.Errorf("header = %q", lines[0])
	}
	wantRows := []string{
		"Alice 1 1 0 0 0 1d 1d",
		"unassigned 2 0 1 1 0 8d12h 12d",
		"TOTAL 3 1 1 1 0 5d 12d",
	}
	for i, want := range wantRows {
		if got := strings.Join(strings.Fields(lines[i+1]), " "); got != want {
			t.Errorf("row %d = %q, want %q", i+1, got, want)
		}
	}
	if !strings.Contains(out, "3 issues evaluated, 1 without a policy label skipped") {
		t.Errorf("missing footer:\n%s", out)
	}
}

func TestSLOReportCommand_PolicyFile(t *testing.T) {
	srv := newSLOMockServer(t)
	path := filepath.Join(t.TempDir(), "policy.txt")
	if err := os.WriteFile(path, []byte("# tiers\np1=2d\np2=14d\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := runSLOCommand(t, srv, "--policy-file", path, "--include-canceled")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var report sloReport
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if report.GroupBy != "none" || len(report.Groups) != 1 || report.Groups[0].Key != "all" || report.Overall.Total != 3 {
		t.Errorf("report = %+v", report)
	}
	if _, ok := srv.filters[0]["canceledAt"]; ok {
		t.Error("--include-canceled should drop the canceledAt filter")
	}
}

func TestSLOReportCommand_UserLabelFilterWins(t *testing.T) {
	srv := newSLOMockServer(t)
	factory := func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(srv.URL))
	}
	cmd := newSLOReportCommand(factory, func() time.Time { return sloTestNow })
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"--policy", "p1=2d", "--label", "bug"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	labelsJSON, _ := json.Marshal(srv.filters[0]["labels"])
	if strings.Contains(string(labelsJSON), "eqIgnoreCase") || !strings.Contains(string(labelsJSON), "label-123") {
		t.Errorf("labels filter = %s, want the --label filter only", labelsJSON)
	}
}

func TestSLOReportCommand_FlagErrors(t *testing.T) {
	srv := newSLOMockServer(t)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "no policy", args: nil, want: "policy"},
		{name: "both policies", args: []string{"--policy", "p1=2d", "--policy-file", "x"}, want: "none of the others can be"},
		{name: "bad policy", args: []string{"--policy", "p1=soon"}, want: "invalid policy entry"},
		{name: "bad group", args: []string{"--policy", "p1=2d", "--group-by", "project"}, want: "invalid --group-by"},
		{name: "bad threshold", args: []string{"--policy", "p1=2d", "--at-risk-threshold", "0"}, want: "at-risk-threshold"},
		{name: "bad output", args: []string{"--policy", "p1=2d", "-o", "yaml"}, want: "invalid --output"},
		{name: "missing file", args: []string{"--policy-file", "/nonexistent/policy"}, want: "failed to open policy file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runSLOCommand(t, srv, tt.args...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
	if srv.pages != 0 {
		t.Errorf("flag errors should not query issues, got %d requests", srv.pages)
	}
}
