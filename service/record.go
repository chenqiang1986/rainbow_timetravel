package service

import (
	"context"
	"errors"
	"time"

	"github.com/chenqiang1986/rainbow_timetravel/entity"
)

var (
	ErrRecordDoesNotExist  = errors.New("record with that id does not exist")
	ErrRecordIDInvalid     = errors.New("record id must > 0")
	ErrRecordAlreadyExists = errors.New("record already exists")
)

// Implements method to get, create, and update record data, plus
// time-travel queries over a record's full version history.
type RecordService interface {

	// GetRecord will retrieve an record.
	GetRecord(ctx context.Context, id int) (entity.Record, error)

	// CreateRecord will insert a new record.
	//
	// If it a record with that id already exists it will fail.
	CreateRecord(ctx context.Context, record entity.Record) error

	// UpdateRecord will change the internal `Map` values of the record if they exist.
	// if the update[key] is null it will delete that key from the record's Map.
	//
	// UpdateRecord will error if id <= 0 or the record does not exist with that id.
	UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error)

	// ListVersions returns every stored version of the record, ordered by
	// version ascending. Errors with ErrRecordDoesNotExist if no version
	// has been written for this id.
	ListVersions(ctx context.Context, id int) ([]entity.RecordVersion, error)

	// GetRecordAtVersion returns the record state at the exact version.
	GetRecordAtVersion(ctx context.Context, id int, version int) (entity.RecordSnapshot, error)

	// GetRecordAtTimestamp returns the latest record version whose
	// created_on is <= the given timestamp.
	GetRecordAtTimestamp(ctx context.Context, id int, ts time.Time) (entity.RecordSnapshot, error)
}
