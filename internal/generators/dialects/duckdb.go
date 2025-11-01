package dialects

import "sqlfuse/internal/stmts/stmts"

// DuckDBFlavorConfig implements FlavorConfig for DuckDB.
// DuckDB is an in-process SQL OLAP database management system that supports
// a wide range of SQL features including advanced analytics capabilities.
// Reference: https://duckdb.org/docs/stable/sql/introduction
type DuckDBFlavorConfig struct{}

// Name returns the SQL flavor name.
func (d *DuckDBFlavorConfig) Name() string {
	return "duckdb"
}

// SupportsFeature checks if a specific SQL feature is supported by DuckDB.
// DuckDB supports most standard SQL features and many advanced features:
// - Window functions: Full support ✓
// - Recursive CTEs: Full support ✓
// - EXISTS and IN subqueries: Full support ✓
// - REGEXP: Full support (REGEXP_MATCHES, etc.) ✓
// - FILTER clauses for aggregates: Full support ✓
// - Schema qualified names: Full support ✓
// - Custom collations: Partial support
//
// Unlike SQLite-based systems, DuckDB:
// - Has full analytical SQL support (WINDOW, PIVOT, etc.)
// - Supports PostgreSQL-style syntax and functions
// - Has native support for many data types (STRUCT, LIST, MAP, etc.)
// - Does NOT support all SQLite pragmas (different configuration system)
//
// Note: DuckDB does not support SQLite PRAGMA statements in the same way.
// Instead, it uses SET statements for configuration.
func (d *DuckDBFlavorConfig) SupportsFeature(feature string) bool {
	unsupported := map[string]bool{
		// DuckDB doesn't support some SQLite-specific features
		"sqlite_pragma":      false, // DuckDB uses SET instead of PRAGMA
		"match":              false, // FTS MATCH operator is SQLite-specific
		"named_transactions": false, // DuckDB uses standard transactions
	}

	// If explicitly marked as unsupported, return false
	if supported, exists := unsupported[feature]; exists {
		return supported
	}

	// DuckDB supports most SQL features, including many that SQLite doesn't
	// Return true for all other features (permissive default)
	return true
}

// ValidateSQL validates SQL syntax for DuckDB.
// Currently not implemented - returns nil.
func (d *DuckDBFlavorConfig) ValidateSQL(sql string) error {
	// Could integrate with DuckDB parser here for validation
	return nil
}

// NewDuckDBFlavorConfig creates a new DuckDB flavor configuration.
func NewDuckDBFlavorConfig() stmts.FlavorConfig {
	return &DuckDBFlavorConfig{}
}
