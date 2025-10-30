package generators

import (
	"sqlsmith-go/internal/stmts/stmts"
	"testing"
)

// TestDuckDBGenerator tests the DuckDB generator
func TestDuckDBGenerator(t *testing.T) {
	gen := NewDuckDBGenerator(12345)

	// Verify it implements the Generator interface
	var _ Generator = gen

	// Test Name method
	if name := gen.Name(); name != "duckdb" {
		t.Errorf("Expected generator name 'duckdb', got '%s'", name)
	}

	// Test SupportedStmts method
	supportedStmts := gen.SupportedStmts()
	if len(supportedStmts) == 0 {
		t.Error("Expected non-empty supported statements map")
	}

	// Test GenerateWithDB method
	sql := gen.GenerateWithDB(nil)
	if sql == "" {
		t.Error("Expected non-empty SQL statement")
	}
}

// TestDuckDBGeneratorStatementCoverage tests that DuckDB generator supports all DuckDB-specific statements
func TestDuckDBGeneratorStatementCoverage(t *testing.T) {
	gen := NewDuckDBGenerator(54321)
	supportedStmts := gen.SupportedStmts()

	// DuckDB-specific statements that must be supported
	requiredStmts := []string{
		"copy",
		"set",
		"reset",
		"create_schema",
		"drop_schema",
		"create_sequence",
		"drop_sequence",
		"create_macro",
		"drop_macro",
		"create_type",
		"drop_type",
		"describe",
		"show",
		"summarize",
		"use",
		"call",
		"checkpoint",
		"export_database",
		"import_database",
		"prepare",
		"execute",
	}

	for _, stmt := range requiredStmts {
		weight, exists := supportedStmts[stmt]
		if !exists {
			t.Errorf("DuckDB generator missing support for '%s'", stmt)
		}
		if weight == 0 {
			t.Errorf("DuckDB generator has zero weight for '%s'", stmt)
		}
		t.Logf("Statement '%s' has weight %d", stmt, weight)
	}
}

// TestDuckDBVsOtherGenerators compares DuckDB generator weights with other generators
func TestDuckDBVsOtherGenerators(t *testing.T) {
	duckdbGen := NewDuckDBGenerator(42)
	tursoGen := NewTursoGenerator(42)

	duckdbStmts := duckdbGen.SupportedStmts()
	tursoStmts := tursoGen.SupportedStmts()

	// DuckDB should have higher weights for analytical features
	analyticalStmts := []string{
		"select_window",
		"select_multiple_windows",
		"select_cte",
		"select_multiple_cte",
		"select_recursive_cte",
	}

	for _, stmt := range analyticalStmts {
		duckdbWeight := duckdbStmts[stmt]
		tursoWeight := tursoStmts[stmt]

		if duckdbWeight <= tursoWeight {
			t.Errorf("Expected DuckDB weight for '%s' (%d) to be greater than Turso weight (%d)",
				stmt, duckdbWeight, tursoWeight)
		}
	}

	// DuckDB-specific statements should only exist in DuckDB
	duckdbOnlyStmts := []string{
		"copy",
		"set",
		"reset",
		"create_schema",
		"drop_schema",
		"create_sequence",
		"drop_sequence",
		"create_macro",
		"drop_macro",
		"create_type",
		"drop_type",
		"describe",
		"show",
		"summarize",
		"checkpoint",
	}

	for _, stmt := range duckdbOnlyStmts {
		if _, exists := duckdbStmts[stmt]; !exists {
			t.Errorf("DuckDB generator missing DuckDB-specific statement '%s'", stmt)
		}

		// These might exist in Turso with 0 weight or not exist at all
		tursoWeight := tursoStmts[stmt]
		if tursoWeight > 0 {
			t.Logf("Note: Turso also supports '%s' with weight %d", stmt, tursoWeight)
		}
	}
}

// TestDuckDBStatementGeneration tests that DuckDB generator includes all statement types
func TestDuckDBStatementGeneration(t *testing.T) {
	gen := NewDuckDBGenerator(99999)

	// List of DuckDB-specific statements
	stmtsToCheck := []string{
		"copy",
		"set",
		"reset",
		"describe",
		"show",
		"summarize",
		"create_schema",
		"create_sequence",
		"create_macro",
		"create_type",
		"checkpoint",
		"prepare",
		"execute",
	}

	supportedStmts := gen.SupportedStmts()

	for _, stmtName := range stmtsToCheck {
		if _, exists := supportedStmts[stmtName]; !exists {
			t.Errorf("Statement '%s' not in supported statements", stmtName)
		} else {
			t.Logf("Statement '%s' is supported", stmtName)
		}
	}
}

// TestDuckDBGeneratorWeightDistribution tests the weight distribution
func TestDuckDBGeneratorWeightDistribution(t *testing.T) {
	weights := DefaultDuckDBStmtWeights()

	// High-priority DuckDB statements should have reasonable weights
	highPriorityStmts := map[string]uint64{
		"copy":      60, // Essential for data import/export
		"describe":  45, // Very useful metadata
		"show":      45, // Very useful metadata
		"summarize": 40, // DuckDB's data profiling feature
		"set":       40, // DuckDB's config mechanism
	}

	for stmt, expectedMin := range highPriorityStmts {
		stmtType := stmts.StmtType(stmt)
		actualWeight := weights[stmtType]
		if actualWeight < expectedMin {
			t.Errorf("Statement '%s' has weight %d, expected at least %d", stmt, actualWeight, expectedMin)
		}
	}

	// Total weight should be reasonable
	var totalWeight uint64
	for _, w := range weights {
		totalWeight += w
	}

	if totalWeight == 0 {
		t.Error("Total weight should not be zero")
	}

	t.Logf("Total weight: %d", totalWeight)
	t.Logf("Number of statement types: %d", len(weights))
}
