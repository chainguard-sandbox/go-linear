package user

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
		userName   string
		wantFilter bool
	}{
		{
			name:       "no name",
			wantFilter: false,
		},
		{
			name:       "with name",
			userName:   "Alice",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("name", tt.userName, "")

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

func TestApplyEmail(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		wantFilter bool
	}{
		{
			name:       "no email",
			wantFilter: false,
		},
		{
			name:       "with email",
			email:      "alice@example.com",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("email", tt.email, "")

			b := NewFilterBuilder(nil)
			err := ApplyEmail(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyEmail() error = %v", err)
			}

			hasFilter := b.filter.Email != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyEmail() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantFilter  bool
	}{
		{
			name:       "no display name",
			wantFilter: false,
		},
		{
			name:        "with display name",
			displayName: "Ally",
			wantFilter:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("display-name", tt.displayName, "")

			b := NewFilterBuilder(nil)
			err := ApplyDisplayName(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyDisplayName() error = %v", err)
			}

			hasFilter := b.filter.DisplayName != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyDisplayName() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyActive(t *testing.T) {
	tests := []struct {
		name       string
		active     bool
		setFlag    bool
		wantFilter bool
	}{
		{
			name:       "flag not set",
			setFlag:    false,
			wantFilter: false,
		},
		{
			name:       "active true",
			active:     true,
			setFlag:    true,
			wantFilter: true,
		},
		{
			name:       "active false",
			active:     false,
			setFlag:    true,
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("active", false, "")
			if tt.setFlag {
				_ = cmd.Flags().Set("active", boolToString(tt.active))
			}

			b := NewFilterBuilder(nil)
			err := ApplyActive(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyActive() error = %v", err)
			}

			hasFilter := b.filter.Active != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyActive() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyAdmin(t *testing.T) {
	tests := []struct {
		name       string
		admin      bool
		setFlag    bool
		wantFilter bool
	}{
		{
			name:       "flag not set",
			setFlag:    false,
			wantFilter: false,
		},
		{
			name:       "admin true",
			admin:      true,
			setFlag:    true,
			wantFilter: true,
		},
		{
			name:       "admin false",
			admin:      false,
			setFlag:    true,
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("admin", false, "")
			if tt.setFlag {
				_ = cmd.Flags().Set("admin", boolToString(tt.admin))
			}

			b := NewFilterBuilder(nil)
			err := ApplyAdmin(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyAdmin() error = %v", err)
			}

			hasFilter := b.filter.Admin != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyAdmin() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyIsMe(t *testing.T) {
	tests := []struct {
		name       string
		isMe       bool
		setFlag    bool
		wantFilter bool
	}{
		{
			name:       "flag not set",
			setFlag:    false,
			wantFilter: false,
		},
		{
			name:       "is-me true",
			isMe:       true,
			setFlag:    true,
			wantFilter: true,
		},
		{
			name:       "is-me false",
			isMe:       false,
			setFlag:    true,
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("is-me", false, "")
			if tt.setFlag {
				_ = cmd.Flags().Set("is-me", boolToString(tt.isMe))
			}

			b := NewFilterBuilder(nil)
			err := ApplyIsMe(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyIsMe() error = %v", err)
			}

			hasFilter := b.filter.IsMe != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyIsMe() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
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
	cmd.Flags().String("name", "Alice", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("display-name", "", "")
	cmd.Flags().Bool("active", false, "")
	cmd.Flags().Bool("admin", false, "")
	cmd.Flags().Bool("is-me", false, "")

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
