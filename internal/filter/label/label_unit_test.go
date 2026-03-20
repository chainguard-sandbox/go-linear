package label

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/filter/common"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
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
		labelName  string
		wantFilter bool
	}{
		{
			name:       "no name",
			wantFilter: false,
		},
		{
			name:       "with name",
			labelName:  "bug",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("name", tt.labelName, "")

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

func TestApplyIsGroup(t *testing.T) {
	tests := []struct {
		name       string
		isGroup    bool
		setFlag    bool
		wantFilter bool
	}{
		{
			name:       "flag not set",
			setFlag:    false,
			wantFilter: false,
		},
		{
			name:       "is-group true",
			isGroup:    true,
			setFlag:    true,
			wantFilter: true,
		},
		{
			name:       "is-group false",
			isGroup:    false,
			setFlag:    true,
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("is-group", false, "")
			if tt.setFlag {
				_ = cmd.Flags().Set("is-group", boolToString(tt.isGroup))
			}

			b := NewFilterBuilder(nil)
			err := ApplyIsGroup(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyIsGroup() error = %v", err)
			}

			hasFilter := b.filter.IsGroup != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyIsGroup() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
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
	cmd.Flags().String("creator", "", "")
	cmd.Flags().String("team", "", "")
	cmd.Flags().String("name", "bug", "")
	cmd.Flags().Bool("is-group", false, "")

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
