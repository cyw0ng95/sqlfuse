package dialects

import (
	"testing"
)

func TestDuckDBFlavorConfig(t *testing.T) {
	cfg := NewDuckDBFlavorConfig()

	if cfg.Name() != "duckdb" {
		t.Errorf("expected Name() = 'duckdb', got %q", cfg.Name())
	}

	// Test features that should be supported
	supportedFeatures := []string{
		"exists_subquery",
		"in_subquery",
		"modulo_operator",
		"regexp",
		"filter_clause",
		"window_functions",
		"cte_recursive",
		"cte_materialized",
		"format_function",
		"collate_custom",
	}

	for _, feature := range supportedFeatures {
		if !cfg.SupportsFeature(feature) {
			t.Errorf("expected SupportsFeature(%q) = true, got false", feature)
		}
	}

	// Test features that should NOT be supported
	unsupportedFeatures := []string{
		"match",          // FTS syntax differs
		"raise_function", // SQLite-specific
	}

	for _, feature := range unsupportedFeatures {
		if cfg.SupportsFeature(feature) {
			t.Errorf("expected SupportsFeature(%q) = false, got true", feature)
		}
	}

	// ValidateSQL should not error (it's a no-op currently)
	if err := cfg.ValidateSQL("SELECT 1"); err != nil {
		t.Errorf("ValidateSQL failed: %v", err)
	}
}

// TestDuckDBVsGoSQLite3FeatureComparison verifies that DuckDB supports
// all the features that go-sqlite3 does (except SQLite-specific ones).
func TestDuckDBVsGoSQLite3FeatureComparison(t *testing.T) {
	duckdb := NewDuckDBFlavorConfig()
	gosqlite3 := NewGoSQLite3FlavorConfig()

	// Features both should support
	commonFeatures := []string{
		"exists_subquery",
		"in_subquery",
		"modulo_operator",
		"regexp",
		"filter_clause",
		"window_functions",
		"cte_recursive",
	}

	for _, feature := range commonFeatures {
		duck := duckdb.SupportsFeature(feature)
		sqlite := gosqlite3.SupportsFeature(feature)
		if duck != sqlite {
			t.Errorf("feature %q: duckdb=%v, go-sqlite3=%v (expected both true)", feature, duck, sqlite)
		}
	}

	// Features DuckDB doesn't support but go-sqlite3 might
	sqliteSpecific := []string{
		"raise_function", // SQLite-specific RAISE()
	}

	for _, feature := range sqliteSpecific {
		if duckdb.SupportsFeature(feature) {
			t.Errorf("DuckDB should not support SQLite-specific feature %q", feature)
		}
	}
}
