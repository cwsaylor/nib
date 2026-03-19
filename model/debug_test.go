package model

import (
	"path/filepath"
	"nib/db"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewNoteFullLifecycle tests the complete new note flow exactly as
// Bubble Tea would drive it, including WindowSizeMsg.
func TestNewNoteFullLifecycle(t *testing.T) {
	dir := t.TempDir()
	store, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	m := New(store)

	// 1. Init returns LoadNotes cmd
	initCmd := m.Init()
	if initCmd == nil {
		t.Fatal("Init should return a cmd")
	}
	initMsg := initCmd()
	result, _ := m.Update(initMsg)
	m = result.(Model)

	// 2. WindowSizeMsg
	result, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(Model)

	// 3. Press 'n' — should synchronously create note and focus textarea
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(Model)

	if m.mode != ModeEdit {
		t.Fatalf("expected ModeEdit, got %d", m.mode)
	}
	if !m.editor.textarea.Focused() {
		t.Fatal("textarea should be focused immediately after pressing 'n'")
	}
	if m.editor.note.ID == 0 {
		t.Fatal("note should have a real ID")
	}

	// 4. Execute the focus/blink cmd
	if cmd != nil {
		blinkMsg := cmd()
		if blinkMsg != nil {
			result, _ = m.Update(blinkMsg)
			m = result.(Model)
		}
	}

	// 5. Type characters — should work immediately
	for _, ch := range "Hello World" {
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(Model)
	}

	val := m.editor.textarea.Value()
	if val != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q — typing is broken!", val)
	}
}
