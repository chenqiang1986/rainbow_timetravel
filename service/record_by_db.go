package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	switch _, _, err := loadLatestRecord(ctx, tx, record.ID); {
	case err == nil:
		return ErrRecordAlreadyExists
	case !errors.Is(err, ErrRecordDoesNotExist):
		return err
	}

	if err := insertRecordVersion(ctx, tx, record.ID, 1, record.Data); err != nil {
		return err
	}

	return tx.Commit()
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

func (s *SQLiteRecordStore) ListVersions(ctx context.Context, id int) ([]entity.RecordVersion, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT version, created_on FROM records WHERE id = ? ORDER BY version ASC`,
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []entity.RecordVersion
	for rows.Next() {
		var v entity.RecordVersion
		if err := rows.Scan(&v.Version, &v.Timestamp); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, ErrRecordDoesNotExist
	}
	return versions, nil
}

func (s *SQLiteRecordStore) GetRecordAtVersion(ctx context.Context, id, version int) (entity.RecordSnapshot, error) {
	var (
		dataStr   string
		createdOn time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT data, created_on FROM records WHERE id = ? AND version = ?`,
		id, version,
	).Scan(&dataStr, &createdOn)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.RecordSnapshot{}, ErrRecordDoesNotExist
	}
	if err != nil {
		return entity.RecordSnapshot{}, err
	}

	data := map[string]string{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return entity.RecordSnapshot{}, err
	}
	return entity.RecordSnapshot{ID: id, Version: version, Timestamp: createdOn, Data: data}, nil
}

func (s *SQLiteRecordStore) GetRecordAtTimestamp(ctx context.Context, id int, ts time.Time) (entity.RecordSnapshot, error) {
	var (
		dataStr   string
		version   int
		createdOn time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT data, version, created_on FROM records
		 WHERE id = ? AND created_on <= ?
		 ORDER BY version DESC LIMIT 1`,
		id, ts,
	).Scan(&dataStr, &version, &createdOn)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.RecordSnapshot{}, ErrRecordDoesNotExist
	}
	if err != nil {
		return entity.RecordSnapshot{}, err
	}

	data := map[string]string{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return entity.RecordSnapshot{}, err
	}
	return entity.RecordSnapshot{ID: id, Version: version, Timestamp: createdOn, Data: data}, nil
}
