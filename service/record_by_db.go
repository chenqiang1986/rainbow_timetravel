package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/chenqiang1986/rainbow_timetravel/entity"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRecordStore persists records in a versioned SQLite table. Each write
// inserts a new row (id, version, data, created_on); reads come from the
// records_latest view, which exposes only the highest version per id.
type SQLiteRecordStore struct {
	db *sql.DB
}

func NewSQLiteRecordStore(db *sql.DB) *SQLiteRecordStore {
	return &SQLiteRecordStore{db: db}
}

func (s *SQLiteRecordStore) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	data, _, err := loadLatestRecord(ctx, s.db, id)
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

	// version 1 is the first version. If any row for this id already exists,
	// the composite PRIMARY KEY (id, version) makes this insert a no-op.
	res, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO records (id, version, data) VALUES (?, 1, ?)`,
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

	data, version, err := loadLatestRecord(ctx, tx, id)
	if err != nil {
		return entity.Record{}, err
	}

	applyUpdates(data, updates)

	if err := insertRecordVersion(ctx, tx, id, version+1, data); err != nil {
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

func loadLatestRecord(ctx context.Context, q rowQuerier, id int) (map[string]string, int, error) {
	var (
		dataStr string
		version int
	)
	err := q.QueryRowContext(ctx,
		`SELECT data, version FROM records_latest WHERE id = ?`, id,
	).Scan(&dataStr, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrRecordDoesNotExist
	}
	if err != nil {
		return nil, 0, err
	}

	data := map[string]string{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, 0, err
	}
	return data, version, nil
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

func insertRecordVersion(ctx context.Context, tx *sql.Tx, id, version int, data map[string]string) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO records (id, version, data) VALUES (?, ?, ?)`,
		id, version, string(dataBytes))
	return err
}
