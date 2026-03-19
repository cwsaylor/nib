package model

import (
	"fmt"
	"socnotes/commands"
	"socnotes/theme"

	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmModel struct {
	noteID    int
	noteTitle string
}

func NewConfirmModel(id int, title string) ConfirmModel {
	if title == "" {
		title = "Untitled"
	}
	return ConfirmModel{noteID: id, noteTitle: title}
}

func (m Model) updateConfirmDelete(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.mode = ModeList
			return m, tea.Batch(
				commands.DeleteNote(m.db, m.confirm.noteID),
				commands.StatusFlash("Moved to trash"),
			)
		case "n", "esc":
			m.mode = ModeList
			return m, nil
		}
	}
	return m, nil
}

func (c ConfirmModel) View() string {
	title := c.noteTitle
	if len(title) > 30 {
		title = title[:30] + "..."
	}
	content := fmt.Sprintf("Delete \"%s\"?\n\nMoves to trash (30 day purge)\n\n[y] Yes     [n] No", title)
	return theme.ConfirmDialog.Render(content)
}
