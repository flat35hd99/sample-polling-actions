package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

// CSVRecord represents a single record in the data.csv file.
type CSVRecord struct {
	ID      int
	Dirty   bool
	Name    string
	Content string
}

// 以下のような json が標準入力で与えられるので、...
// {
//   "items": [
//     {
//       "id": 1,
//       "name": "sample-polling-actions",
//       "content": "This is a sample repository for polling actions."
//     },
//     {
//       "id": 2,
//       "name": "sample.json",
//       "content": "This JSON file contains metadata about the repository and its files."
//     },
//     {
//       "id": 3,
//       "name": "polling.yaml",
//       "content": "This is a GitHub Actions workflow file for polling an API."
//     }
//   ]
// }

// data.csv に追記していく。idが同じだった場合、追記するのではなく既存のレコードを更新してdirtyフラグをたてる。data.csv の内容は以下のようになる。
// id,dirty,name,content
// 1,true,sample-polling-actions,This is a sample repository for polling actions.

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

// readCSVData reads existing data from data.csv into a map for easy lookup.
func readCSVData(filePath string) (map[int]CSVRecord, error) {
	records := make(map[int]CSVRecord)

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
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

		records[id] = CSVRecord{
			ID:      id,
			Dirty:   dirty,
			Name:    line[2],
			Content: line[3],
		}
	}
	return records, nil
}

// mergeData merges new items with existing CSV records, updating existing ones and marking them dirty.
func mergeData(existingRecords map[int]CSVRecord, newItems *InputJSON) map[int]CSVRecord {
	for _, newItem := range newItems.Items {
		if _, ok := existingRecords[newItem.ID]; ok {
			newRecord := CSVRecord{
				ID:      newItem.ID,
				Dirty:   true, // Mark as dirty
				Name:    newItem.Name,
				Content: newItem.Content,
			}
			// Update existing record
			existingRecords[newItem.ID] = newRecord
		} else {
			// Add new record
			newRecord := CSVRecord{
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

// writeCSVData writes the merged data back to data.csv.
func writeCSVData(filePath string, records map[int]CSVRecord) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

func main() {
	input, err := readJSONInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON input: %v\n", err)
		os.Exit(1)
	}

	csvFilePath := "data.csv"
	existingRecords, err := readCSVData(csvFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading existing CSV data: %v\n", err)
		os.Exit(1)
	}

	mergedRecords := mergeData(existingRecords, input)

	if err := writeCSVData(csvFilePath, mergedRecords); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Data successfully inserted/updated in data.csv")
}
