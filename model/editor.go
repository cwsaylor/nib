package model

import (
	"nib/commands"
	"nib/keys"
	"nib/messages"
	"nib/theme"
	"nib/types"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditorModel struct {
	textarea textarea.Model
	note     types.Note
	width    int
	height   int
}

func NewEditorModel() EditorModel {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	return EditorModel{textarea: ta}
}

func (e *EditorModel) SetSize(w, h int) {
	e.width = w
	e.height = h
	e.textarea.SetWidth(w - 4)
	e.textarea.SetHeight(h - 6)
}

func (e *EditorModel) SetNote(n types.Note) {
	e.note = n
	content := n.Title
	if n.Body != "" {
		content += "\n" + n.Body
	}
	e.textarea.SetValue(content)
}

func (e *EditorModel) Focus() tea.Cmd {
	return e.textarea.Focus()
}

func splitTitleBody(content string) (string, string) {
	parts := strings.SplitN(content, "\n", 2)
	title := strings.TrimSpace(parts[0])
	if title == "" {
		title = "Untitled"
	}
	if len(title) > 80 {
		title = title[:80] + "..."
	}
	body := ""
	if len(parts) > 1 {
		body = parts[1]
	}
	return title, body
}

func (m Model) updateEdit(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.EditorKeys.Save):
			return m.saveAndReturn()

		case key.Matches(msg, keys.EditorKeys.Back):
			return m.saveAndReturn()

		case key.Matches(msg, keys.EditorKeys.Discard):
			// If it's a new empty note, hard delete it
			if m.editor.note.Title == "" && m.editor.note.Body == "" {
				m.mode = ModeList
				m.inputMode = InputNavigation
				m.editor.textarea.Blur()
				return m, tea.Batch(
					commands.PurgeNote(m.db, m.editor.note.ID),
					commands.LoadNotes(m.db),
				)
			}
			m.mode = ModeList
			m.inputMode = InputNavigation
			m.editor.textarea.Blur()
			return m, commands.LoadNotes(m.db)
		}

	case messages.NoteSavedMsg:
		m.mode = ModeList
		m.inputMode = InputNavigation
		m.editor.textarea.Blur()
		return m, tea.Batch(
			commands.LoadNotes(m.db),
			commands.StatusFlash("Saved!"),
		)
	}

	var cmd tea.Cmd
	m.editor.textarea, cmd = m.editor.textarea.Update(msg)
	return m, cmd
}

func (m Model) saveAndReturn() (Model, tea.Cmd) {
	content := m.editor.textarea.Value()
	title, body := splitTitleBody(content)
	m.editor.note.Title = title
	m.editor.note.Body = body
	return m, commands.SaveNote(m.db, m.editor.note)
}

func (e EditorModel) View() string {
	titleBar := theme.TitleBar.Width(e.width - 4).Render("Editing")
	cmdBar := theme.HelpStyle.Render("  ctrl+s save  esc save & back  ctrl+d discard")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		e.textarea.View(),
		cmdBar,
	)
}
