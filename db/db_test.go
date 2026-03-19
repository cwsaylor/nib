package db

import (
	"os"
	"path/filepath"
	"socnotes/types"
	"testing"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestCreateAndGet(t *testing.T) {
	d := testDB(t)

	note, err := d.Create("Hello", "World")
	if err != nil {
		t.Fatal(err)
	}
	if note.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if note.Title != "Hello" {
		t.Fatalf("expected title Hello, got %s", note.Title)
	}

	got, err := d.GetNote(note.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Body != "World" {
		t.Fatalf("expected body World, got %s", got.Body)
	}
}

func TestListActive(t *testing.T) {
	d := testDB(t)

	d.Create("Note 1", "Body 1")
	d.Create("Note 2", "Body 2")

	notes, err := d.ListActive()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
}

func TestSoftDeleteAndRestore(t *testing.T) {
	d := testDB(t)

	note, _ := d.Create("Delete me", "Body")

	if err := d.SoftDelete(note.ID); err != nil {
		t.Fatal(err)
	}

	active, _ := d.ListActive()
	if len(active) != 0 {
		t.Fatal("expected 0 active notes after soft delete")
	}

	trashed, _ := d.ListTrashed()
	if len(trashed) != 1 {
		t.Fatal("expected 1 trashed note")
	}

	if err := d.Restore(note.ID); err != nil {
		t.Fatal(err)
	}

	active, _ = d.ListActive()
	if len(active) != 1 {
		t.Fatal("expected 1 active note after restore")
	}
}

func TestSearch(t *testing.T) {
	d := testDB(t)

	d.Create("Meeting notes", "Discussed the roadmap #meeting")
	d.Create("Shopping list", "Buy groceries and supplies")

	results, err := d.Search("roadmap")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Meeting notes" {
		t.Fatalf("expected Meeting notes, got %s", results[0].Title)
	}
}

func TestSearchHashtag(t *testing.T) {
	d := testDB(t)

	d.Create("Tagged note", "This has #projectx tag")
	d.Create("Other note", "No tags here")

	results, err := d.Search("#projectx")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for hashtag, got %d", len(results))
	}
}

func TestUpsert(t *testing.T) {
	d := testDB(t)

	// Insert
	note, err := d.Upsert(newNote(0, "New", "Body"))
	if err != nil {
		t.Fatal(err)
	}
	if note.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	// Update
	note.Title = "Updated"
	updated, err := d.Upsert(note)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("expected Updated, got %s", updated.Title)
	}
}

func TestHardDelete(t *testing.T) {
	d := testDB(t)

	note, _ := d.Create("Delete me", "Body")
	if err := d.HardDelete(note.ID); err != nil {
		t.Fatal(err)
	}

	_, err := d.GetNote(note.ID)
	if err == nil {
		t.Fatal("expected error getting deleted note")
	}
}

func TestPurgeExpiredTrash(t *testing.T) {
	d := testDB(t)

	note, _ := d.Create("Old trash", "Body")
	d.SoftDelete(note.ID)

	// Manually set deleted_at to 31 days ago
	d.conn.Exec(`UPDATE notes SET deleted_at = datetime('now', '-31 days') WHERE id = ?`, note.ID)

	purged, err := d.PurgeExpiredTrash()
	if err != nil {
		t.Fatal(err)
	}
	if purged != 1 {
		t.Fatalf("expected 1 purged, got %d", purged)
	}
}

func TestOpenCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "test.db")
	os.MkdirAll(filepath.Dir(path), 0755)
	d, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	d.Close()
}

// helper to construct a Note for Upsert
func newNote(id int, title, body string) types.Note {
	return types.Note{ID: id, Title: title, Body: body}
}
