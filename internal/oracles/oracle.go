package oracles

import (
	"database/sql"
	"strconv"
)

// ComparisonResult represents the outcome of comparing query results.
type ComparisonResult int

const (
	// Pass indicates the query results match as expected.
	Pass ComparisonResult = iota
	// Fail indicates a potential logical bug was found.
	Fail
	// Error indicates an error occurred during execution or comparison.
	Error
)

// String returns a string representation of the ComparisonResult.
func (c ComparisonResult) String() string {
	switch c {
	case Pass:
		return "Pass"
	case Fail:
		return "Fail"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

// QueryResult holds the result of executing a query.
type QueryResult struct {
	SQL    string
	Result string // Result as string (rows concatenated)
	Error  error
}

// Oracle defines the interface for SQL test oracles.
// Oracles generate equivalent queries and validate that they produce consistent results,
// helping to detect logical bugs in database systems.
type Oracle interface {
	// Name returns the name of this oracle (e.g., "NOREC", "TLP").
	Name() string

	// TransformQuery takes a base SELECT query and transforms it into one or more
	// equivalent queries that should produce consistent results.
	// Returns empty slice if the query is not suitable for this oracle.
	TransformQuery(baseQuery string) []string

	// CompareResults compares the results from equivalent queries.
	// Returns Pass if results are as expected, Fail if a bug is detected, Error on execution errors.
	CompareResults(results []QueryResult) ComparisonResult

	// IsApplicable checks if this oracle can be applied to the given query.
	IsApplicable(query string) bool
}

// BaseOracle provides common functionality for oracle implementations.
type BaseOracle struct {
	name string
}

// NewBaseOracle creates a new base oracle with the given name.
func NewBaseOracle(name string) *BaseOracle {
	return &BaseOracle{name: name}
}

// Name returns the oracle's name.
func (b *BaseOracle) Name() string {
	return b.name
}

// ExecuteQuery executes a query and returns the result as a string.
func ExecuteQuery(db *sql.DB, query string) QueryResult {
	result := QueryResult{SQL: query}

	rows, err := db.Query(query)
	if err != nil {
		result.Error = err
		return result
	}
	defer rows.Close()

	var resultStr string
	cols, err := rows.Columns()
	if err != nil {
		result.Error = err
		return result
	}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			result.Error = err
			return result
		}

		for i, val := range values {
			if i > 0 {
				resultStr += "|"
			}
			if val == nil {
				resultStr += "NULL"
			} else {
				// Handle different types returned by SQLite
				switch v := val.(type) {
				case []byte:
					resultStr += string(v)
				case string:
					resultStr += v
				case int64:
					// Convert int64 to string manually
					num := v
					if num == 0 {
						resultStr += "0"
					} else {
						if num < 0 {
							resultStr += "-"
							num = -num
						}
						digits := ""
						for num > 0 {
							digits = string(rune('0'+num%10)) + digits
							num /= 10
						}
						resultStr += digits
					}
				case float64:
					// Convert float64 to string
					resultStr += strconv.FormatFloat(v, 'f', -1, 64)
				default:
					resultStr += "?"
				}
			}
		}
		resultStr += "\n"
	}

	if err := rows.Err(); err != nil {
		result.Error = err
		return result
	}

	result.Result = resultStr
	return result
}

// ValidateWithOracle executes a base query and its oracle transformations,
// then compares the results using the oracle's comparison logic.
func ValidateWithOracle(db *sql.DB, oracle Oracle, baseQuery string) ComparisonResult {
	if !oracle.IsApplicable(baseQuery) {
		return Error
	}

	transformedQueries := oracle.TransformQuery(baseQuery)
	if len(transformedQueries) == 0 {
		return Error
	}

	results := make([]QueryResult, 0, len(transformedQueries)+1)

	// Execute base query
	baseResult := ExecuteQuery(db, baseQuery)
	results = append(results, baseResult)

	// Execute transformed queries
	for _, tq := range transformedQueries {
		result := ExecuteQuery(db, tq)
		results = append(results, result)
	}

	return oracle.CompareResults(results)
}
