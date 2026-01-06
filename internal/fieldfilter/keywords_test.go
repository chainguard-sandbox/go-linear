package fieldfilter

import (
	"testing"
)

func TestNewWithDefaults(t *testing.T) {
	defaults := []string{"id", "title", "url"}

	t.Run("defaults keyword", func(t *testing.T) {
		fs, err := New("defaults", defaults)
		if err != nil {
			t.Fatalf("New('defaults') error: %v", err)
		}
		if fs == nil {
			t.Fatal("defaults keyword should return selector with defaults")
		}
		// Check it has the default fields
		if !fs.fields["id"] || !fs.fields["title"] || !fs.fields["url"] {
			t.Error("defaults keyword should include all default fields")
		}
	})

	t.Run("none keyword", func(t *testing.T) {
		fs, err := New("none", defaults)
		if err != nil {
			t.Fatalf("New('none') error: %v", err)
		}
		if fs != nil {
			t.Error("none keyword should return nil (no filtering)")
		}
	})

	t.Run("defaults with extra", func(t *testing.T) {
		fs, err := New("defaults,body,description", defaults)
		if err != nil {
			t.Fatalf("New('defaults,body,description') error: %v", err)
		}
		if fs == nil {
			t.Fatal("should return selector")
		}
		// Should have: id, title, url (defaults) + body, description (extras)
		if !fs.fields["id"] || !fs.fields["body"] || !fs.fields["description"] {
			t.Error("defaults,extra should include defaults + extras")
		}
		if len(fs.fields) != 5 {
			t.Errorf("defaults,extra should have 5 fields, got %d", len(fs.fields))
		}
	})

	t.Run("deduplication", func(t *testing.T) {
		fs, err := New("defaults,title,body", defaults) // title already in defaults
		if err != nil {
			t.Fatalf("New with duplicate error: %v", err)
		}
		if fs == nil {
			t.Fatal("should return selector")
		}
		if len(fs.fields) != 4 {
			t.Errorf("Deduplication should give 4 unique fields, got %d", len(fs.fields))
		}
	})

	t.Run("custom fields", func(t *testing.T) {
		fs, err := New("id,body", defaults)
		if err != nil {
			t.Fatalf("New('id,body') error: %v", err)
		}
		if fs == nil {
			t.Fatal("should return selector")
		}
		if len(fs.fields) != 2 {
			t.Errorf("Custom fields should override defaults, got %d fields", len(fs.fields))
		}
		if fs.fields["title"] {
			t.Error("Custom fields should NOT include defaults")
		}
	})

	t.Run("empty with defaults", func(t *testing.T) {
		fs, err := New("", defaults)
		if err != nil {
			t.Fatalf("New('') error: %v", err)
		}
		if fs == nil {
			t.Fatal("Empty string with defaults should apply defaults")
		}
		if len(fs.fields) != 3 {
			t.Errorf("Empty string should use defaults, got %d fields", len(fs.fields))
		}
	})

	t.Run("empty with no defaults", func(t *testing.T) {
		fs, err := New("", nil)
		if err != nil {
			t.Fatalf("New('', nil) error: %v", err)
		}
		if fs != nil {
			t.Error("Empty string with no defaults should return nil")
		}
	})
}

func TestNewForList(t *testing.T) {
	defaults := []string{"id", "title"}

	t.Run("preserves nodes and pageInfo", func(t *testing.T) {
		fs, err := NewForList("defaults", defaults)
		if err != nil {
			t.Fatalf("NewForList error: %v", err)
		}
		if fs == nil {
			t.Fatal("NewForList should return selector")
		}
		if fs.preserveFields == nil {
			t.Fatal("NewForList should set preserveFields")
		}
		if !fs.preserveFields["nodes"] || !fs.preserveFields["pageInfo"] || !fs.preserveFields["totalCount"] {
			t.Error("NewForList should preserve nodes, pageInfo, totalCount")
		}
	})

	t.Run("none returns nil", func(t *testing.T) {
		fs, err := NewForList("none", defaults)
		if err != nil {
			t.Fatalf("NewForList('none') error: %v", err)
		}
		if fs != nil {
			t.Error("none should return nil even for list")
		}
	})
}
