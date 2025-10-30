package dialects

import (
	"testing"
)

func TestDuckDBFlavorConfig_Name(t *testing.T) {
	config := NewDuckDBFlavorConfig()
	expected := "duckdb"
	if got := config.Name(); got != expected {
		t.Errorf("Name() = %v, want %v", got, expected)
	}
}

func TestDuckDBFlavorConfig_SupportsFeature(t *testing.T) {
	config := NewDuckDBFlavorConfig()

	// Features that DuckDB supports (unlike some SQLite variants)
	supportedFeatures := []string{
		"exists_subquery",
		"in_subquery",
		"modulo_operator",
		"regexp",
		"filter_clause",
		"window_functions",
		"raise_function",
		"format_function",
		"cte_recursive",
		"cte_materialized",
		"schema_qualified",
		"collate_custom",
	}

	for _, feature := range supportedFeatures {
		if !config.SupportsFeature(feature) {
			t.Errorf("SupportsFeature(%q) = false, want true", feature)
		}
	}

	// Features that DuckDB does NOT support (SQLite-specific)
	unsupportedFeatures := []string{
		"sqlite_pragma",
		"match", // SQLite FTS-specific
		"named_transactions",
	}

	for _, feature := range unsupportedFeatures {
		if config.SupportsFeature(feature) {
			t.Errorf("SupportsFeature(%q) = true, want false", feature)
		}
	}
}

func TestDuckDBFlavorConfig_ValidateSQL(t *testing.T) {
	config := NewDuckDBFlavorConfig()

	// Currently validation is not implemented, should return nil
	if err := config.ValidateSQL("SELECT 1;"); err != nil {
		t.Errorf("ValidateSQL() returned error: %v", err)
	}

	// Even invalid SQL should return nil (not implemented)
	if err := config.ValidateSQL("INVALID SQL"); err != nil {
		t.Errorf("ValidateSQL() returned error for invalid SQL: %v", err)
	}
}
