package state

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter/common"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

func TestNewFilterBuilder(t *testing.T) {
	b := NewFilterBuilder(nil)
	if b == nil {
		t.Fatal("NewFilterBuilder() returned nil")
	}
	if b.Filter() == nil {
		t.Fatal("Filter() returned nil")
	}
}

func TestFilterBuilder_Build_Empty(t *testing.T) {
	b := NewFilterBuilder(nil)
	filter := b.Build()
	if filter != nil {
		t.Errorf("Build() = %v, want nil for empty filter", filter)
	}
}

func TestApplyCreatedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
		wantErr    bool
	}{
		{
			name:       "no flags",
			wantFilter: false,
		},
		{
			name:       "after only",
			after:      "2025-01-01",
			wantFilter: true,
		},
		{
			name:       "before only",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:       "both flags",
			after:      "2025-01-01",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:    "invalid date",
			after:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("created-after", tt.after, "")
			cmd.Flags().String("created-before", tt.before, "")

			b := NewFilterBuilder(nil)
			err := common.ApplyCreatedAt(context.Background(), cmd, b)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyCreatedAt() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				hasFilter := b.filter.CreatedAt != nil
				if hasFilter != tt.wantFilter {
					t.Errorf("ApplyCreatedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
				}
			}
		})
	}
}

func TestApplyUpdatedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
		wantErr    bool
	}{
		{
			name:       "no flags",
			wantFilter: false,
		},
		{
			name:       "after only",
			after:      "2025-01-01",
			wantFilter: true,
		},
		{
			name:       "before only",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:    "invalid date",
			before:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("updated-after", tt.after, "")
			cmd.Flags().String("updated-before", tt.before, "")

			b := NewFilterBuilder(nil)
			err := common.ApplyUpdatedAt(context.Background(), cmd, b)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyUpdatedAt() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				hasFilter := b.filter.UpdatedAt != nil
				if hasFilter != tt.wantFilter {
					t.Errorf("ApplyUpdatedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
				}
			}
		})
	}
}

func TestApplyID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantFilter bool
	}{
		{
			name:       "no id",
			wantFilter: false,
		},
		{
			name:       "with id",
			id:         "abc-123",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("id", tt.id, "")

			b := NewFilterBuilder(nil)
			err := common.ApplyID(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyID() error = %v", err)
			}

			hasFilter := b.filter.ID != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyID() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyName(t *testing.T) {
	tests := []struct {
		name       string
		stateName  string
		wantFilter bool
	}{
		{
			name:       "no name",
			wantFilter: false,
		},
		{
			name:       "with name",
			stateName:  "In Progress",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("name", tt.stateName, "")

			b := NewFilterBuilder(nil)
			err := ApplyName(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyName() error = %v", err)
			}

			hasFilter := b.filter.Name != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyName() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyDescription(t *testing.T) {
	tests := []struct {
		name       string
		desc       string
		wantFilter bool
	}{
		{
			name:       "no description",
			wantFilter: false,
		},
		{
			name:       "with description",
			desc:       "work in progress",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("description", tt.desc, "")

			b := NewFilterBuilder(nil)
			err := ApplyDescription(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyDescription() error = %v", err)
			}

			hasFilter := b.filter.Description != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyDescription() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyType(t *testing.T) {
	tests := []struct {
		name       string
		stateType  string
		wantFilter bool
	}{
		{
			name:       "no type",
			wantFilter: false,
		},
		{
			name:       "with type",
			stateType:  "started",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("type", tt.stateType, "")

			b := NewFilterBuilder(nil)
			err := ApplyType(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyType() error = %v", err)
			}

			hasFilter := b.filter.Type != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyType() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyTeam(t *testing.T) {
	tests := []struct {
		name       string
		team       string
		wantFilter bool
	}{
		{
			name:       "no team",
			wantFilter: false,
		},
		{
			name:       "with team",
			team:       "ENG",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("team", tt.team, "")

			b := NewFilterBuilder(nil)
			err := ApplyTeam(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyTeam() error = %v", err)
			}

			hasFilter := b.filter.Team != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyTeam() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestFilterBuilder_FromFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("created-after", "2025-01-01", "")
	cmd.Flags().String("created-before", "", "")
	cmd.Flags().String("updated-after", "", "")
	cmd.Flags().String("updated-before", "", "")
	cmd.Flags().String("id", "", "")
	cmd.Flags().String("name", "In Progress", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("type", "", "")
	cmd.Flags().String("team", "", "")

	b := NewFilterBuilder(nil)
	err := b.FromFlags(context.Background(), cmd)
	if err != nil {
		t.Fatalf("FromFlags() error = %v", err)
	}

	filter := b.Build()
	if filter == nil {
		t.Fatal("Build() returned nil, expected filter")
	}
	if filter.CreatedAt == nil {
		t.Error("expected CreatedAt filter to be set")
	}
	if filter.Name == nil {
		t.Error("expected Name filter to be set")
	}
}

func TestFilterBuilder_Build_WithFilters(t *testing.T) {
	b := NewFilterBuilder(nil)
	name := "test"
	b.filter.Name = &intgraphql.StringComparator{ContainsIgnoreCase: &name}

	filter := b.Build()
	if filter == nil {
		t.Fatal("Build() returned nil, expected filter")
	}
}
