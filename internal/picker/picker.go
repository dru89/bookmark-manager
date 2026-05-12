package picker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dru89/bookmark-manager/internal/bookmarks"
	"github.com/sahilm/fuzzy"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	tagStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	urlStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

// Pick launches the interactive fuzzy finder and returns the selected entry.
func Pick(entries []bookmarks.Entry, initialQuery string) (*bookmarks.Entry, error) {
	m := newModel(entries, initialQuery)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := result.(model)
	if final.cancelled || final.selected < 0 {
		return nil, nil
	}

	return &entries[final.filtered[final.selected].index], nil
}

type scoredEntry struct {
	index int
	score int
}

type model struct {
	entries   []bookmarks.Entry
	input     textinput.Model
	filtered  []scoredEntry
	selected  int
	cancelled bool
	height    int
}

// nameSource implements fuzzy.Source for matching against entry names only.
type nameSource []bookmarks.Entry

func (s nameSource) String(i int) string {
	return s[i].Name
}

func (s nameSource) Len() int {
	return len(s)
}

func newModel(entries []bookmarks.Entry, initialQuery string) model {
	ti := textinput.New()
	ti.Placeholder = "Search bookmarks..."
	ti.Focus()
	ti.SetValue(initialQuery)

	m := model{
		entries:  entries,
		input:    ti,
		selected: 0,
		height:   20,
	}
	m.updateFilter()
	return m
}

// scoreEntry scores a single entry against the query tokens.
// Returns (score, matched). Score tiers:
//
//	300+ = exact tag match
//	200+ = tag prefix match
//	150+ = name substring match
//	100+ = tag substring match
//	 50+ = URL substring match
//	 1+  = name fuzzy match only
//	 0   = no match
func scoreEntry(entry bookmarks.Entry, queryTokens []string, nameScore int, nameMatched bool) (int, bool) {
	if len(queryTokens) == 0 {
		return 0, true
	}

	tags := strings.Fields(strings.ToLower(entry.Tags))
	nameLower := strings.ToLower(entry.Name)
	urlLower := strings.ToLower(entry.URL)
	totalScore := 0
	allMatched := true

	for _, qt := range queryTokens {
		tokenScore := 0

		// Check tags: exact > prefix > substring
		for _, tag := range tags {
			if tag == qt {
				tokenScore = max(tokenScore, 300)
			} else if strings.HasPrefix(tag, qt) {
				tokenScore = max(tokenScore, 200)
			} else if strings.Contains(tag, qt) {
				tokenScore = max(tokenScore, 100)
			}
		}

		// Check name as substring
		if tokenScore == 0 {
			if strings.Contains(nameLower, qt) {
				tokenScore = 150
			}
		}

		// Check URL as substring
		if tokenScore == 0 {
			if strings.Contains(urlLower, qt) {
				tokenScore = 50
			}
		}

		// Fall back to name fuzzy match
		if tokenScore == 0 && nameMatched {
			tokenScore = 1 + nameScore
		}

		if tokenScore == 0 {
			allMatched = false
			break
		}
		totalScore += tokenScore
	}

	return totalScore, allMatched
}

func (m *model) updateFilter() {
	query := m.input.Value()
	if query == "" {
		// Show all entries when no filter
		m.filtered = make([]scoredEntry, len(m.entries))
		for i := range m.entries {
			m.filtered[i] = scoredEntry{index: i, score: 0}
		}
	} else {
		queryTokens := strings.Fields(strings.ToLower(query))

		// Get fuzzy name matches for fallback scoring
		nameMatches := fuzzy.FindFrom(query, nameSource(m.entries))
		nameScores := make(map[int]int)
		for _, nm := range nameMatches {
			nameScores[nm.Index] = nm.Score
		}

		var results []scoredEntry
		for i, entry := range m.entries {
			nameScore := nameScores[i]
			nameMatched := nameScore > 0
			score, matched := scoreEntry(entry, queryTokens, nameScore, nameMatched)
			if matched {
				results = append(results, scoredEntry{index: i, score: score})
			}
		}

		// Sort by score descending
		sort.Slice(results, func(a, b int) bool {
			return results[a].score > results[b].score
		})

		m.filtered = results
	}
	// Reset selection if out of bounds
	if m.selected >= len(m.filtered) {
		m.selected = max(0, len(m.filtered)-1)
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height - 4 // Leave room for input + borders

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				return m, tea.Quit
			}
		case tea.KeyUp, tea.KeyCtrlP:
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case tea.KeyDown, tea.KeyCtrlN:
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.updateFilter()
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(m.input.View())
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d", len(m.filtered), len(m.entries))))
	b.WriteString("\n")

	visible := min(m.height, len(m.filtered))
	// Scroll window to keep selected visible
	start := 0
	if m.selected >= visible {
		start = m.selected - visible + 1
	}

	for i := start; i < start+visible && i < len(m.filtered); i++ {
		entry := m.entries[m.filtered[i].index]

		if i == m.selected {
			b.WriteString(selectedStyle.Render("▸ " + entry.Name))
			if entry.Tags != "" {
				b.WriteString("  ")
				b.WriteString(tagStyle.Render(entry.Tags))
			}
			b.WriteString("\n")
			b.WriteString("    ")
			b.WriteString(urlStyle.Render(truncate(entry.URL, 76)))
		} else {
			b.WriteString("  " + entry.Name)
			if entry.Tags != "" {
				b.WriteString("  ")
				b.WriteString(dimStyle.Render(entry.Tags))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
