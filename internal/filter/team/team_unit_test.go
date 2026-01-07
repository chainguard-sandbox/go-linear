package team

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
		teamName   string
		wantFilter bool
	}{
		{
			name:       "no name",
			wantFilter: false,
		},
		{
			name:       "with name",
			teamName:   "Engineering",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("name", tt.teamName, "")

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

func TestApplyKey(t *testing.T) {
	tests := []struct {
		name       string
		teamKey    string
		wantFilter bool
	}{
		{
			name:       "no key",
			wantFilter: false,
		},
		{
			name:       "with key",
			teamKey:    "ENG",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("key", tt.teamKey, "")

			b := NewFilterBuilder(nil)
			err := ApplyKey(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyKey() error = %v", err)
			}

			hasFilter := b.filter.Key != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyKey() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
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
			desc:       "platform",
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

func TestApplyPrivate(t *testing.T) {
	tests := []struct {
		name       string
		private    bool
		setFlag    bool
		wantFilter bool
	}{
		{
			name:       "flag not set",
			setFlag:    false,
			wantFilter: false,
		},
		{
			name:       "private true",
			private:    true,
			setFlag:    true,
			wantFilter: true,
		},
		{
			name:       "private false",
			private:    false,
			setFlag:    true,
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("private", false, "")
			if tt.setFlag {
				_ = cmd.Flags().Set("private", boolToString(tt.private))
			}

			b := NewFilterBuilder(nil)
			err := ApplyPrivate(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyPrivate() error = %v", err)
			}

			hasFilter := b.filter.Private != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyPrivate() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
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
	cmd.Flags().String("name", "Engineering", "")
	cmd.Flags().String("key", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().Bool("private", false, "")

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

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
