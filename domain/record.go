package domain

// Record represents a single row in the CSV file, unifying the data structure
// used across different parts of the application.
type Record struct {
	ID      int
	Dirty   bool
	Name    string
	Content string
}
