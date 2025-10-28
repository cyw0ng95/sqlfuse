package dialects

import "sqlsmith-go/internal/stmts/stmts"

// GoSQLite3FlavorConfig implements FlavorConfig for go-sqlite3 (pure SQLite3).
// go-sqlite3 is the canonical Go SQLite driver that supports full SQLite3 features.
// Unlike Turso LibSQL, it supports all standard SQLite features including:
// - Window functions
// - Recursive CTEs
// - EXISTS and IN subqueries
// - REGEXP, MATCH operators
// - FILTER clauses for aggregates
// - And many more advanced SQLite features
type GoSQLite3FlavorConfig struct{}

// Name returns the SQL flavor name.
func (g *GoSQLite3FlavorConfig) Name() string {
	return "go-sqlite3"
}

// SupportsFeature checks if a specific SQL feature is supported by go-sqlite3.
// go-sqlite3 supports full SQLite3, so most features are enabled.
// This is the opposite of Turso which has several restrictions.
//
// Supported features that Turso doesn't have:
// - "exists_subquery": NOT EXISTS (subquery) expressions ✓
// - "in_subquery": IN (subquery) expressions ✓
// - "modulo_operator": % modulo operator ✓
// - "regexp": REGEXP operator ✓ (requires loadable extension or custom function)
// - "match": MATCH operator ✓ (requires FTS)
// - "filter_clause": aggregate FILTER (WHERE ...) clause ✓
// - "window_functions": OVER (...) window functions ✓
// - "raise_function": RAISE() function ✓
// - "format_function": format() function ✓ (SQLite 3.38.0+)
// - "cte_recursive": RECURSIVE keyword in WITH clause ✓
// - "cte_materialized": MATERIALIZED keyword in WITH clause ✓
// - "schema_qualified": schema.table.column syntax ✓
// - "collate_custom": custom collations ✓
//
// Note: Some features like REGEXP may require extensions or custom functions
// but are syntactically supported by SQLite3.
func (g *GoSQLite3FlavorConfig) SupportsFeature(feature string) bool {
	// go-sqlite3 supports full SQLite3, so we return true for all features.
	// This is the permissive default that allows maximum SQL generation coverage.
	return true
}

// ValidateSQL validates SQL syntax for go-sqlite3.
// Currently not implemented - returns nil.
func (g *GoSQLite3FlavorConfig) ValidateSQL(sql string) error {
	// Could integrate with SQLite3 parser here for validation
	return nil
}

// NewGoSQLite3FlavorConfig creates a new go-sqlite3 flavor configuration.
func NewGoSQLite3FlavorConfig() stmts.FlavorConfig {
	return &GoSQLite3FlavorConfig{}
}
