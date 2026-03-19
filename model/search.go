package model

import (
	"fmt"
	"socnotes/commands"
	"socnotes/messages"
	"socnotes/theme"
	"socnotes/types"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchModel struct {
	input     textinput.Model
	results   []types.Note
	cursor    int
	lastQuery string
	width     int
	height    int
}

func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search notes..."
	ti.CharLimit = 200
	return SearchModel{input: ti}
}

func (s *SearchModel) SetSize(w, h int) {
	s.width = w
	s.height = h
	s.input.Width = w - 8
}

func (s *SearchModel) Focus() {
	s.input.Focus()
	s.input.SetValue("")
	s.results = nil
	s.cursor = 0
	s.lastQuery = ""
}

func (m Model) updateSearch(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "esc":
			m.search.input.Blur()
			m.inputMode = InputNavigation
			m.mode = ModeList
			return m, nil

		case msg.String() == "enter":
			if len(m.search.results) > 0 && m.search.cursor < len(m.search.results) {
				n := m.search.results[m.search.cursor]
				m.search.input.Blur()
				return m, commands.LoadNote(m.db, n.ID)
			}

		case msg.Type == tea.KeyUp:
			if m.search.cursor > 0 {
				m.search.cursor--
			}
			return m, nil

		case msg.Type == tea.KeyDown:
			if m.search.cursor < len(m.search.results)-1 {
				m.search.cursor++
			}
			return m, nil
		}

	case messages.SearchResultsMsg:
		m.search.results = msg.Results
		m.search.cursor = 0
		return m, nil

	case messages.DebounceTickMsg:
		query := m.search.input.Value()
		if query == m.search.lastQuery && query != "" {
			return m, commands.SearchNotes(m.db, query)
		}
		return m, nil

	case messages.EditNoteMsg:
		m.mode = ModeEdit
		m.inputMode = InputText
		m.editor.SetNote(msg.Note)
		return m, m.editor.Focus()
	}

	// Pass to textinput, schedule debounce on change
	oldVal := m.search.input.Value()
	var cmd tea.Cmd
	m.search.input, cmd = m.search.input.Update(msg)
	if m.search.input.Value() != oldVal {
		m.search.lastQuery = m.search.input.Value()
		cmd = tea.Batch(cmd, tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
			return messages.DebounceTickMsg{}
		}))
	}
	return m, cmd
}

func (s SearchModel) View(width, height int) string {
	titleBar := theme.TitleBar.Width(width - 4).Render("Search")

	input := s.input.View()

	var resultItems []string
	for i, n := range s.results {
		title := n.Title
		if title == "" {
			title = "Untitled"
		}
		preview := strings.ReplaceAll(n.Body, "\n", " ")
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		line := fmt.Sprintf("%s\n    %s", title, theme.SubtleText.Render(preview))
		if i == s.cursor {
			resultItems = append(resultItems, theme.SelectedItem.Width(width-6).Render("> "+line))
		} else {
			resultItems = append(resultItems, theme.NormalItem.Width(width-6).Render("  "+line))
		}
	}

	results := strings.Join(resultItems, "\n")
	if len(resultItems) == 0 && s.input.Value() != "" {
		results = theme.SubtleText.Render("  No results")
	}

	cmdBar := theme.HelpStyle.Render("  enter open  ↑↓ navigate  esc back")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		"  "+input,
		"",
		results,
		cmdBar,
	)
}
