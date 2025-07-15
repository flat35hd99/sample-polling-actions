package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/flat35hd99/sample-polling-actions/domain"
	"github.com/flat35hd99/sample-polling-actions/repository"
)

// InputJSON represents the structure of the JSON input from stdin.
type InputJSON struct {
	Items []Item `json:"items"`
}

// Item represents a single item in the JSON input.
type Item struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// readJSONInput reads JSON from stdin and unmarshals it into an InputJSON struct.
func readJSONInput() (*InputJSON, error) {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	var input InputJSON
	if err := json.Unmarshal(bytes, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &input, nil
}

// mergeData merges new items with existing records, updating existing ones and marking them dirty.
func mergeData(existingRecords map[int]domain.Record, newItems *InputJSON) map[int]domain.Record {
	for _, newItem := range newItems.Items {
		if _, ok := existingRecords[newItem.ID]; ok {
			newRecord := domain.Record{
				ID:      newItem.ID,
				Dirty:   true, // Mark as dirty
				Name:    newItem.Name,
				Content: newItem.Content,
			}
			// Update existing record
			existingRecords[newItem.ID] = newRecord
		} else {
			// Add new record
			newRecord := domain.Record{
				ID:      newItem.ID,
				Dirty:   true, // New records are dirty initially
				Name:    newItem.Name,
				Content: newItem.Content,
			}
			existingRecords[newItem.ID] = newRecord
		}
	}
	return existingRecords
}

func main() {
	input, err := readJSONInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON input: %v\n", err)
		os.Exit(1)
	}

	sqliteDBPath := "data.db"
	repo, err := repository.NewSQLiteRecordRepository(sqliteDBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SQLite repository: %v\n", err)
		os.Exit(1)
	}

	existingRecords, err := repo.GetAllRecords()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading existing SQLite data: %v\n", err)
		os.Exit(1)
	}

	mergedRecords := mergeData(existingRecords, input)

	if err := repo.SaveRecords(mergedRecords); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing SQLite data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Data successfully inserted/updated in data.db")
}
