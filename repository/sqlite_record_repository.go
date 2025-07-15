package repository

import (
	"database/sql"
	"fmt"

	"github.com/flat35hd99/sample-polling-actions/domain"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteRecordRepository implements domain.RecordRepository for SQLite database.
type SQLiteRecordRepository struct {
	DB *sql.DB
}

// NewSQLiteRecordRepository creates a new SQLiteRecordRepository and initializes the database.
func NewSQLiteRecordRepository(dataSourceName string) (domain.RecordRepository, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	repo := &SQLiteRecordRepository{DB: db}

	// Create table if not exists
	err = repo.createTable()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create records table: %w", err)
	}

	return repo, nil
}

// createTable creates the records table if it does not exist.
func (r *SQLiteRecordRepository) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS records (
		id INTEGER PRIMARY KEY,
		dirty BOOLEAN,
		name TEXT,
		content TEXT
	);`
	_, err := r.DB.Exec(query)
	return err
}

// GetAllRecords reads all records from the SQLite database.
func (r *SQLiteRecordRepository) GetAllRecords() (map[int]domain.Record, error) {
	records := make(map[int]domain.Record)
	rows, err := r.DB.Query("SELECT id, dirty, name, content FROM records")
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var record domain.Record
		if err := rows.Scan(&record.ID, &record.Dirty, &record.Name, &record.Content); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		records[record.ID] = record
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return records, nil
}

// SaveRecords writes the given records to the SQLite database, overwriting existing content.
func (r *SQLiteRecordRepository) SaveRecords(records map[int]domain.Record) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	// Clear existing records
	_, err = tx.Exec("DELETE FROM records")
	if err != nil {
		return fmt.Errorf("failed to clear existing records: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO records(id, dirty, name, content) VALUES(?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		_, err := stmt.Exec(record.ID, record.Dirty, record.Name, record.Content)
		if err != nil {
			return fmt.Errorf("failed to insert record %d: %w", record.ID, err)
		}
	}

	return tx.Commit()
}
