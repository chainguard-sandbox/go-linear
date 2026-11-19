package issue

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

// opCapture records the GraphQL variables of every request, keyed by
// operation name (lowercased), so tests can assert on the payload of a
// specific operation in a multi-request command run.
type opCapture struct {
	mu   sync.Mutex
	vars map[string][]map[string]any
}

func (c *opCapture) last(op string) map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	reqs := c.vars[strings.ToLower(op)]
	if len(reqs) == 0 {
		return nil
	}
	return reqs[len(reqs)-1]
}

func (c *opCapture) count(op string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.vars[strings.ToLower(op)])
}

// captureOpsServer serves handlers by operation name and records variables.
func captureOpsServer(t *testing.T, handlers map[string]string) (*httptest.Server, *opCapture) {
	t.Helper()
	c := &opCapture{vars: map[string][]map[string]any{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		c.mu.Lock()
		op := strings.ToLower(body.OperationName)
		c.vars[op] = append(c.vars[op], body.Variables)
		c.mu.Unlock()
		for key, resp := range handlers {
			if strings.EqualFold(key, body.OperationName) {
				_, _ = w.Write([]byte(resp))
				return
			}
		}
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	t.Cleanup(server.Close)
	return server, c
}

func runCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// inputOf returns variables["input"] as a map.
func inputOf(t *testing.T, vars map[string]any) map[string]any {
	t.Helper()
	if vars == nil {
		t.Fatal("operation not captured")
	}
	input, ok := vars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input variable missing or not an object: %v", vars["input"])
	}
	return input
}

// assertNull fails unless m[key] is present and null.
func assertNull(t *testing.T, m map[string]any, key string) {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Errorf("%s missing from input, want explicit null", key)
		return
	}
	if v != nil {
		t.Errorf("%s = %v, want explicit null", key, v)
	}
}

func listFilteredHandlers() map[string]string {
	h := defaultHandlers()
	h["ListIssuesFiltered"] = mockIssuesResponse
	return h
}

