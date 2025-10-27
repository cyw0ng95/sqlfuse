package turso

import (
	"sqlsmith-go/internal/generators/sqlite/stmts"
	"testing"
)

func TestTursoFlavorConfig(t *testing.T) {
	flavor := NewTursoFlavorConfig()

	// Test flavor name
	if flavor.Name() != "turso" {
		t.Errorf("Expected flavor name 'turso', got '%s'", flavor.Name())
	}

	// Test unsupported features based on TURSO_COMPAT.md
	unsupportedFeatures := []string{
		"exists_subquery",
		"in_subquery",
		"modulo_operator",
		"not_less_than",
		"not_greater_than",
		"regexp",
		"match",
		"filter_clause",
		"window_functions",
		"raise_function",
		"format_function",
		"cte_recursive",
		"cte_materialized",
		"schema_qualified",
		"named_transactions",
		"collate_custom",
	}

	for _, feature := range unsupportedFeatures {
		if flavor.SupportsFeature(feature) {
			t.Errorf("Turso should NOT support '%s' but SupportsFeature returned true", feature)
		}
	}

	// Test that supported features still work
	supportedFeatures := []string{
		"basic_select",
		"joins",
		"aggregates",
		"subqueries", // scalar subqueries are supported
		"unions",
		"order_by",
		"group_by",
		"limit",
		"transactions", // unnamed transactions
	}

	for _, feature := range supportedFeatures {
		if !flavor.SupportsFeature(feature) {
			t.Errorf("Turso should support '%s' but SupportsFeature returned false", feature)
		}
	}
}

func TestTursoFlavorWithGenContext(t *testing.T) {
	flavor := NewTursoFlavorConfig()
	ctx := stmts.NewGenContextWithFlavor(nil, nil, 2, flavor)

	// Verify context has Turso flavor
	if ctx.Flavor.Name() != "turso" {
		t.Errorf("Expected context to have Turso flavor, got '%s'", ctx.Flavor.Name())
	}

	// Verify feature checks work through context
	if ctx.SupportsFeature("window_functions") {
		t.Error("Turso context should not support window functions")
	}

	if ctx.SupportsFeature("cte_recursive") {
		t.Error("Turso context should not support recursive CTEs")
	}

	// Verify supported features
	if !ctx.SupportsFeature("basic_select") {
		t.Error("Turso context should support basic SELECT")
	}
}

func TestTursoGeneratorUsesFlavorConfig(t *testing.T) {
	gen := NewGenerator(42)

	// Verify the generator has Turso flavor configured
	if gen.flavorConfig == nil {
		t.Fatal("Generator should have flavor config")
	}

	if gen.flavorConfig.Name() != "turso" {
		t.Errorf("Generator should use Turso flavor, got '%s'", gen.flavorConfig.Name())
	}

	// Verify that window functions are not supported
	if gen.flavorConfig.SupportsFeature("window_functions") {
		t.Error("Generator's flavor should not support window functions")
	}

	// Verify that recursive CTEs are not supported
	if gen.flavorConfig.SupportsFeature("cte_recursive") {
		t.Error("Generator's flavor should not support recursive CTEs")
	}
}
