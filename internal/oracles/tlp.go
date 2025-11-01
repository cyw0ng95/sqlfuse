package oracles

import (
	"regexp"
	"sort"
	"strings"
)

// TLPOracle implements the TLP (Ternary Logic Partitioning) oracle.
// It partitions query results based on a predicate into three groups:
// - WHERE predicate IS TRUE
// - WHERE predicate IS FALSE
// - WHERE predicate IS NULL
// Then verifies that UNION of these partitions equals the original result.
type TLPOracle struct {
	*BaseOracle
}

// NewTLPOracle creates a new TLP oracle.
func NewTLPOracle() *TLPOracle {
	return &TLPOracle{
		BaseOracle: NewBaseOracle("TLP"),
	}
}

// IsApplicable checks if the query is a SELECT with WHERE clause.
func (t *TLPOracle) IsApplicable(query string) bool {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))
	return strings.HasPrefix(queryUpper, "SELECT") &&
		strings.Contains(queryUpper, "FROM") &&
		strings.Contains(queryUpper, "WHERE")
}

// extractWhereCondition extracts the WHERE condition from a query.
func (t *TLPOracle) extractWhereCondition(query string) string {
	// Simple regex to extract WHERE condition
	// This is a simplified version - production code would need better parsing
	re := regexp.MustCompile(`(?i)WHERE\s+(.+?)(?:\s+ORDER\s+|\s+LIMIT\s+|\s+GROUP\s+|$)`)
	matches := re.FindStringSubmatch(query)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// TransformQuery creates three partitioned queries based on ternary logic.
// For a query like: SELECT * FROM t WHERE x > 5
// Generates:
// - SELECT * FROM t WHERE (x > 5) IS TRUE
// - SELECT * FROM t WHERE (x > 5) IS FALSE  
// - SELECT * FROM t WHERE (x > 5) IS NULL
// And wraps them in a UNION to compare with the original.
func (t *TLPOracle) TransformQuery(baseQuery string) []string {
	if !t.IsApplicable(baseQuery) {
		return []string{}
	}

	condition := t.extractWhereCondition(baseQuery)
	if condition == "" {
		return []string{}
	}

	// Replace the WHERE clause with partitioned versions
	queryWithoutWhere := regexp.MustCompile(`(?i)WHERE\s+.+`).ReplaceAllString(baseQuery, "")
	queryWithoutWhere = strings.TrimSpace(queryWithoutWhere)

	// Generate three partitioned queries
	trueQuery := queryWithoutWhere + " WHERE (" + condition + ") IS TRUE"
	falseQuery := queryWithoutWhere + " WHERE (" + condition + ") IS FALSE"
	nullQuery := queryWithoutWhere + " WHERE (" + condition + ") IS NULL"

	// Create UNION query
	unionQuery := "SELECT * FROM (" + trueQuery + " UNION ALL " + falseQuery + " UNION ALL " + nullQuery + ")"

	return []string{unionQuery}
}

// CompareResults verifies that the partitioned UNION equals the original result.
func (t *TLPOracle) CompareResults(results []QueryResult) ComparisonResult {
	if len(results) < 2 {
		return Error
	}

	baseResult := results[0]
	unionResult := results[1]

	// Check for errors
	if baseResult.Error != nil || unionResult.Error != nil {
		return Error
	}

	// Normalize and compare results
	baseRows := normalizeRows(baseResult.Result)
	unionRows := normalizeRows(unionResult.Result)

	if !equalRowSets(baseRows, unionRows) {
		return Fail
	}

	return Pass
}

// normalizeRows splits result into sorted rows for comparison.
func normalizeRows(result string) []string {
	if result == "" {
		return []string{}
	}
	rows := strings.Split(strings.TrimSpace(result), "\n")
	// Sort for consistent comparison
	sort.Strings(rows)
	return rows
}

// equalRowSets compares two row sets for equality.
func equalRowSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
