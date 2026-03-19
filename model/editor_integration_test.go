package model

import (
	"fmt"
	"path/filepath"
	"nib/db"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewNoteCanType(t *testing.T) {
	dir := t.TempDir()
	store, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	m := New(store)
	m.width = 80
	m.height = 40
	m.list.SetSize(80, 40)
	m.editor.SetSize(80, 40)
	m.search.SetSize(80, 40)
	m.trash.SetSize(80, 40)

	// Press 'n' — note is created synchronously, textarea focused immediately
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(Model)

	if m.mode != ModeEdit {
		t.Fatalf("expected ModeEdit, got %d", m.mode)
	}
	if !m.editor.textarea.Focused() {
		t.Fatal("textarea should be focused")
	}

	// Execute the blink cmd
	if cmd != nil {
		blinkMsg := cmd()
		if blinkMsg != nil {
			result, _ = m.Update(blinkMsg)
			m = result.(Model)
		}
	}

	// Type characters
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	m = result.(Model)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = result.(Model)

	val := m.editor.textarea.Value()
	if val != "Hi" {
		t.Fatalf("expected textarea value 'Hi', got %q", val)
	}

	view := m.View()
	if len(view) == 0 {
		t.Fatal("expected non-empty view")
	}
	fmt.Println(view)
}
