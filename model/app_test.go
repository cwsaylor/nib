package model

import (
	"path/filepath"
	"nib/db"
	"nib/messages"
	"nib/types"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func setupTestModel(t *testing.T) Model {
	t.Helper()
	dir := t.TempDir()
	store, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	m := New(store)
	m.width = 80
	m.height = 40
	m.list.SetSize(80, 40)
	m.editor.SetSize(80, 40)
	m.search.SetSize(80, 40)
	m.trash.SetSize(80, 40)
	return m
}

func TestInitialMode(t *testing.T) {
	m := setupTestModel(t)
	if m.mode != ModeList {
		t.Fatalf("expected ModeList, got %d", m.mode)
	}
	if m.inputMode != InputNavigation {
		t.Fatalf("expected InputNavigation, got %d", m.inputMode)
	}
}

func TestNotesLoadedSetsNotes(t *testing.T) {
	m := setupTestModel(t)
	notes := []types.Note{
		{ID: 1, Title: "Note 1", Body: "Body 1", UpdatedAt: time.Now()},
		{ID: 2, Title: "Note 2", Body: "Body 2", UpdatedAt: time.Now()},
	}

	result, _ := m.Update(messages.NotesLoadedMsg{Notes: notes})
	m = result.(Model)

	if len(m.list.notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(m.list.notes))
	}
}

func TestNavigateUpDown(t *testing.T) {
	m := setupTestModel(t)
	m.list.SetNotes([]types.Note{
		{ID: 1, Title: "A", UpdatedAt: time.Now()},
		{ID: 2, Title: "B", UpdatedAt: time.Now()},
		{ID: 3, Title: "C", UpdatedAt: time.Now()},
	})

	// Move down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.list.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", m.list.cursor)
	}

	// Move up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(Model)
	if m.list.cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", m.list.cursor)
	}

	// Don't go below 0
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(Model)
	if m.list.cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", m.list.cursor)
	}
}

func TestNewNoteFocusedImmediately(t *testing.T) {
	m := setupTestModel(t)

	// Press n to create new note — should synchronously create and focus
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(Model)
	if m.mode != ModeEdit {
		t.Fatalf("expected ModeEdit, got %d", m.mode)
	}
	if m.editor.note.ID == 0 {
		t.Fatal("expected note to be created with a real ID")
	}
	if !m.editor.textarea.Focused() {
		t.Fatal("expected textarea to be focused immediately")
	}
	if cmd == nil {
		t.Fatal("expected Focus/blink command to be returned")
	}
}

func TestDeleteConfirmFlow(t *testing.T) {
	m := setupTestModel(t)
	m.list.SetNotes([]types.Note{
		{ID: 1, Title: "Delete me", UpdatedAt: time.Now()},
	})

	// Press d to enter confirm mode
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = result.(Model)
	if m.mode != ModeConfirmDelete {
		t.Fatalf("expected ModeConfirmDelete, got %d", m.mode)
	}
	if m.confirm.noteID != 1 {
		t.Fatalf("expected noteID 1, got %d", m.confirm.noteID)
	}

	// Press n to cancel
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(Model)
	if m.mode != ModeList {
		t.Fatalf("expected ModeList after cancel, got %d", m.mode)
	}
}

func TestDeleteConfirmYes(t *testing.T) {
	m := setupTestModel(t)

	// Create a real note
	note, _ := m.db.Create("Delete me", "Body")
	m.list.SetNotes([]types.Note{note})

	// Enter confirm, then confirm delete
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = result.(Model)
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = result.(Model)
	if m.mode != ModeList {
		t.Fatalf("expected ModeList after confirm, got %d", m.mode)
	}
	if cmd == nil {
		t.Fatal("expected command to be returned")
	}
}

func TestSearchModeTransition(t *testing.T) {
	m := setupTestModel(t)

	// Press s to enter search
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = result.(Model)
	if m.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got %d", m.mode)
	}
	if m.inputMode != InputText {
		t.Fatalf("expected InputText, got %d", m.inputMode)
	}

	// Press esc to go back
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = result.(Model)
	if m.mode != ModeList {
		t.Fatalf("expected ModeList, got %d", m.mode)
	}
}

func TestTrashModeTransition(t *testing.T) {
	m := setupTestModel(t)

	// Press t to enter trash
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = result.(Model)
	if m.mode != ModeTrash {
		t.Fatalf("expected ModeTrash, got %d", m.mode)
	}
}

func TestSplitTitleBody(t *testing.T) {
	tests := []struct {
		input string
		title string
		body  string
	}{
		{"Hello\nWorld", "Hello", "World"},
		{"Hello", "Hello", ""},
		{"", "Untitled", ""},
		{"\nBody only", "Untitled", "Body only"},
	}

	for _, tt := range tests {
		title, body := splitTitleBody(tt.input)
		if title != tt.title {
			t.Errorf("splitTitleBody(%q): title = %q, want %q", tt.input, title, tt.title)
		}
		if body != tt.body {
			t.Errorf("splitTitleBody(%q): body = %q, want %q", tt.input, body, tt.body)
		}
	}
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		dur    time.Duration
		expect string
	}{
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5 min ago"},
		{1 * time.Minute, "1 min ago"},
		{2 * time.Hour, "2 hours ago"},
		{25 * time.Hour, "yesterday"},
		{72 * time.Hour, "3 days ago"},
		{14 * 24 * time.Hour, "2 weeks ago"},
	}

	for _, tt := range tests {
		got := relativeTime(time.Now().Add(-tt.dur))
		if got != tt.expect {
			t.Errorf("relativeTime(-%v) = %q, want %q", tt.dur, got, tt.expect)
		}
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	m := setupTestModel(t)
	m.list.SetNotes([]types.Note{
		{ID: 1, Title: "Test", Body: "Body", UpdatedAt: time.Now()},
	})

	// Should not panic
	v := m.View()
	if v == "" {
		t.Fatal("expected non-empty view")
	}
}
