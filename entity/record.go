package entity

import "time"

type Record struct {
	ID   int               `json:"id"`
	Data map[string]string `json:"data"`
}

func (d *Record) Copy() Record {
	values := d.Data

	newMap := map[string]string{}
	for key, value := range values {
		newMap[key] = value
	}

	return Record{
		ID:   d.ID,
		Data: newMap,
	}
}

// RecordVersion identifies one stored version of a record.
type RecordVersion struct {
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// RecordSnapshot is a record's state at a specific version/time.
type RecordSnapshot struct {
	ID        int               `json:"id"`
	Version   int               `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Data      map[string]string `json:"data"`
}
