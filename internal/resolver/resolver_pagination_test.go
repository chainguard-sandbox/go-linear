package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

// Tests for fetchAllPages and cursor handling in resolver fetches.

func TestFetchAllPages_MultiplePages(t *testing.T) {
	pages := map[string]struct {
		nodes   []string
		hasNext bool
		cursor  string
	}{
		"":   {nodes: []string{"a", "b"}, hasNext: true, cursor: "c1"},
		"c1": {nodes: []string{"c", "d"}, hasNext: true, cursor: "c2"},
		"c2": {nodes: []string{"e"}, hasNext: false, cursor: ""},
	}

	var calls int
	got, err := fetchAllPages(context.Background(), func(_ context.Context, after *string) ([]string, bool, *string, error) {
		calls++
		key := ""
		if after != nil {
			key = *after
		}
		p, ok := pages[key]
		if !ok {
			t.Fatalf("unexpected cursor %q", key)
		}
		return p.nodes, p.hasNext, &p.cursor, nil
	})
	if err != nil {
		t.Fatalf("fetchAllPages() error = %v", err)
	}
	if calls != 3 {
		t.Errorf("fetchAllPages() made %d calls, want 3", calls)
	}
	want := []string{"a", "b", "c", "d", "e"}
	if len(got) != len(want) {
		t.Fatalf("fetchAllPages() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("fetchAllPages()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFetchAllPages_NilEndCursorStops(t *testing.T) {
	var calls int
	got, err := fetchAllPages(context.Background(), func(_ context.Context, after *string) ([]string, bool, *string, error) {
		calls++
		// hasNextPage true but no cursor to advance with - must not loop
		return []string{"a"}, true, nil, nil
	})
	if err != nil {
		t.Fatalf("fetchAllPages() error = %v", err)
	}
	if calls != 1 {
		t.Errorf("fetchAllPages() made %d calls, want 1", calls)
	}
	if len(got) != 1 {
		t.Errorf("fetchAllPages() returned %d nodes, want 1", len(got))
	}
}

func TestFetchAllPages_RepeatedCursorStops(t *testing.T) {
	cursor := "stuck"
	var calls int
	_, err := fetchAllPages(context.Background(), func(_ context.Context, after *string) ([]string, bool, *string, error) {
		calls++
		return []string{"a"}, true, &cursor, nil
	})
	if err != nil {
		t.Fatalf("fetchAllPages() error = %v", err)
	}
	if calls != 2 {
		t.Errorf("fetchAllPages() made %d calls, want 2 (stop when cursor repeats)", calls)
	}
}

func TestFetchAllPages_ErrorPropagates(t *testing.T) {
	wantErr := errors.New("boom")
	cursor := "c1"
	var calls int
	_, err := fetchAllPages(context.Background(), func(_ context.Context, after *string) ([]string, bool, *string, error) {
		calls++
		if calls == 1 {
			return []string{"a"}, true, &cursor, nil
		}
		return nil, false, nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("fetchAllPages() error = %v, want %v", err, wantErr)
	}
}

// TestResolveUser_Pagination verifies that a user beyond the first page is
// found (regression test: ResolveUser previously fetched only the first 250
// users and reported "not found" for anyone past that page).
func TestResolveUser_Pagination(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		after, _ := req.Variables["after"].(string)
		var page map[string]any
		switch after {
		case "":
			page = map[string]any{
				"nodes": []map[string]any{
					{"id": "user-1", "name": "Alice Smith", "email": "alice@example.com", "displayName": "alice"},
					{"id": "user-2", "name": "Bob Jones", "email": "bob@example.com", "displayName": "bob"},
				},
				"pageInfo": map[string]any{"hasNextPage": true, "endCursor": "cursor-1"},
			}
		case "cursor-1":
			page = map[string]any{
				"nodes": []map[string]any{
					{"id": "user-3", "name": "Carol Deep", "email": "carol@example.com", "displayName": "carol"},
				},
				"pageInfo": map[string]any{"hasNextPage": false},
			}
		default:
			t.Errorf("unexpected after cursor %q", after)
			page = map[string]any{
				"nodes":    []map[string]any{},
				"pageInfo": map[string]any{"hasNextPage": false},
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"users": page},
		})
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveUser(ctx, "carol@example.com")
	if err != nil {
		t.Fatalf("ResolveUser() error = %v", err)
	}
	if got != "user-3" {
		t.Errorf("ResolveUser() = %v, want user-3", got)
	}
}

// TestResolveTeam_Pagination verifies the same for teams (first page size 100).
func TestResolveTeam_Pagination(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		after, _ := req.Variables["after"].(string)
		var page map[string]any
		if after == "" {
			page = map[string]any{
				"nodes": []map[string]any{
					{"id": "team-1", "name": "Engineering", "key": "ENG"},
				},
				"pageInfo": map[string]any{"hasNextPage": true, "endCursor": "cursor-1"},
			}
		} else {
			page = map[string]any{
				"nodes": []map[string]any{
					{"id": "team-2", "name": "Platform", "key": "PLT"},
				},
				"pageInfo": map[string]any{"hasNextPage": false},
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"teams": page},
		})
	}

	resolver := newMockResolver(t, handler)
	ctx := context.Background()

	got, err := resolver.ResolveTeam(ctx, "PLT")
	if err != nil {
		t.Fatalf("ResolveTeam() error = %v", err)
	}
	if got != "team-2" {
		t.Errorf("ResolveTeam() = %v, want team-2", got)
	}
}
