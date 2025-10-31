package generators_test

import (
	"testing"

	"sqlfuse/internal/generators"
)

func TestGeneratorInterface(t *testing.T) {
	// Create a turso generator
	gen := generators.NewTursoGenerator(12345)

	// Verify it implements the Generator interface
	var _ generators.Generator = gen

	// Test Name method
	if name := gen.Name(); name != "turso" {
		t.Errorf("Expected generator name 'turso', got '%s'", name)
	}

	// Test SupportedStmts method
	stmts := gen.SupportedStmts()
	if len(stmts) == 0 {
		t.Error("Expected non-empty supported statements map")
	}

	// Test TokensUsed method (should start at 0 or low value)
	initialTokens := gen.TokensUsed()
	if initialTokens < 0 {
		t.Error("Expected non-negative token count")
	}

	// Test GenerateWithDB method
	sql := gen.GenerateWithDB(nil)
	if sql == "" {
		t.Error("Expected non-empty SQL statement")
	}

	// Verify tokens increased after generation
	afterTokens := gen.TokensUsed()
	if afterTokens <= initialTokens {
		t.Error("Expected token count to increase after generation")
	}
}

func TestBaseGeneratorWeights(t *testing.T) {
	gen := generators.NewTursoGenerator(54321)

	// Get initial weights
	weights := gen.GetWeights()
	if len(weights) == 0 {
		t.Error("Expected non-empty weights map")
	}

	// Test GetMaxRecursionDepth
	depth := gen.GetMaxRecursionDepth()
	if depth < 0 {
		t.Error("Expected non-negative recursion depth")
	}

	// Test SetMaxRecursionDepth
	gen.SetMaxRecursionDepth(5)
	if newDepth := gen.GetMaxRecursionDepth(); newDepth != 5 {
		t.Errorf("Expected recursion depth 5, got %d", newDepth)
	}

	// Verify generator still works after configuration changes
	sql := gen.GenerateWithDB(nil)
	if sql == "" {
		t.Error("Expected non-empty SQL statement after configuration change")
	}
}

func TestGoSQLite3Generator(t *testing.T) {
	// Create a go-sqlite3 generator
	gen := generators.NewGoSQLite3Generator(12345)

	// Verify it implements the Generator interface
	var _ generators.Generator = gen

	// Test Name method
	if name := gen.Name(); name != "go-sqlite3" {
		t.Errorf("Expected generator name 'go-sqlite3', got '%s'", name)
	}

	// Test SupportedStmts method
	stmts := gen.SupportedStmts()
	if len(stmts) == 0 {
		t.Error("Expected non-empty supported statements map")
	}

	// Test GenerateWithDB method
	sql := gen.GenerateWithDB(nil)
	if sql == "" {
		t.Error("Expected non-empty SQL statement")
	}
}

func TestGoSQLite3VsTursoWeights(t *testing.T) {
	// Create both generators with the same seed for comparison
	goSQLite3Gen := generators.NewGoSQLite3Generator(42)
	tursoGen := generators.NewTursoGenerator(42)

	goSQLite3Stmts := goSQLite3Gen.SupportedStmts()
	tursoStmts := tursoGen.SupportedStmts()

	// Verify both have statements
	if len(goSQLite3Stmts) == 0 || len(tursoStmts) == 0 {
		t.Fatal("Expected both generators to have non-empty statement maps")
	}

	// Check that go-sqlite3 has higher weights for advanced features
	// that it fully supports but Turso doesn't
	advancedStmts := []string{
		"select_window",           // Window functions
		"select_multiple_windows", // Multiple window functions
		"select_cte",              // CTEs
		"select_multiple_cte",     // Multiple CTEs
		"select_recursive_cte",    // Recursive CTEs
	}

	for _, stmt := range advancedStmts {
		goSQLite3Weight := goSQLite3Stmts[stmt]
		tursoWeight := tursoStmts[stmt]

		if goSQLite3Weight <= tursoWeight {
			t.Errorf("Expected go-sqlite3 weight for '%s' (%d) to be greater than Turso weight (%d)",
				stmt, goSQLite3Weight, tursoWeight)
		}
	}

	// Check that go-sqlite3 has lower weights for extension functions
	// that Turso has built-in but go-sqlite3 doesn't
	extensionStmts := []string{
		"select_uuid",   // UUID functions (Turso built-in)
		"select_vector", // Vector functions (Turso built-in)
	}

	for _, stmt := range extensionStmts {
		goSQLite3Weight := goSQLite3Stmts[stmt]
		tursoWeight := tursoStmts[stmt]

		if goSQLite3Weight >= tursoWeight {
			t.Errorf("Expected go-sqlite3 weight for '%s' (%d) to be less than Turso weight (%d)",
				stmt, goSQLite3Weight, tursoWeight)
		}
	}
}
