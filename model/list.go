package model

import (
	"fmt"
	"socnotes/commands"
	"socnotes/keys"
	"socnotes/messages"
	"socnotes/theme"
	"socnotes/types"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ListModel struct {
	notes    []types.Note
	cursor   int
	preview  viewport.Model
	width    int
	height   int
}

func NewListModel() ListModel {
	return ListModel{
		preview: viewport.New(0, 0),
	}
}

func (l *ListModel) SetSize(w, h int) {
	l.width = w
	l.height = h
	previewH := h*50/100 - 2
	if previewH < 3 {
		previewH = 3
	}
	l.preview.Width = w - 6
	l.preview.Height = previewH
}

func (l *ListModel) SetNotes(notes []types.Note) {
	l.notes = notes
	if l.cursor >= len(l.notes) {
		l.cursor = len(l.notes) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
	l.updatePreview()
}

func (l *ListModel) SelectedNote() *types.Note {
	if len(l.notes) == 0 {
		return nil
	}
	return &l.notes[l.cursor]
}

func (l *ListModel) updatePreview() {
	if n := l.SelectedNote(); n != nil {
		content := n.Title + "\n\n" + n.Body
		l.preview.SetContent(content)
		l.preview.GotoTop()
	} else {
		l.preview.SetContent("No notes yet. Press 'n' to create one.")
	}
}

func (m Model) updateList(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.NotesLoadedMsg:
		m.list.SetNotes(msg.Notes)
		return m, nil

	case messages.NoteSavedMsg:
		return m, commands.LoadNotes(m.db)

	case messages.NoteDeletedMsg:
		return m, commands.LoadNotes(m.db)

	case messages.YankDoneMsg:
		return m, commands.StatusFlash("Copied!")

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.ListKeys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.ListKeys.Up):
			if m.list.cursor > 0 {
				m.list.cursor--
				m.list.updatePreview()
			}

		case key.Matches(msg, keys.ListKeys.Down):
			if m.list.cursor < len(m.list.notes)-1 {
				m.list.cursor++
				m.list.updatePreview()
			}

		case key.Matches(msg, keys.ListKeys.New):
			note, err := m.db.Create("", "")
			if err != nil {
				m.err = err
				return m, nil
			}
			m.mode = ModeEdit
			m.inputMode = InputText
			m.editor.SetNote(note)
			return m, m.editor.Focus()

		case key.Matches(msg, keys.ListKeys.Edit):
			if n := m.list.SelectedNote(); n != nil {
				return m, commands.LoadNote(m.db, n.ID)
			}

		case key.Matches(msg, keys.ListKeys.Search):
			m.mode = ModeSearch
			m.inputMode = InputText
			m.search.Focus()
			return m, nil

		case key.Matches(msg, keys.ListKeys.Delete):
			if n := m.list.SelectedNote(); n != nil {
				m.mode = ModeConfirmDelete
				m.confirm = NewConfirmModel(n.ID, n.Title)
			}

		case key.Matches(msg, keys.ListKeys.Yank):
			if n := m.list.SelectedNote(); n != nil {
				text := n.Title + "\n\n" + n.Body
				return m, commands.Yank(text)
			}

		case key.Matches(msg, keys.ListKeys.Trash):
			m.mode = ModeTrash
			m.inputMode = InputNavigation
			return m, commands.LoadTrashed(m.db)
		}

	case messages.EditNoteMsg:
		m.mode = ModeEdit
		m.inputMode = InputText
		m.editor.SetNote(msg.Note)
		return m, m.editor.Focus()

	}

	// Scroll preview with mouse/keys
	var cmd tea.Cmd
	m.list.preview, cmd = m.list.preview.Update(msg)
	return m, cmd
}

func (l ListModel) View(width, height int) string {
	titleBar := theme.TitleBar.Width(width - 4).Render("SocNotes")

	// Note list
	listHeight := height*40/100 - 3
	if listHeight < 3 {
		listHeight = 3
	}
	var listItems []string
	for i, n := range l.notes {
		title := n.Title
		if title == "" {
			title = "Untitled"
		}
		timestamp := relativeTime(n.UpdatedAt)
		line := fmt.Sprintf("%s — %s", title, timestamp)

		if i == l.cursor {
			listItems = append(listItems, theme.SelectedItem.Width(width-6).Render("> "+line))
		} else {
			listItems = append(listItems, theme.NormalItem.Width(width-6).Render("  "+line))
		}
	}

	if len(listItems) == 0 {
		listItems = append(listItems, theme.SubtleText.Render("  No notes yet. Press 'n' to create one."))
	}

	list := strings.Join(listItems, "\n")
	if len(listItems) > listHeight {
		// Simple scroll: show items around cursor
		start := l.cursor - listHeight/2
		if start < 0 {
			start = 0
		}
		end := start + listHeight
		if end > len(listItems) {
			end = len(listItems)
			start = end - listHeight
			if start < 0 {
				start = 0
			}
		}
		list = strings.Join(listItems[start:end], "\n")
	}

	// Preview pane
	preview := theme.PreviewBorder.Width(width - 4).Render(l.preview.View())

	// Command bar
	cmdBar := theme.HelpStyle.Render("  n new  e edit  s search  d delete  y yank  t trash  q quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		list,
		preview,
		cmdBar,
	)
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}
