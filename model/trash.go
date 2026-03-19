package model

import (
	"fmt"
	"nib/commands"
	"nib/keys"
	"nib/messages"
	"nib/theme"
	"nib/types"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TrashModel struct {
	notes  []types.Note
	cursor int
	width  int
	height int
}

func NewTrashModel() TrashModel {
	return TrashModel{}
}

func (t *TrashModel) SetSize(w, h int) {
	t.width = w
	t.height = h
}

func (m Model) updateTrash(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.TrashedNotesMsg:
		m.trash.notes = msg.Notes
		m.trash.cursor = 0
		return m, nil

	case messages.NoteRestoredMsg:
		return m, tea.Batch(
			commands.LoadTrashed(m.db),
			commands.StatusFlash("Restored!"),
		)

	case messages.NotePurgedMsg:
		return m, tea.Batch(
			commands.LoadTrashed(m.db),
			commands.StatusFlash("Permanently deleted"),
		)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.TrashKeys.Back):
			m.mode = ModeList
			m.inputMode = InputNavigation
			return m, commands.LoadNotes(m.db)

		case key.Matches(msg, keys.TrashKeys.Up):
			if m.trash.cursor > 0 {
				m.trash.cursor--
			}

		case key.Matches(msg, keys.TrashKeys.Down):
			if m.trash.cursor < len(m.trash.notes)-1 {
				m.trash.cursor++
			}

		case key.Matches(msg, keys.TrashKeys.Restore):
			if len(m.trash.notes) > 0 {
				n := m.trash.notes[m.trash.cursor]
				return m, commands.RestoreNote(m.db, n.ID)
			}

		case key.Matches(msg, keys.TrashKeys.Purge):
			if len(m.trash.notes) > 0 {
				n := m.trash.notes[m.trash.cursor]
				return m, commands.PurgeNote(m.db, n.ID)
			}
		}
	}
	return m, nil
}

func (t TrashModel) View(width, height int) string {
	titleBar := theme.TitleBar.Width(width - 4).Render("Trash")

	var listItems []string
	for i, n := range t.notes {
		title := n.Title
		if title == "" {
			title = "Untitled"
		}
		deletedAt := ""
		if n.DeletedAt != nil {
			deletedAt = fmt.Sprintf("deleted %s", relativeTime(*n.DeletedAt))
		}
		line := fmt.Sprintf("%s — %s", title, deletedAt)
		if i == t.cursor {
			listItems = append(listItems, theme.SelectedItem.Width(width-6).Render("> "+line))
		} else {
			listItems = append(listItems, theme.NormalItem.Width(width-6).Render("  "+line))
		}
	}

	list := strings.Join(listItems, "\n")
	if len(listItems) == 0 {
		list = theme.SubtleText.Render("  Trash is empty")
	}

	cmdBar := theme.HelpStyle.Render("  r restore  x purge  esc back")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		list,
		"",
		cmdBar,
	)
}
