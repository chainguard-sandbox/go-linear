package linear

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// planningClient serves body for the named operation and fails the test if
// any other operation is sent, or if the page size and cursor are not passed
// through.
func planningClient(t *testing.T, operation, body string) *Client {
	t.Helper()
	return newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
		}
		var req struct {
			OperationName string `json:"operationName"`
			Query         string `json:"query"`
			Variables     struct {
				First *int64  `json:"first"`
				After *string `json:"after"`
			} `json:"variables"`
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if req.OperationName != operation {
			t.Errorf("operation = %q, want %q", req.OperationName, operation)
		}
		if strings.Contains(req.Query, "mutation") {
			t.Errorf("%s sends a mutation", operation)
		}
		if req.Variables.First == nil || *req.Variables.First != 25 {
			t.Errorf("first = %v, want 25", req.Variables.First)
		}
		if req.Variables.After == nil || *req.Variables.After != "cursor-1" {
			t.Errorf("after = %v, want cursor-1", req.Variables.After)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
}

func planningPage() (*int64, *string) {
	first, after := int64(25), "cursor-1"
	return &first, &after
}

func TestClient_PlanningProjects(t *testing.T) {
	client := planningClient(t, "PlanningProjects", `{"data":{"projects":{"nodes":[
		{"id":"p1","name":"Q1","status":{"type":"started"},"startedAt":"2026-01-05T09:00:00.000Z","completedAt":null,"canceledAt":null,"targetDate":"2026-03-31","updatedAt":"2026-02-01T00:00:00.000Z","lastUpdate":{"createdAt":"2026-01-20T00:00:00.000Z","health":"atRisk"},"teams":{"nodes":[{"id":"t1"},{"id":"t2"}],"pageInfo":{"hasNextPage":true}}},
		{"id":"p2","name":"Q2","status":{"type":"backlog"},"startedAt":null,"completedAt":null,"canceledAt":null,"targetDate":null,"updatedAt":"2026-02-01T00:00:00.000Z","lastUpdate":null,"teams":{"nodes":[],"pageInfo":{"hasNextPage":false}}}
	],"pageInfo":{"hasNextPage":true,"endCursor":"cursor-2"}}}}`)

	first, after := planningPage()
	projects, err := client.PlanningProjects(context.Background(), first, after)
	if err != nil {
		t.Fatalf("PlanningProjects() error = %v", err)
	}
	if len(projects.Nodes) != 2 {
		t.Fatalf("len(Nodes) = %d, want 2", len(projects.Nodes))
	}
	if !projects.PageInfo.HasNextPage || projects.PageInfo.EndCursor == nil || *projects.PageInfo.EndCursor != "cursor-2" {
		t.Errorf("PageInfo = %+v, want another page at cursor-2", projects.PageInfo)
	}

	started := projects.Nodes[0]
	if started.Status.Type != "started" {
		t.Errorf("Status.Type = %q, want %q", started.Status.Type, "started")
	}
	if started.StartedAt == nil || started.StartedAt.UTC().Format("2006-01-02") != "2026-01-05" {
		t.Errorf("StartedAt = %v, want 2026-01-05", started.StartedAt)
	}
	if started.CompletedAt != nil || started.CanceledAt != nil {
		t.Errorf("CompletedAt = %v, CanceledAt = %v, want neither", started.CompletedAt, started.CanceledAt)
	}
	if started.TargetDate == nil || *started.TargetDate != "2026-03-31" {
		t.Errorf("TargetDate = %v, want 2026-03-31", started.TargetDate)
	}
	if started.LastUpdate == nil || started.LastUpdate.Health != "atRisk" {
		t.Errorf("LastUpdate = %+v, want health atRisk", started.LastUpdate)
	}
	if len(started.Teams.Nodes) != 2 || started.Teams.Nodes[1].ID != "t2" || !started.Teams.PageInfo.HasNextPage {
		t.Errorf("Teams = %+v, want t1 and t2 with more to come", started.Teams)
	}

	backlog := projects.Nodes[1]
	if backlog.StartedAt != nil || backlog.TargetDate != nil || backlog.LastUpdate != nil {
		t.Errorf("unstarted project = %+v, want no start time, target date or update", backlog)
	}
}

func TestClient_PlanningInitiatives(t *testing.T) {
	client := planningClient(t, "PlanningInitiatives", `{"data":{"initiatives":{"nodes":[
		{"id":"i1","name":"Program","status":"Active","visibility":"public","targetDate":null,"updatedAt":"2026-02-01T00:00:00.000Z","parentInitiative":null,"subInitiatives":{"nodes":[{"id":"i2"}]},"lastUpdate":null},
		{"id":"i2","name":"Phase 1","status":"Planned","visibility":"restricted","targetDate":"2026-06-30","updatedAt":"2026-02-02T00:00:00.000Z","parentInitiative":{"id":"i1"},"subInitiatives":{"nodes":[]},"lastUpdate":{"createdAt":"2026-01-20T00:00:00.000Z","health":"onTrack"}}
	],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`)

	first, after := planningPage()
	initiatives, err := client.PlanningInitiatives(context.Background(), first, after)
	if err != nil {
		t.Fatalf("PlanningInitiatives() error = %v", err)
	}
	if len(initiatives.Nodes) != 2 {
		t.Fatalf("len(Nodes) = %d, want 2", len(initiatives.Nodes))
	}
	parent, child := initiatives.Nodes[0], initiatives.Nodes[1]
	if parent.Status != "Active" || parent.Visibility != "public" || parent.ParentInitiative != nil || len(parent.SubInitiatives.Nodes) != 1 {
		t.Errorf("parent = %+v, want an Active public initiative with a sub-initiative and no parent", parent)
	}
	if child.Visibility != "restricted" || child.ParentInitiative == nil || child.ParentInitiative.ID != "i1" || len(child.SubInitiatives.Nodes) != 0 {
		t.Errorf("child = %+v, want a restricted initiative under i1 with no sub-initiative", child)
	}
	if child.TargetDate == nil || *child.TargetDate != "2026-06-30" || child.LastUpdate == nil || child.LastUpdate.Health != "onTrack" {
		t.Errorf("child = %+v, want target 2026-06-30 and an onTrack update", child)
	}
}

func TestClient_PlanningTeams(t *testing.T) {
	client := planningClient(t, "PlanningTeams", `{"data":{"teams":{"nodes":[
		{"id":"t1","visibility":"public","updatedAt":"2026-02-01T00:00:00.000Z"},
		{"id":"t2","visibility":"private","updatedAt":"2026-02-02T00:00:00.000Z"}
	],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`)

	first, after := planningPage()
	teams, err := client.PlanningTeams(context.Background(), first, after)
	if err != nil {
		t.Fatalf("PlanningTeams() error = %v", err)
	}
	if len(teams.Nodes) != 2 || teams.Nodes[0].Visibility != "public" || teams.Nodes[1].Visibility != "private" {
		t.Errorf("Nodes = %+v, want a public and a private team", teams.Nodes)
	}
	if teams.Nodes[1].UpdatedAt.UTC().Format("2006-01-02") != "2026-02-02" {
		t.Errorf("UpdatedAt = %v, want 2026-02-02", teams.Nodes[1].UpdatedAt)
	}
}

func TestClient_PlanningProjectRelations(t *testing.T) {
	client := planningClient(t, "PlanningProjectRelations", `{"data":{"projectRelations":{"nodes":[
		{"id":"r1","type":"dependency","anchorType":"end","relatedAnchorType":"milestone","updatedAt":"2026-02-01T00:00:00.000Z","project":{"id":"p1"},"relatedProject":{"id":"p2"},"projectMilestone":null,"relatedProjectMilestone":{"id":"m1"}}
	],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`)

	first, after := planningPage()
	relations, err := client.PlanningProjectRelations(context.Background(), first, after)
	if err != nil {
		t.Fatalf("PlanningProjectRelations() error = %v", err)
	}
	if len(relations.Nodes) != 1 {
		t.Fatalf("len(Nodes) = %d, want 1", len(relations.Nodes))
	}
	r := relations.Nodes[0]
	if r.Type != "dependency" || r.AnchorType != "end" || r.RelatedAnchorType != "milestone" || r.Project.ID != "p1" || r.RelatedProject.ID != "p2" {
		t.Errorf("relation = %+v, want a dependency from the end of p1 to a milestone of p2", r)
	}
	if r.ProjectMilestone != nil || r.RelatedProjectMilestone == nil || r.RelatedProjectMilestone.ID != "m1" {
		t.Errorf("milestones = %+v, %+v; want none and m1", r.ProjectMilestone, r.RelatedProjectMilestone)
	}
}

func TestClient_PlanningProjectMilestones(t *testing.T) {
	client := planningClient(t, "PlanningProjectMilestones", `{"data":{"projectMilestones":{"nodes":[
		{"id":"m1","name":"Beta","targetDate":"2026-03-01","status":"next","updatedAt":"2026-02-01T00:00:00.000Z","project":{"id":"p1"}},
		{"id":"m2","name":"GA","targetDate":null,"status":"unstarted","updatedAt":"2026-02-01T00:00:00.000Z","project":{"id":"p1"}}
	],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`)

	first, after := planningPage()
	milestones, err := client.PlanningProjectMilestones(context.Background(), first, after)
	if err != nil {
		t.Fatalf("PlanningProjectMilestones() error = %v", err)
	}
	if len(milestones.Nodes) != 2 {
		t.Fatalf("len(Nodes) = %d, want 2", len(milestones.Nodes))
	}
	next, later := milestones.Nodes[0], milestones.Nodes[1]
	if next.Status != "next" || next.Project.ID != "p1" || next.TargetDate == nil || *next.TargetDate != "2026-03-01" {
		t.Errorf("first milestone = %+v, want the next milestone of p1, due 2026-03-01", next)
	}
	if later.Status != "unstarted" || later.TargetDate != nil {
		t.Errorf("second milestone = %+v, want an unstarted milestone with no date", later)
	}
}

func TestClient_PlanningQueriesReportErrors(t *testing.T) {
	calls := map[string]func(*Client) error{
		"PlanningProjects": func(c *Client) error {
			_, err := c.PlanningProjects(context.Background(), nil, nil)
			return err
		},
		"PlanningInitiatives": func(c *Client) error {
			_, err := c.PlanningInitiatives(context.Background(), nil, nil)
			return err
		},
		"PlanningTeams": func(c *Client) error {
			_, err := c.PlanningTeams(context.Background(), nil, nil)
			return err
		},
		"PlanningProjectRelations": func(c *Client) error {
			_, err := c.PlanningProjectRelations(context.Background(), nil, nil)
			return err
		},
		"PlanningProjectMilestones": func(c *Client) error {
			_, err := c.PlanningProjectMilestones(context.Background(), nil, nil)
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"errors":[{"message":"no access","extensions":{"code":"FORBIDDEN"}}]}`))
			})
			if err := call(client); err == nil {
				t.Errorf("%s() error = nil, want an error", name)
			}
		})
	}
}
