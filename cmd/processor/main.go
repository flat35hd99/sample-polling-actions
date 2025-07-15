package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Record represents a single row in the CSV file.
type Record struct {
	ID    int
	Name  string
	Value string
	Dirty bool
}

// processor is a placeholder function that will be implemented by the user.
// It processes a single record that has been marked as dirty.
func processor(record Record) error {
	fmt.Printf("Processing dirty record: %+v\n", record)
	return nil
}

func main() {
	// Open the CSV file
	file, err := os.Open("data.csv")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Assuming the first row is a header and skipping it
	_, err = reader.Read()
	if err != nil && err != io.EOF {
		fmt.Printf("Error reading header: %v\n", err)
		return
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			fmt.Printf("Error reading row: %v\n", err)
			continue
		}

		// Parse the record
		id, err := strconv.Atoi(row[0])
		if err != nil {
			fmt.Printf("Error parsing ID '%s': %v\n", row[0], err)
			continue
		}
		value := row[2]
		dirty, err := strconv.ParseBool(row[1])
		if err != nil {
			fmt.Printf("Error parsing Dirty '%s': %v\n", row[3], err)
			continue
		}

		record := Record{
			ID:    id,
			Name:  row[1],
			Value: value,
			Dirty: dirty,
		}

		// Process only if dirty is true
		if record.Dirty {
			err := processor(record)
			if err != nil {
				fmt.Printf("Error processing record ID %d: %v\n", record.ID, err)
				continue
			} else {
				// Mark the record as clean after processing
				record.Dirty = false
				fmt.Printf("Record ID %d processed successfully and marked as clean.\n", record.ID)
			}
		}
	}
}