func TestRunList_SLABreachesFilters(t *testing.T) {
	t.Run("absolute dates map to gte/lt", func(t *testing.T) {
		server, captured := captureOpsServer(t, listFilteredHandlers())
		cmd := NewListCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "--sla-breaches-after=2026-11-01", "--sla-breaches-before=2026-12-01"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		vars := captured.last("ListIssuesFiltered")
		filter, _ := vars["filter"].(map[string]any)
		cmp, ok := filter["slaBreachesAt"].(map[string]any)
		if !ok {
			t.Fatalf("filter.slaBreachesAt missing: %v", vars["filter"])
		}
		if cmp["gte"] != "2026-11-01T00:00:00.000Z" {
			t.Errorf("slaBreachesAt.gte = %v, want 2026-11-01T00:00:00.000Z", cmp["gte"])
		}
		if cmp["lt"] != "2026-12-01T00:00:00.000Z" {
			t.Errorf("slaBreachesAt.lt = %v, want 2026-12-01T00:00:00.000Z", cmp["lt"])
		}
	})

	t.Run("relative before points into the future", func(t *testing.T) {
		server, captured := captureOpsServer(t, listFilteredHandlers())
		cmd := NewListCommand(testutil.TestFactory(t, server.URL))
		before := time.Now().UTC()
		if _, err := runCmd(t, cmd, "--sla-breaches-before=7d"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		filter, _ := captured.last("ListIssuesFiltered")["filter"].(map[string]any)
		cmp, _ := filter["slaBreachesAt"].(map[string]any)
		lt, err := time.Parse("2006-01-02T15:04:05.000Z", cmp["lt"].(string))
		if err != nil {
			t.Fatalf("lt not parseable: %v", err)
		}
		want := before.Add(7 * 24 * time.Hour)
		if d := lt.Sub(want); d < -time.Minute || d > time.Minute {
			t.Errorf("slaBreachesAt.lt = %v, want ~%v (7 days from now)", lt, want)
		}
		if _, ok := cmp["gte"]; ok {
			t.Error("slaBreachesAt.gte should be omitted when only --sla-breaches-before is set")
		}
	})

	t.Run("invalid date errors before querying", func(t *testing.T) {
		server, captured := captureOpsServer(t, listFilteredHandlers())
		cmd := NewListCommand(testutil.TestFactory(t, server.URL))
		_, err := runCmd(t, cmd, "--sla-breaches-before=soon")
		if err == nil || !strings.Contains(err.Error(), "invalid sla-breaches-before date") {
			t.Fatalf("error = %v, want invalid sla-breaches-before date", err)
		}
		if captured.count("ListIssuesFiltered") != 0 {
			t.Error("no query should be sent for an invalid date")
		}
	})
}

func TestRunUpdate_SLAFields(t *testing.T) {
	t.Run("sets the SLA type", func(t *testing.T) {
		server, captured := captureOpsServer(t, defaultHandlers())
		cmd := NewUpdateCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "ENG-123", "--sla-type=onlyBusinessDays"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		input := inputOf(t, captured.last("UpdateIssue"))
		if input["slaType"] != "onlyBusinessDays" || len(input) != 1 {
			t.Errorf("input = %v, want only slaType=onlyBusinessDays", input)
		}
	})

	t.Run("type all combined with a regular field", func(t *testing.T) {
		server, captured := captureOpsServer(t, defaultHandlers())
		cmd := NewUpdateCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "ENG-123", "--sla-type=all", "--title=New"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		input := inputOf(t, captured.last("UpdateIssue"))
		if input["slaType"] != "all" || input["title"] != "New" {
			t.Errorf("input = %v, want slaType=all and title=New", input)
		}
	})

	for _, v := range []string{"", "none"} {
		t.Run("clear type with "+strconv.Quote(v), func(t *testing.T) {
			server, captured := captureOpsServer(t, defaultHandlers())
			cmd := NewUpdateCommand(testutil.TestFactory(t, server.URL))
			if _, err := runCmd(t, cmd, "ENG-123", "--sla-type="+v); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			input := inputOf(t, captured.last("UpdateIssue"))
			assertNull(t, input, "slaType")
			if len(input) != 1 {
				t.Errorf("input = %v, want only slaType", input)
			}
		})
	}

	invalidTests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"bad type", []string{"--sla-type=weekdays"}, "must be 'all' or 'onlyBusinessDays'"},
		// Linear removed slaBreachesAt/slaStartedAt from IssueUpdateInput.
		{"removed breach flag", []string{"--sla-breaches-at=2026-12-01"}, "--sla-breaches-at was removed from the Linear API"},
		{"removed breach flag, clear", []string{"--sla-breaches-at="}, "--sla-breaches-at was removed from the Linear API"},
		{"removed start flag", []string{"--sla-started-at=2026-11-01", "--sla-type=all"}, "--sla-started-at was removed from the Linear API"},
	}
	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			server, captured := captureOpsServer(t, defaultHandlers())
			cmd := NewUpdateCommand(testutil.TestFactory(t, server.URL))
			_, err := runCmd(t, cmd, append([]string{"ENG-123"}, tt.args...)...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
			if captured.count("UpdateIssue") != 0 || captured.count("GetIssue") != 0 {
				t.Error("no API call should be made for an invalid or removed SLA flag")
			}
		})
	}

	t.Run("removed flags are hidden", func(t *testing.T) {
		cmd := NewUpdateCommand(testutil.TestFactory(t, "http://unused"))
		for _, name := range []string{"sla-breaches-at", "sla-started-at"} {
			if f := cmd.Flags().Lookup(name); f == nil || !f.Hidden {
				t.Errorf("--%s should exist and be hidden", name)
			}
		}
	})
}

