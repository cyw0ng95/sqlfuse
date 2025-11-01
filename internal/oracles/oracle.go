package oracles

import (
	"database/sql"
)

// Oracle represents a test oracle that validates database behavior
type Oracle interface {
	// Check performs the oracle check and returns an error if a bug is detected
	Check() error
	
	// Name returns the name of this oracle
	Name() string
}

// Result represents query execution result for comparison
type Result struct {
	Rows    [][]interface{}
	RowCount int
	Error   error
}

// ExecuteQuery executes a query and returns the result
func ExecuteQuery(db *sql.DB, query string) (*Result, error) {
	rows, err := db.Query(query)
	if err != nil {
		return &Result{Error: err}, err
	}
	defer rows.Close()

	var results [][]interface{}
	cols, err := rows.Columns()
	if err != nil {
		return &Result{Error: err}, err
	}

	for rows.Next() {
		// Create a slice of interface{} to represent each row
		row := make([]interface{}, len(cols))
		rowPointers := make([]interface{}, len(cols))
		for i := range row {
			rowPointers[i] = &row[i]
		}

		if err := rows.Scan(rowPointers...); err != nil {
			return &Result{Error: err}, err
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return &Result{Error: err}, err
	}

	return &Result{
		Rows:     results,
		RowCount: len(results),
		Error:    nil,
	}, nil
}

// CompareResults compares two result sets for equality
func CompareResults(r1, r2 *Result) bool {
	if r1.RowCount != r2.RowCount {
		return false
	}

	for i := range r1.Rows {
		if len(r1.Rows[i]) != len(r2.Rows[i]) {
			return false
		}
		for j := range r1.Rows[i] {
			// Handle NULL values
			if r1.Rows[i][j] == nil && r2.Rows[i][j] == nil {
				continue
			}
			if r1.Rows[i][j] == nil || r2.Rows[i][j] == nil {
				return false
			}
			// Compare values (this is simplified, might need type-specific comparison)
			if r1.Rows[i][j] != r2.Rows[i][j] {
				return false
			}
		}
	}

	return true
}
