package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

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

// applyUpdates applies the updates to the data map. If an update value is nil,
// the key is deleted from the data map.
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
