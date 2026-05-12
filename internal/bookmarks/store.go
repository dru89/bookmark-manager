package bookmarks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Entry represents a single bookmark.
type Entry struct {
	Name string
	Tags string
	URL  string
}

// Store manages reading and writing bookmarks to a markdown table file.
type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string {
	return s.path
}

// Load reads all bookmarks from the file.
func (s *Store) Load() ([]Entry, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	headerSeen := false
	separatorSeen := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || !strings.HasPrefix(line, "|") {
			continue
		}

		// Skip header row
		if !headerSeen {
			headerSeen = true
			continue
		}

		// Skip separator row (|---|---|---|)
		if !separatorSeen {
			separatorSeen = true
			continue
		}

		entry, ok := parseLine(line)
		if ok {
			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}

// Add appends a bookmark to the file. Creates the file with headers if needed.
func (s *Store) Add(entry Entry) error {
	entries, err := s.Load()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	entries = append(entries, entry)
	return s.write(entries)
}

// Remove deletes the first bookmark matching the given name (case-insensitive).
func (s *Store) Remove(name string) (bool, error) {
	entries, err := s.Load()
	if err != nil {
		return false, err
	}

	lower := strings.ToLower(name)
	found := false
	var remaining []Entry
	for _, e := range entries {
		if !found && strings.ToLower(e.Name) == lower {
			found = true
			continue
		}
		remaining = append(remaining, e)
	}

	if !found {
		return false, nil
	}

	return true, s.write(remaining)
}

// Update modifies an existing bookmark by name. Returns false if not found.
func (s *Store) Update(name string, updated Entry) (bool, error) {
	entries, err := s.Load()
	if err != nil {
		return false, err
	}

	lower := strings.ToLower(name)
	found := false
	for i, e := range entries {
		if strings.ToLower(e.Name) == lower {
			entries[i] = updated
			found = true
			break
		}
	}

	if !found {
		return false, nil
	}

	return true, s.write(entries)
}

// Search returns entries where the query matches the name or tags (case-insensitive).
func (s *Store) Search(query string) ([]Entry, error) {
	entries, err := s.Load()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return entries, nil
	}

	tokens := strings.Fields(strings.ToLower(query))
	var results []Entry

	for _, e := range entries {
		searchable := strings.ToLower(e.Name + " " + e.Tags)
		match := true
		for _, t := range tokens {
			if !strings.Contains(searchable, t) {
				match = false
				break
			}
		}
		if match {
			results = append(results, e)
		}
	}

	return results, nil
}

func (s *Store) write(entries []Entry) error {
	// Ensure parent directory exists
	dir := filepath.Dir(s.path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	// Write header
	fmt.Fprintln(w, "| Name | Tags | URL |")
	fmt.Fprintln(w, "|------|------|-----|")

	for _, e := range entries {
		// Escape pipes in values
		name := strings.ReplaceAll(e.Name, "|", "\\|")
		tags := strings.ReplaceAll(e.Tags, "|", "\\|")
		url := strings.ReplaceAll(e.URL, "|", "\\|")
		fmt.Fprintf(w, "| %s | %s | %s |\n", name, tags, url)
	}

	return w.Flush()
}

func parseLine(line string) (Entry, bool) {
	// Split on | and trim
	parts := strings.Split(line, "|")
	// Line like "| a | b | c |" splits into ["", " a ", " b ", " c ", ""]
	if len(parts) < 4 {
		return Entry{}, false
	}

	// Handle escaped pipes by rejoining parts that were split mid-value
	// Simple approach: take columns 1, 2, 3 (0 and last are empty)
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		cleaned = append(cleaned, strings.TrimSpace(p))
	}

	// Remove empty first and last elements
	if len(cleaned) > 0 && cleaned[0] == "" {
		cleaned = cleaned[1:]
	}
	if len(cleaned) > 0 && cleaned[len(cleaned)-1] == "" {
		cleaned = cleaned[:len(cleaned)-1]
	}

	if len(cleaned) < 3 {
		return Entry{}, false
	}

	return Entry{
		Name: cleaned[0],
		Tags: cleaned[1],
		URL:  cleaned[2],
	}, true
}
