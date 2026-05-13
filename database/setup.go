package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Open returns a *sql.DB for the SQLite file at dbPath. It assumes the schema
// already exists. The caller owns the returned handle.
func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %q: %w", dbPath, err)
	}
	return db, nil
}

// Setup opens (or creates) the SQLite database at dbPath and applies the DDL
// from schemaPath. The returned *sql.DB is ready to use; the caller owns it
// and is responsible for closing it.
func Setup(dbPath, schemaPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db %q: %w", dbPath, err)
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read schema %q: %w", schemaPath, err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return db, nil
}
