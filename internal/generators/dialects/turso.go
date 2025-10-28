package dialects

import "sqlsmith-go/internal/stmts/stmts"

// TursoFlavorConfig implements FlavorConfig for Turso LibSQL.
// It enforces Turso-specific compatibility constraints based on:
// https://github.com/tursodatabase/turso/blob/main/COMPAT.md
type TursoFlavorConfig struct{}

// Name returns the SQL flavor name.
func (t *TursoFlavorConfig) Name() string {
	return "turso"
}

// SupportsFeature checks if a specific SQL feature is supported by Turso LibSQL.
//
// Unsupported features (based on TURSO_COMPAT.md):
// - "exists_subquery": NOT EXISTS (subquery) expressions
// - "in_subquery": IN (subquery) expressions
// - "modulo_operator": % modulo operator
// - "not_less_than": !< operator
// - "not_greater_than": !> operator
// - "regexp": REGEXP operator
// - "match": MATCH operator
// - "filter_clause": aggregate FILTER (WHERE ...) clause
// - "window_functions": OVER (...) window functions
// - "raise_function": RAISE() function
// - "format_function": format() function
// - "cte_recursive": RECURSIVE keyword in WITH clause
// - "cte_materialized": MATERIALIZED keyword in WITH clause
// - "schema_qualified": schema.table.column syntax
// - "named_transactions": named BEGIN/COMMIT transactions
//
// Partially supported features:
// - "collate_custom": custom collations (only BINARY, NOCASE, RTRIM supported)
func (t *TursoFlavorConfig) SupportsFeature(feature string) bool {
	unsupported := map[string]bool{
		"exists_subquery":    false,
		"in_subquery":        false,
		"modulo_operator":    false,
		"not_less_than":      false,
		"not_greater_than":   false,
		"regexp":             false,
		"match":              false,
		"filter_clause":      false,
		"window_functions":   false,
		"raise_function":     false,
		"format_function":    false,
		"cte_recursive":      false,
		"cte_materialized":   false,
		"schema_qualified":   false,
		"named_transactions": false,
		"collate_custom":     false, // Only default collations supported
	}

	// If explicitly marked as unsupported, return false
	if supported, exists := unsupported[feature]; exists {
		return supported
	}

	// Otherwise, assume it's supported (permissive default for SQLite features)
	return true
}

// ValidateSQL validates SQL syntax for Turso LibSQL.
// Currently not implemented - returns nil.
func (t *TursoFlavorConfig) ValidateSQL(sql string) error {
	// Could integrate with ANTLR parser here for validation
	return nil
}

// NewTursoFlavorConfig creates a new Turso flavor configuration.
func NewTursoFlavorConfig() stmts.FlavorConfig {
	return &TursoFlavorConfig{}
}