func TestRunUpdate_SLAErrorHint(t *testing.T) {
	tests := []struct {
		name     string
		apiError string
		wantHint bool
	}{
		{"API error mentioning SLA gets hint", `{"errors":[{"message":"SLAs are not enabled for this team"}]}`, true},
		{"API error naming an SLA field gets hint", `{"errors":[{"message":"Argument Validation Error","extensions":{"invalidArgs":["slaType"]}}]}`, true},
		{"unrelated API error has no hint", `{"errors":[{"message":"Entity not found"}]}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := defaultHandlers()
			handlers["UpdateIssue"] = tt.apiError
			server, _ := captureOpsServer(t, handlers)
			cmd := NewUpdateCommand(testutil.TestFactory(t, server.URL))
			_, err := runCmd(t, cmd, "ENG-123", "--sla-type=all")
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "failed to update issue: GraphQL error:") {
				t.Errorf("error = %q, want the API error surfaced", err)
			}
			if got := strings.Contains(err.Error(), slaHint); got != tt.wantHint {
				t.Errorf("hint present = %v, want %v; error = %q", got, tt.wantHint, err)
			}
		})
	}
}

func TestWithSLAHint(t *testing.T) {
	if withSLAHint(nil) != nil {
		t.Error("nil error should stay nil")
	}
	base := errors.New("Slack integration missing")
	if got := withSLAHint(base); got.Error() != base.Error() {
		t.Errorf("'Slack' must not trigger the hint, got %q", got)
	}
	wrapped := withSLAHint(errors.New("SLA not available"))
	if !strings.Contains(wrapped.Error(), slaHint) {
		t.Errorf("hint missing: %q", wrapped)
	}
}

func batchHandlers() map[string]string {
	h := defaultHandlers()
	h["ListIssuesFiltered"] = `{"data":{"issues":{"nodes":[
		{"id":"issue-123","identifier":"ENG-123","title":"Test Issue 1","priority":2,"createdAt":"2024-01-01T00:00:00.000Z"},
		{"id":"issue-456","identifier":"ENG-456","title":"Test Issue 2","priority":3,"createdAt":"2024-01-02T00:00:00.000Z"}
	],"pageInfo":{"hasNextPage":false}}}}`
	h["BatchUpdateIssues"] = `{"data":{"issueBatchUpdate":{"success":true,"lastSyncId":1,"issues":[
		{"id":"issue-123","identifier":"ENG-123","title":"Test Issue 1"},
		{"id":"issue-456","identifier":"ENG-456","title":"Test Issue 2"}
	]}}}`
	return h
}

func TestRunBatchUpdate_SLAFields(t *testing.T) {
	t.Run("sets SLA type on every matched issue", func(t *testing.T) {
		server, captured := captureOpsServer(t, batchHandlers())
		cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
		out, err := runCmd(t, cmd, "--team=ENG", "--yes", "--set-sla-type=all", "--set-state=Done")
		if err != nil {
			t.Fatalf("Execute() error = %v\n%s", err, out)
		}

		vars := captured.last("BatchUpdateIssues")
		ids, _ := vars["ids"].([]any)
		if len(ids) != 2 || ids[0] != "issue-123" || ids[1] != "issue-456" {
			t.Errorf("ids = %v, want [issue-123 issue-456]", vars["ids"])
		}
		input := inputOf(t, vars)
		want := map[string]any{"slaType": "all", "stateId": "state-789"}
		for k, v := range want {
			if input[k] != v {
				t.Errorf("input[%s] = %v, want %v", k, input[k], v)
			}
		}
		if len(input) != len(want) {
			t.Errorf("input = %v, want exactly %v", input, want)
		}
		if !strings.Contains(out, "issue-456") {
			t.Errorf("output should include updated issues, got %s", out)
		}
	})

	t.Run("none clears the SLA type", func(t *testing.T) {
		server, captured := captureOpsServer(t, batchHandlers())
		cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "--team=ENG", "--yes", "--set-sla-type=none"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		input := inputOf(t, captured.last("BatchUpdateIssues"))
		assertNull(t, input, "slaType")
		if len(input) != 1 {
			t.Errorf("input = %v, want only slaType", input)
		}
	})

	t.Run("keeps other clears as empty strings", func(t *testing.T) {
		server, captured := captureOpsServer(t, batchHandlers())
		cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "--team=ENG", "--yes", "--set-sla-type=", "--set-cycle=none", "--set-estimate=3"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		input := inputOf(t, captured.last("BatchUpdateIssues"))
		if v, ok := input["cycleId"]; !ok || v != "" {
			t.Errorf("cycleId = %v (present=%v), want empty string as on the regular path", v, ok)
		}
		if input["estimate"] != float64(3) {
			t.Errorf("estimate = %v, want 3", input["estimate"])
		}
		assertNull(t, input, "slaType")
	})

	t.Run("stays within the 50-issue batch limit", func(t *testing.T) {
		server, captured := captureOpsServer(t, batchHandlers())
		cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "--team=ENG", "--yes", "--batch-limit=500", "--set-sla-type=all"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if first := captured.last("ListIssuesFiltered")["first"]; first != float64(50) {
			t.Errorf("ListIssuesFiltered first = %v, want 50", first)
		}
		if captured.count("BatchUpdateIssues") != 1 {
			t.Errorf("BatchUpdateIssues calls = %d, want 1", captured.count("BatchUpdateIssues"))
		}
	})

	t.Run("filters by SLA breach date", func(t *testing.T) {
		server, captured := captureOpsServer(t, batchHandlers())
		cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "--sla-breaches-before=2026-12-01", "--yes", "--set-sla-type=onlyBusinessDays"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		filter, _ := captured.last("ListIssuesFiltered")["filter"].(map[string]any)
		cmp, _ := filter["slaBreachesAt"].(map[string]any)
		if cmp["lt"] != "2026-12-01T00:00:00.000Z" {
			t.Errorf("filter.slaBreachesAt = %v, want lt 2026-12-01T00:00:00.000Z", filter["slaBreachesAt"])
		}
	})

	t.Run("dry run sends no mutation", func(t *testing.T) {
		server, captured := captureOpsServer(t, batchHandlers())
		cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
		out, err := runCmd(t, cmd, "--team=ENG", "--dry-run", "--set-sla-type=all")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(out, "Would update 2 issues") {
			t.Errorf("dry-run output = %q", out)
		}
		if captured.count("BatchUpdateIssues") != 0 {
			t.Error("dry run must not send the mutation")
		}
	})

	for _, tt := range []struct {
		name    string
		arg     string
		wantErr string
	}{
		{"invalid type", "--set-sla-type=calendar", "invalid --set-sla-type"},
		{"removed breach flag", "--set-sla-breaches-at=2026-12-01", "--set-sla-breaches-at was removed from the Linear API"},
		{"removed start flag", "--set-sla-started-at=", "--set-sla-started-at was removed from the Linear API"},
	} {
		t.Run(tt.name+" fails before any API call", func(t *testing.T) {
			server, captured := captureOpsServer(t, batchHandlers())
			cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
			_, err := runCmd(t, cmd, "--team=ENG", "--yes", tt.arg)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
			if captured.count("ListIssuesFiltered") != 0 {
				t.Error("no API call should be made for an invalid or removed SLA flag")
			}
		})
	}

	t.Run("SLA API error is surfaced with hint", func(t *testing.T) {
		handlers := batchHandlers()
		handlers["BatchUpdateIssues"] = `{"errors":[{"message":"SLAs are not enabled"}]}`
		server, _ := captureOpsServer(t, handlers)
		cmd := NewBatchUpdateCommand(testutil.TestFactory(t, server.URL))
		_, err := runCmd(t, cmd, "--team=ENG", "--yes", "--set-sla-type=all")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "SLAs are not enabled") || !strings.Contains(err.Error(), slaHint) {
			t.Errorf("error = %q, want API message and hint", err)
		}
	})
}

func TestRunList_SortBySLAStatus(t *testing.T) {
	sortOf := func(t *testing.T, vars map[string]any) map[string]any {
		t.Helper()
		list, ok := vars["sort"].([]any)
		if !ok || len(list) != 1 {
			t.Fatalf("sort = %v, want one criterion", vars["sort"])
		}
		item, _ := list[0].(map[string]any)
		sla, ok := item["slaStatus"].(map[string]any)
		if !ok {
			t.Fatalf("sort[0] = %v, want slaStatus", item)
		}
		return sla
	}

	for _, tt := range []struct {
		name  string
		args  []string
		order string
	}{
		{"default ascending, no filters", []string{"--sort=sla-status"}, "Ascending"},
		{"descending with a filter", []string{"--sort=sla-status:desc", "--priority=1"}, "Descending"},
		{"explicit asc", []string{"--sort=sla-status:asc"}, "Ascending"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server, captured := captureOpsServer(t, listFilteredHandlers())
			cmd := NewListCommand(testutil.TestFactory(t, server.URL))
			if _, err := runCmd(t, cmd, tt.args...); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if captured.count("ListIssues") != 0 {
				t.Error("--sort must use the filtered query, not ListIssues")
			}
			sla := sortOf(t, captured.last("ListIssuesFiltered"))
			if sla["order"] != tt.order {
				t.Errorf("slaStatus.order = %v, want %s", sla["order"], tt.order)
			}
			if sla["nulls"] != "last" {
				t.Errorf("slaStatus.nulls = %v, want last", sla["nulls"])
			}
		})
	}

	t.Run("no sort sends null", func(t *testing.T) {
		server, captured := captureOpsServer(t, listFilteredHandlers())
		cmd := NewListCommand(testutil.TestFactory(t, server.URL))
		if _, err := runCmd(t, cmd, "--priority=1"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		vars := captured.last("ListIssuesFiltered")
		if v, ok := vars["sort"]; ok && v != nil {
			t.Errorf("sort = %v, want null or absent", v)
		}
	})

	for _, bad := range []string{"--sort=priority", "--sort=sla-status:up"} {
		t.Run("invalid "+bad, func(t *testing.T) {
			server, captured := captureOpsServer(t, listFilteredHandlers())
			cmd := NewListCommand(testutil.TestFactory(t, server.URL))
			if _, err := runCmd(t, cmd, bad); err == nil || !strings.Contains(err.Error(), "invalid --sort") {
				t.Fatalf("error = %v, want invalid --sort", err)
			}
			if captured.count("ListIssuesFiltered")+captured.count("ListIssues") != 0 {
				t.Error("no query should be sent for an invalid --sort")
			}
		})
	}
}
