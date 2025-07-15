package main

import (
	"fmt"

	"github.com/flat35hd99/sample-polling-actions/domain"
	"github.com/flat35hd99/sample-polling-actions/repository"
)

// processor is a placeholder function that will be implemented by the user.
// It processes a single record that has been marked as dirty.
func processor(record domain.Record) error {
	fmt.Printf("Processing dirty record: %+v\n", record)
	return nil
}

func main() {
	sqliteDBPath := "data.db"
	repo, err := repository.NewSQLiteRecordRepository(sqliteDBPath)
	if err != nil {
		fmt.Printf("Error initializing SQLite repository: %v\n", err)
		return
	}

	records, err := repo.GetAllRecords()
	if err != nil {
		fmt.Printf("Error reading records from SQLite: %v\n", err)
		return
	}

	for id, record := range records {
		// Process only if dirty is true
		if record.Dirty {
			err := processor(record)
			if err != nil {
				fmt.Printf("Error processing record ID %d: %v\n", record.ID, err)
				continue
			}
			// Mark the record as clean after processing
			record.Dirty = false
			records[id] = record // Update the record in the map
			fmt.Printf("Record ID %d processed successfully and marked as clean.\n", record.ID)
		}
	}

	// Save the updated records back to the CSV
	if err := repo.SaveRecords(records); err != nil {
		fmt.Printf("Error saving records to SQLite: %v\n", err)
		return
	}
}
