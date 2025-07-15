package repository

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/flat35hd99/sample-polling-actions/domain"
)

// CSVRecordRepository implements RecordRepository for CSV files.
type CSVRecordRepository struct {
	FilePath string
}

// NewCSVRecordRepository creates a new CSVRecordRepository.
func NewCSVRecordRepository(filePath string) *CSVRecordRepository {
	return &CSVRecordRepository{FilePath: filePath}
}

// GetAllRecords reads all records from the CSV file.
func (r *CSVRecordRepository) GetAllRecords() (map[int]domain.Record, error) {
	records := make(map[int]domain.Record)

	file, err := os.OpenFile(r.FilePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	_, err = reader.Read()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		id, err := strconv.Atoi(line[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse ID from CSV: %w", err)
		}
		dirty, err := strconv.ParseBool(line[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse dirty flag from CSV: %w", err)
		}

		records[id] = domain.Record{
			ID:      id,
			Dirty:   dirty,
			Name:    line[2],
			Content: line[3],
		}
	}
	return records, nil
}

// SaveRecords writes the given records to the CSV file, overwriting existing content.
func (r *CSVRecordRepository) SaveRecords(records map[int]domain.Record) error {
	file, err := os.OpenFile(r.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file for writing: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "dirty", "name", "content"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write records
	for _, record := range records {
		row := []string{
			strconv.Itoa(record.ID),
			strconv.FormatBool(record.Dirty),
			record.Name,
			record.Content,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}
	return nil
}
