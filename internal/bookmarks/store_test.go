package bookmarks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	store := NewStore(path)

	// Add entries
	if err := store.Add(Entry{Name: "Vacation", Tags: "pto leave", URL: "https://example.com/vacation"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Add(Entry{Name: "Time Entry", Tags: "time hr portal", URL: "https://example.com/time"}); err != nil {
		t.Fatal(err)
	}

	// Load and verify
	entries, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "Vacation" {
		t.Errorf("expected name 'Vacation', got %q", entries[0].Name)
	}
	if entries[1].URL != "https://example.com/time" {
		t.Errorf("expected URL 'https://example.com/time', got %q", entries[1].URL)
	}
}

func TestSearch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	store := NewStore(path)
	store.Add(Entry{Name: "Vacation balance", Tags: "pto leave balance request", URL: "https://example.com/vacation"})
	store.Add(Entry{Name: "Enter time off", Tags: "submit vacation sick time hr portal", URL: "https://example.com/time"})
	store.Add(Entry{Name: "Jenkins", Tags: "deploy pipeline ci cd", URL: "https://example.com/jenkins"})

	tests := []struct {
		query string
		want  int
	}{
		{"vacation", 2}, // matches both by name/tags
		{"pto", 1},
		{"deploy", 1},
		{"portal", 1},
		{"", 3},
	}

	for _, tt := range tests {
		results, err := store.Search(tt.query)
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != tt.want {
			t.Errorf("Search(%q): got %d results, want %d", tt.query, len(results), tt.want)
		}
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	store := NewStore(path)
	store.Add(Entry{Name: "Test", Tags: "test", URL: "https://example.com"})

	removed, err := store.Remove("test") // case-insensitive
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Error("expected removal to succeed")
	}

	entries, _ := store.Load()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after removal, got %d", len(entries))
	}
}

func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	store := NewStore(path)
	store.Add(Entry{Name: "Test", Tags: "old tags", URL: "https://old.example.com"})

	updated, err := store.Update("Test", Entry{Name: "Test Updated", Tags: "new tags", URL: "https://new.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Error("expected update to succeed")
	}

	entries, _ := store.Load()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "Test Updated" {
		t.Errorf("expected name 'Test Updated', got %q", entries[0].Name)
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	store := NewStore("/nonexistent/path/bookmarks.md")
	entries, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if entries != nil {
		t.Errorf("expected nil entries for nonexistent file, got %v", entries)
	}
}

func TestURLsWithSpecialChars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	store := NewStore(path)

	// URL with anchors, query params, etc.
	complexURL := "https://portal.example.com/app/ui5/shells/abap/Launchpad.html#LeaveRequest-create"
	store.Add(Entry{Name: "Leave Request", Tags: "pto vacation", URL: complexURL})

	entries, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].URL != complexURL {
		t.Errorf("URL mangled during round-trip:\n  got:  %s\n  want: %s", entries[0].URL, complexURL)
	}
}

func TestFileContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	store := NewStore(path)
	store.Add(Entry{Name: "Test", Tags: "foo bar", URL: "https://example.com"})

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := "| Name | Tags | URL |\n|------|------|-----|\n| Test | foo bar | https://example.com |\n"
	if string(content) != expected {
		t.Errorf("file content mismatch:\ngot:\n%s\nwant:\n%s", string(content), expected)
	}
}
