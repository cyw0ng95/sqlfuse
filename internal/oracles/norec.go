package oracles

import (
	"strconv"
	"strings"
)

// NoRecOracle implements the NOREC (No Empty Result Check) oracle.
// It transforms a SELECT query into a query that counts the results,
// then verifies that the count matches the actual number of rows returned.
// This helps detect bugs where COUNT(*) returns incorrect values.
type NoRecOracle struct {
	*BaseOracle
}

// NewNoRecOracle creates a new NOREC oracle.
func NewNoRecOracle() *NoRecOracle {
	return &NoRecOracle{
		BaseOracle: NewBaseOracle("NOREC"),
	}
}

// IsApplicable checks if the query is a SELECT statement with a FROM clause.
func (n *NoRecOracle) IsApplicable(query string) bool {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	return strings.HasPrefix(queryUpper, "SELECT") &&
		strings.Contains(queryUpper, "FROM")
}

// TransformQuery transforms a SELECT query into a COUNT(*) version.
// For example:
//   SELECT a, b FROM t WHERE x > 5
// becomes:
//   SELECT COUNT(*) FROM (SELECT a, b FROM t WHERE x > 5)
func (n *NoRecOracle) TransformQuery(baseQuery string) []string {
	if !n.IsApplicable(baseQuery) {
		return []string{}
	}

	// Wrap the original query in a COUNT subquery
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ")"

	return []string{countQuery}
}

// CompareResults verifies that COUNT(*) matches the number of rows from the base query.
func (n *NoRecOracle) CompareResults(results []QueryResult) ComparisonResult {
	if len(results) < 2 {
		return Error
	}

	baseResult := results[0]
	countResult := results[1]

	// Check for errors
	if baseResult.Error != nil || countResult.Error != nil {
		return Error
	}

	// Count actual rows in base result
	actualRows := 0
	if baseResult.Result != "" {
		actualRows = strings.Count(baseResult.Result, "\n")
	}

	// Parse COUNT(*) result
	countStr := strings.TrimSpace(countResult.Result)
	var expectedCount int
	if countStr == "" || countStr == "NULL\n" {
		expectedCount = 0
	} else {
		// Extract count value (first column, first row)
		lines := strings.Split(countStr, "\n")
		if len(lines) > 0 && lines[0] != "" {
			countValue := strings.TrimSpace(lines[0])
			var err error
			expectedCount, err = strconv.Atoi(countValue)
			if err != nil {
				// If parsing fails, return Error
				return Error
			}
		}
	}

	// Compare counts
	if actualRows != expectedCount {
		return Fail
	}

	return Pass
}
