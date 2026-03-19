package db

import (
	"database/sql"
	"fmt"
	"socnotes/types"
	"strings"
	"time"
)

const PageSize = 50

// ListCursor holds the keyset pagination cursor for ListActive.
type ListCursor struct {
	UpdatedAt time.Time
	ID        int
}

func (d *DB) ListActive(limit int, before *ListCursor) ([]types.Note, error) {
	var rows *sql.Rows
	var err error

	if before != nil {
		rows, err = d.conn.Query(`
			SELECT id, title, substr(body, 1, 120), created_at, updated_at
			FROM notes WHERE deleted_at IS NULL
			  AND (updated_at < ? OR (updated_at = ? AND id < ?))
			ORDER BY updated_at DESC, id DESC
			LIMIT ?`, before.UpdatedAt.Format("2006-01-02 15:04:05"),
			before.UpdatedAt.Format("2006-01-02 15:04:05"), before.ID, limit)
	} else {
		rows, err = d.conn.Query(`
			SELECT id, title, substr(body, 1, 120), created_at, updated_at
			FROM notes WHERE deleted_at IS NULL
			ORDER BY updated_at DESC, id DESC
			LIMIT ?`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []types.Note
	for rows.Next() {
		var n types.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (d *DB) GetNote(id int) (types.Note, error) {
	var n types.Note
	err := d.conn.QueryRow(`
		SELECT id, title, body, created_at, updated_at, deleted_at
		FROM notes WHERE id = ?`, id).
		Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt)
	return n, err
}

func (d *DB) Create(title, body string) (types.Note, error) {
	result, err := d.conn.Exec(`
		INSERT INTO notes (title, body) VALUES (?, ?)`, title, body)
	if err != nil {
		return types.Note{}, err
	}
	id, _ := result.LastInsertId()

	// Sync FTS index
	d.conn.Exec(`INSERT INTO notes_fts(rowid, title, body) VALUES (?, ?, ?)`, id, title, body)

	return d.GetNote(int(id))
}

func (d *DB) Update(n types.Note) (types.Note, error) {
	// Get old values for FTS delete
	old, err := d.GetNote(n.ID)
	if err != nil {
		return types.Note{}, err
	}

	_, err = d.conn.Exec(`
		UPDATE notes SET title = ?, body = ?, updated_at = datetime('now')
		WHERE id = ?`, n.Title, n.Body, n.ID)
	if err != nil {
		return types.Note{}, err
	}

	// Sync FTS: remove old, add new
	d.conn.Exec(`DELETE FROM notes_fts WHERE rowid = ?`, old.ID)
	d.conn.Exec(`INSERT INTO notes_fts(rowid, title, body) VALUES (?, ?, ?)`, n.ID, n.Title, n.Body)

	return d.GetNote(n.ID)
}

func (d *DB) Upsert(n types.Note) (types.Note, error) {
	if n.ID == 0 {
		return d.Create(n.Title, n.Body)
	}
	return d.Update(n)
}

func (d *DB) SoftDelete(id int) error {
	_, err := d.conn.Exec(`
		UPDATE notes SET deleted_at = datetime('now'), updated_at = datetime('now')
		WHERE id = ?`, id)
	if err != nil {
		return err
	}
	// Remove from FTS index
	d.conn.Exec(`DELETE FROM notes_fts WHERE rowid = ?`, id)
	return nil
}

func (d *DB) Restore(id int) error {
	_, err := d.conn.Exec(`
		UPDATE notes SET deleted_at = NULL, updated_at = datetime('now')
		WHERE id = ?`, id)
	if err != nil {
		return err
	}
	// Re-add to FTS index
	n, err := d.GetNote(id)
	if err != nil {
		return err
	}
	d.conn.Exec(`INSERT INTO notes_fts(rowid, title, body) VALUES (?, ?, ?)`, n.ID, n.Title, n.Body)
	return nil
}

func (d *DB) HardDelete(id int) error {
	// Remove from FTS first
	d.conn.Exec(`DELETE FROM notes_fts WHERE rowid = ?`, id)
	_, err := d.conn.Exec(`DELETE FROM notes WHERE id = ?`, id)
	return err
}

func (d *DB) PurgeExpiredTrash() (int64, error) {
	// Get IDs to purge for FTS cleanup
	rows, err := d.conn.Query(`
		SELECT id FROM notes WHERE deleted_at IS NOT NULL
		AND deleted_at < datetime('now', '-30 days')`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return 0, nil
	}

	// Remove from FTS and notes
	for _, id := range ids {
		d.conn.Exec(`DELETE FROM notes_fts WHERE rowid = ?`, id)
	}
	result, err := d.conn.Exec(`
		DELETE FROM notes WHERE deleted_at IS NOT NULL
		AND deleted_at < datetime('now', '-30 days')`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (d *DB) ListTrashed() ([]types.Note, error) {
	rows, err := d.conn.Query(`
		SELECT id, title, body, created_at, updated_at, deleted_at
		FROM notes WHERE deleted_at IS NOT NULL
		ORDER BY deleted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []types.Note
	for rows.Next() {
		var n types.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (d *DB) Search(query string, limit int, offset int) ([]types.Note, error) {
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	ftsQuery := sanitizeFTSQuery(query)

	rows, err := d.conn.Query(`
		SELECT n.id, n.title,
			snippet(notes_fts, 1, '>>', '<<', '...', 32) AS preview,
			n.updated_at
		FROM notes_fts fts
		JOIN notes n ON n.id = fts.rowid
		WHERE notes_fts MATCH ?
		  AND n.deleted_at IS NULL
		ORDER BY bm25(notes_fts, 10.0, 1.0)
		LIMIT ? OFFSET ?`, ftsQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []types.Note
	for rows.Next() {
		var n types.Note
		var updatedAt time.Time
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &updatedAt); err != nil {
			return nil, err
		}
		n.UpdatedAt = updatedAt
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func sanitizeFTSQuery(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return q
	}
	if strings.HasPrefix(q, "#") {
		return fmt.Sprintf(`"%s"`, q)
	}
	words := strings.Fields(q)
	for i, w := range words {
		if !strings.HasSuffix(w, "*") && !strings.HasPrefix(w, `"`) {
			words[i] = w + "*"
		}
	}
	return strings.Join(words, " ")
}
