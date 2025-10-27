package stmts

import (
	"sqlsmith-go/internal/common"
	"strings"
	"testing"
)

// MockFlavorConfig is a test implementation of FlavorConfig
type MockFlavorConfig struct {
	name              string
	supportedFeatures map[string]bool
}

func (m *MockFlavorConfig) Name() string {
	return m.name
}

func (m *MockFlavorConfig) SupportsFeature(feature string) bool {
	if m.supportedFeatures == nil {
		return true
	}
	supported, exists := m.supportedFeatures[feature]
	if !exists {
		return true // Default to supported
	}
	return supported
}

func (m *MockFlavorConfig) ValidateSQL(sql string) error {
	return nil
}

// TestFlavorConfig tests basic FlavorConfig functionality
func TestFlavorConfig(t *testing.T) {
	// Test default flavor (supports everything)
	defaultFlavor := &DefaultFlavorConfig{}
	if defaultFlavor.Name() != "sqlite" {
		t.Errorf("Default flavor name should be 'sqlite', got %s", defaultFlavor.Name())
	}
	if !defaultFlavor.SupportsFeature("window_functions") {
		t.Error("Default flavor should support window_functions")
	}
	if !defaultFlavor.SupportsFeature("cte_recursive") {
		t.Error("Default flavor should support cte_recursive")
	}

	// Test custom flavor with restrictions
	restrictiveFlavor := &MockFlavorConfig{
		name: "test_db",
		supportedFeatures: map[string]bool{
			"window_functions": false,
			"cte_recursive":    false,
		},
	}
	if restrictiveFlavor.Name() != "test_db" {
		t.Errorf("Restrictive flavor name should be 'test_db', got %s", restrictiveFlavor.Name())
	}
	if restrictiveFlavor.SupportsFeature("window_functions") {
		t.Error("Restrictive flavor should not support window_functions")
	}
	if restrictiveFlavor.SupportsFeature("cte_recursive") {
		t.Error("Restrictive flavor should not support cte_recursive")
	}
	// Unknown features should default to supported
	if !restrictiveFlavor.SupportsFeature("unknown_feature") {
		t.Error("Unknown features should default to supported")
	}
}

// TestGenContextWithFlavor tests GenContext with different flavors
func TestGenContextWithFlavor(t *testing.T) {
	lcg := common.NewLCG(42)

	// Test default context (no flavor specified)
	ctx1 := NewGenContext(nil, lcg, 2)
	if ctx1.Flavor == nil {
		t.Error("Context should have default flavor")
	}
	if ctx1.Flavor.Name() != "sqlite" {
		t.Errorf("Default context should have sqlite flavor, got %s", ctx1.Flavor.Name())
	}
	if !ctx1.SupportsFeature("window_functions") {
		t.Error("Default context should support window_functions")
	}

	// Test context with custom flavor
	customFlavor := &MockFlavorConfig{
		name: "custom",
		supportedFeatures: map[string]bool{
			"window_functions": false,
		},
	}
	ctx2 := NewGenContextWithFlavor(nil, lcg, 2, customFlavor)
	if ctx2.Flavor.Name() != "custom" {
		t.Errorf("Context should have custom flavor, got %s", ctx2.Flavor.Name())
	}
	if ctx2.SupportsFeature("window_functions") {
		t.Error("Custom context should not support window_functions")
	}

	// Test that Descend() preserves flavor
	ctx3 := ctx2.Descend()
	if ctx3.Flavor.Name() != "custom" {
		t.Error("Descended context should preserve flavor")
	}
	if ctx3.SupportsFeature("window_functions") {
		t.Error("Descended context should preserve flavor restrictions")
	}
}

// TestRecursiveCTEWithFlavor tests that recursive CTE generation respects flavor
func TestRecursiveCTEWithFlavor(t *testing.T) {
	lcg := common.NewLCG(42)

	// Test with flavor that supports RECURSIVE
	supportiveFlavor := &DefaultFlavorConfig{}
	ctx1 := NewGenContextWithFlavor(nil, lcg, 2, supportiveFlavor)
	stmt1, err := genSelectWithRecursiveCTEInternal(ctx1)
	if err != nil {
		t.Fatalf("Failed to generate recursive CTE with supportive flavor: %v", err)
	}
	if !strings.Contains(stmt1.SQL(), "RECURSIVE") {
		t.Error("Supportive flavor should generate RECURSIVE CTE")
	}

	// Test with flavor that does NOT support RECURSIVE (like Turso)
	restrictiveFlavor := &MockFlavorConfig{
		name: "turso",
		supportedFeatures: map[string]bool{
			"cte_recursive": false,
		},
	}
	ctx2 := NewGenContextWithFlavor(nil, lcg, 2, restrictiveFlavor)
	stmt2, err := genSelectWithRecursiveCTEInternal(ctx2)
	if err != nil {
		t.Fatalf("Failed to generate CTE with restrictive flavor: %v", err)
	}
	if strings.Contains(stmt2.SQL(), "RECURSIVE") {
		t.Error("Restrictive flavor should NOT generate RECURSIVE CTE")
	}
	// Should still have a WITH clause, just not recursive
	if !strings.Contains(stmt2.SQL(), "WITH") {
		t.Error("Should still generate WITH clause for non-recursive CTE")
	}
}

