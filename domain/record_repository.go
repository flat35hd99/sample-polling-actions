package domain

// RecordRepository defines the interface for data operations on Records.
type RecordRepository interface {
	GetAllRecords() (map[int]Record, error)
	SaveRecords(records map[int]Record) error
}
