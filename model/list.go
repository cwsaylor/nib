package model

import (
	"fmt"
	"socnotes/commands"
	"socnotes/db"
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
	hasMore  bool
	loading  bool
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
	if w > 2*h {
		// Landscape: preview takes right half
		l.preview.Width = w/2 - 6
		l.preview.Height = h - 5
	} else {
		// Portrait: preview below list
		previewH := h*50/100 - 2
		if previewH < 3 {
			previewH = 3
		}
		l.preview.Width = w - 6
		l.preview.Height = previewH
	}
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
		m.list.hasMore = msg.HasMore
		m.list.loading = false
		m.list.SetNotes(msg.Notes)
		return m, nil

	case messages.MoreNotesLoadedMsg:
		m.list.loading = false
		m.list.hasMore = msg.HasMore
		if len(msg.Notes) > 0 {
			m.list.notes = append(m.list.notes, msg.Notes...)
		}
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
			// Load more when near the end
			if m.list.hasMore && !m.list.loading && m.list.cursor >= len(m.list.notes)-5 {
				m.list.loading = true
				last := m.list.notes[len(m.list.notes)-1]
				return m, commands.LoadMoreNotes(m.db, db.ListCursor{
					UpdatedAt: last.UpdatedAt,
					ID:        last.ID,
				})
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
	landscape := width > 2*height
	titleBar := theme.TitleBar.Width(width - 4).Render("SocNotes")
	cmdBar := theme.HelpStyle.Render("  n new  e edit  s search  d delete  y yank  t trash  q quit")

	// Compute list dimensions based on orientation
	var listHeight, listWidth int
	if landscape {
		listHeight = height - 4
		listWidth = width / 2
	} else {
		listHeight = height*40/100 - 3
		listWidth = width
	}
	if listHeight < 3 {
		listHeight = 3
	}

	if len(l.notes) == 0 {
		list := theme.SubtleText.Render("  No notes yet. Press 'n' to create one.")
		preview := theme.PreviewBorder.Width(width - 4).Render(l.preview.View())
		return lipgloss.JoinVertical(lipgloss.Left, titleBar, list, preview, cmdBar)
	}

	// Compute visible window, then render only those items
	start := l.cursor - listHeight/2
	if start < 0 {
		start = 0
	}
	end := start + listHeight
	if end > len(l.notes) {
		end = len(l.notes)
		start = end - listHeight
		if start < 0 {
			start = 0
		}
	}

	var listItems []string
	itemWidth := listWidth - 6
	if itemWidth < 10 {
		itemWidth = 10
	}
	for i := start; i < end; i++ {
		n := l.notes[i]
		title := n.Title
		if title == "" {
			title = "Untitled"
		}
		timestamp := relativeTime(n.UpdatedAt)
		line := fmt.Sprintf("%s — %s", title, timestamp)

		if i == l.cursor {
			listItems = append(listItems, theme.SelectedItem.Width(itemWidth).Render("> "+line))
		} else {
			listItems = append(listItems, theme.NormalItem.Width(itemWidth).Render("  "+line))
		}
	}

	list := strings.Join(listItems, "\n")

	if landscape {
		// Side-by-side: list on left, preview on right
		previewWidth := width - listWidth - 4
		if previewWidth < 10 {
			previewWidth = 10
		}
		preview := theme.PreviewBorder.Width(previewWidth).Render(l.preview.View())
		listPane := lipgloss.NewStyle().Width(listWidth).Height(listHeight).Render(list)
		middle := lipgloss.JoinHorizontal(lipgloss.Top, listPane, preview)
		return lipgloss.JoinVertical(lipgloss.Left, titleBar, middle, cmdBar)
	}

	// Portrait: vertical stack
	preview := theme.PreviewBorder.Width(width - 4).Render(l.preview.View())
	return lipgloss.JoinVertical(lipgloss.Left, titleBar, list, preview, cmdBar)
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
