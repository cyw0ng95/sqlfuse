package generators_test

import (
	"testing"

	"sqlsmith-go/internal/generators"
	"sqlsmith-go/internal/generators/turso"
)

func TestGeneratorInterface(t *testing.T) {
	// Create a turso generator
	gen := turso.NewGenerator(12345)

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
	gen := turso.NewGenerator(54321)

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
