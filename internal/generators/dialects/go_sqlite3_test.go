package dialects

import (
	"testing"
)

func TestGoSQLite3FlavorConfig(t *testing.T) {
	config := NewGoSQLite3FlavorConfig()

	// Test name
	if config.Name() != "go-sqlite3" {
		t.Errorf("Expected name 'go-sqlite3', got '%s'", config.Name())
	}

	// Test that go-sqlite3 supports features that Turso doesn't
	advancedFeatures := []string{
		"exists_subquery",
		"in_subquery",
		"modulo_operator",
		"regexp",
		"match",
		"filter_clause",
		"window_functions",
		"raise_function",
		"format_function",
		"cte_recursive",
		"cte_materialized",
		"schema_qualified",
		"collate_custom",
	}

	for _, feature := range advancedFeatures {
		if !config.SupportsFeature(feature) {
			t.Errorf("go-sqlite3 should support feature '%s' but doesn't", feature)
		}
	}

	// Test validation (should return nil for now)
	if err := config.ValidateSQL("SELECT 1"); err != nil {
		t.Errorf("ValidateSQL should return nil, got %v", err)
	}
}

func TestGoSQLite3VSTypeComparison(t *testing.T) {
	goSQLite3 := NewGoSQLite3FlavorConfig()
	turso := NewTursoFlavorConfig()

	// Features that go-sqlite3 supports but Turso doesn't
	differingFeatures := []string{
		"exists_subquery",
		"in_subquery",
		"window_functions",
		"cte_recursive",
		"filter_clause",
		"regexp",
		"match",
	}

	for _, feature := range differingFeatures {
		goSQLite3Supports := goSQLite3.SupportsFeature(feature)
		tursoSupports := turso.SupportsFeature(feature)

		if !goSQLite3Supports {
			t.Errorf("go-sqlite3 should support '%s'", feature)
		}

		if tursoSupports {
			t.Errorf("Turso should NOT support '%s' (for testing comparison)", feature)
		}
	}
}
