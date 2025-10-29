package dialects

import "sqlsmith-go/internal/stmts/stmts"

// DuckDBFlavorConfig implements FlavorConfig for DuckDB.
// DuckDB is an analytical database with extensive SQL support,
// including many features beyond standard SQLite.
//
// Supported features that SQLite doesn't have:
// - Full window functions with FILTER clause
// - Advanced analytical functions (PERCENTILE_CONT, PERCENT_RANK, etc.)
// - ARRAY and STRUCT data types
// - Advanced join types (ASOF, LATERAL)
// - FULL OUTER JOIN
// - More comprehensive date/time functions
// - SAMPLE clause for SELECT
// - PIVOT and UNPIVOT
// - QUALIFY clause for filtering window function results
//
// Note: DuckDB uses different syntax for some SQLite-specific features:
// - ATTACH DATABASE works differently (uses .db file format)
// - Some PRAGMAs are not supported or have different names
// - Virtual tables are handled differently
type DuckDBFlavorConfig struct{}

// Name returns the SQL flavor name.
func (d *DuckDBFlavorConfig) Name() string {
	return "duckdb"
}

// SupportsFeature checks if a specific SQL feature is supported by DuckDB.
// DuckDB has excellent SQL support, including most modern SQL features.
//
// Supported features:
// - "exists_subquery": NOT EXISTS (subquery) expressions ✓
// - "in_subquery": IN (subquery) expressions ✓
// - "modulo_operator": % modulo operator ✓
// - "regexp": REGEXP operator ✓
// - "filter_clause": aggregate FILTER (WHERE ...) clause ✓
// - "window_functions": OVER (...) window functions ✓
// - "cte_recursive": RECURSIVE keyword in WITH clause ✓
// - "cte_materialized": MATERIALIZED keyword in WITH clause ✓
//
// Unsupported or different features:
// - "match": MATCH operator (FTS syntax differs from SQLite)
// - "raise_function": RAISE() function (SQLite-specific)
// - Some SQLite-specific PRAGMAs
func (d *DuckDBFlavorConfig) SupportsFeature(feature string) bool {
	unsupported := map[string]bool{
		"match":          false, // FTS syntax is different in DuckDB
		"raise_function": false, // SQLite-specific, not in DuckDB
	}

	// If explicitly marked as unsupported, return false
	if supported, exists := unsupported[feature]; exists {
		return supported
	}

	// DuckDB supports most modern SQL features
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