// TestWindowFunctionsWithFlavor tests that window function generation respects flavor
func TestWindowFunctionsWithFlavor(t *testing.T) {
	lcg := common.NewLCG(42)

	// Test with flavor that supports window functions
	supportiveFlavor := &DefaultFlavorConfig{}
	ctx1 := NewGenContextWithFlavor(nil, lcg, 2, supportiveFlavor)
	stmt1, err := genSelectWithWindowFunctionInternal(ctx1)
	if err != nil {
		t.Fatalf("Failed to generate window function with supportive flavor: %v", err)
	}
	sql1 := stmt1.SQL()
	// When no tables exist, it generates literal queries
	// Check if it's either a literal or contains OVER
	if !strings.Contains(sql1, "OVER") && !strings.Contains(sql1, "ROW_NUMBER") {
		t.Error("Supportive flavor should generate window functions or literals")
	}

	// Test with flavor that does NOT support window functions (like Turso)
	restrictiveFlavor := &MockFlavorConfig{
		name: "turso",
		supportedFeatures: map[string]bool{
			"window_functions": false,
		},
	}
	ctx2 := NewGenContextWithFlavor(nil, lcg, 2, restrictiveFlavor)
	stmt2, err := genSelectWithWindowFunctionInternal(ctx2)
	if err != nil {
		t.Fatalf("Failed to generate query with restrictive flavor: %v", err)
	}
	sql2 := stmt2.SQL()
	if strings.Contains(sql2, "OVER (") {
		t.Errorf("Restrictive flavor should NOT generate OVER clause, got: %s", sql2)
	}
	// Should still generate a valid SELECT
	if !strings.Contains(sql2, "SELECT") {
		t.Error("Should still generate a SELECT query")
	}
}

// TestMultipleWindowsWithFlavor tests that multiple window functions respect flavor
func TestMultipleWindowsWithFlavor(t *testing.T) {
	lcg := common.NewLCG(42)

	// Test with flavor that does NOT support window functions
	restrictiveFlavor := &MockFlavorConfig{
		name: "turso",
		supportedFeatures: map[string]bool{
			"window_functions": false,
		},
	}
	ctx := NewGenContextWithFlavor(nil, lcg, 2, restrictiveFlavor)
	stmt, err := genSelectWithMultipleWindowsInternal(ctx)
	if err != nil {
		t.Fatalf("Failed to generate query with restrictive flavor: %v", err)
	}
	sql := stmt.SQL()
	if strings.Contains(sql, "OVER (") {
		t.Errorf("Restrictive flavor should NOT generate OVER clause, got: %s", sql)
	}
	// Should still generate a valid SELECT, possibly with GROUP BY as alternative
	if !strings.Contains(sql, "SELECT") {
		t.Error("Should still generate a SELECT query")
	}
}

// TestFlavorConfigBackwardCompatibility ensures old code still works
func TestFlavorConfigBackwardCompatibility(t *testing.T) {
	lcg := common.NewLCG(42)

	// Old-style function calls should still work
	stmt1, err := GenSelectWithRecursiveCTE(nil, lcg)
	if err != nil {
		t.Fatalf("Old-style GenSelectWithRecursiveCTE should still work: %v", err)
	}
	// With default flavor, should generate RECURSIVE
	if !strings.Contains(stmt1.SQL(), "RECURSIVE") && !strings.Contains(stmt1.SQL(), "WITH") {
		t.Error("Old-style call should generate recursive or WITH clause")
	}

	stmt2, err := GenSelectWithWindowFunction(nil, lcg)
	if err != nil {
		t.Fatalf("Old-style GenSelectWithWindowFunction should still work: %v", err)
	}
	// Should generate some kind of SELECT
	if !strings.Contains(stmt2.SQL(), "SELECT") {
		t.Error("Old-style call should generate SELECT")
	}
}
