package repository

import "github.com/flat35hd99/sample-polling-actions/domain"

// RecordRepository defines the interface for data operations on Records.
type RecordRepository interface {
	GetAllRecords() (map[int]domain.Record, error)
	SaveRecords(records map[int]domain.Record) error
}
