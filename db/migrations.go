package db

import "database/sql"

var migrations = []string{
	// Migration 1: core schema
	`CREATE TABLE IF NOT EXISTS notes (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		title       TEXT NOT NULL DEFAULT '',
		body        TEXT NOT NULL DEFAULT '',
		created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
		updated_at  DATETIME NOT NULL DEFAULT (datetime('now')),
		deleted_at  DATETIME DEFAULT NULL
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(
		title,
		body,
		tokenize="unicode61 tokenchars '#@'"
	);

	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY
	);

	CREATE INDEX IF NOT EXISTS idx_notes_deleted_at ON notes(deleted_at)
		WHERE deleted_at IS NOT NULL;`,

	// Migration 2: covering index for paginated active notes list
	`CREATE INDEX IF NOT EXISTS idx_notes_active ON notes(updated_at DESC)
		WHERE deleted_at IS NULL;`,
}

func runMigrations(db *sql.DB) error {
	// Ensure schema_version table exists (bootstrapping)
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY)`)
	if err != nil {
		return err
	}

	var current int
	row := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`)
	if err := row.Scan(&current); err != nil {
		return err
	}

	for i := current; i < len(migrations); i++ {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, i+1); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
