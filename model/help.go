package model

import (
	"nib/theme"

	"github.com/charmbracelet/lipgloss"
)

var helpText = `nib — Keyboard Shortcuts

  Navigation
    j / ↑      Move up
    k / ↓      Move down
    q          Quit

  Notes
    n          New note
    e / Enter  Edit selected note
    d          Delete (move to trash)
    y          Yank (copy to clipboard)

  Search
    s / /      Open search
    ↑ / ↓      Navigate results
    Enter      Open result
    Esc        Close search

  Editor
    Ctrl+S     Save
    Esc        Save and go back
    Ctrl+D     Discard changes

  Trash
    t          Open trash
    r          Restore note
    x          Permanently delete
    Esc        Go back

  Press ? to close this help`

func helpView(width, height int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8AABF4")).
		Padding(1, 3).
		Width(50)

	dialog := style.Render(theme.SubtleText.Render(helpText))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
