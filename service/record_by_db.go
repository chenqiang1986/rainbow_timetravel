package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/chenqiang1986/rainbow_timetravel/entity"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRecordStore persists records in a SQLite table with columns (id, data JSON).
type SQLiteRecordStore struct {
	db *sql.DB
}

func NewSQLiteRecordStore(db *sql.DB) *SQLiteRecordStore {
	return &SQLiteRecordStore{db: db}
}

func (s *SQLiteRecordStore) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	data, err := loadRecordData(ctx, s.db, id)
	if err != nil {
		return entity.Record{}, err
	}
	return entity.Record{ID: id, Data: data}, nil
}

func (s *SQLiteRecordStore) CreateRecord(ctx context.Context, record entity.Record) error {
	if record.ID <= 0 {
		return ErrRecordIDInvalid
	}

	dataBytes, err := json.Marshal(record.Data)
	if err != nil {
		return err
	}

	// Use INSERT and let SQLite enforce the PRIMARY KEY uniqueness.
	res, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO records (id, data) VALUES (?, ?)`,
		record.ID, string(dataBytes))
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrRecordAlreadyExists
	}
	return nil
}

func (s *SQLiteRecordStore) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.Record{}, err
	}
	defer tx.Rollback()

	data, err := loadRecordData(ctx, tx, id)
	if err != nil {
		return entity.Record{}, err
	}

	applyUpdates(data, updates)

	if err := saveRecordData(ctx, tx, id, data); err != nil {
		return entity.Record{}, err
	}

	if err := tx.Commit(); err != nil {
		return entity.Record{}, err
	}

	return entity.Record{ID: id, Data: data}, nil
}

// rowQuerier is satisfied by both *sql.DB and *sql.Tx.
type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func loadRecordData(ctx context.Context, q rowQuerier, id int) (map[string]string, error) {
	var dataStr string
	err := q.QueryRowContext(ctx, `SELECT data FROM records WHERE id = ?`, id).Scan(&dataStr)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecordDoesNotExist
	}
	if err != nil {
		return nil, err
	}

	data := map[string]string{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func applyUpdates(data map[string]string, updates map[string]*string) {
	for key, value := range updates {
		if value == nil {
			delete(data, key)
		} else {
			data[key] = *value
		}
	}
}

func saveRecordData(ctx context.Context, tx *sql.Tx, id int, data map[string]string) error {
	newBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`UPDATE records SET data = ? WHERE id = ?`,
		string(newBytes), id)
	return err
}
