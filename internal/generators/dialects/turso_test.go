package dialects_test

import (
	"testing"

	"sqlfuse/internal/generators/dialects"
)

func TestTursoFlavorConfig(t *testing.T) {
	config := dialects.NewTursoFlavorConfig()

	// Test Name
	if name := config.Name(); name != "turso" {
		t.Errorf("Expected flavor name 'turso', got '%s'", name)
	}

	// Test unsupported features
	unsupportedFeatures := []string{
		"exists_subquery",
		"in_subquery",
		"window_functions",
		"cte_recursive",
		"regexp",
	}

	for _, feature := range unsupportedFeatures {
		if config.SupportsFeature(feature) {
			t.Errorf("Feature '%s' should not be supported by Turso", feature)
		}
	}

	// Test supported features (default SQLite features)
	supportedFeatures := []string{
		"select",
		"insert",
		"update",
		"delete",
		"join",
	}

	for _, feature := range supportedFeatures {
		if !config.SupportsFeature(feature) {
			t.Errorf("Feature '%s' should be supported by Turso", feature)
		}
	}

	// Test ValidateSQL (should return nil as it's not implemented)
	if err := config.ValidateSQL("SELECT * FROM test"); err != nil {
		t.Errorf("ValidateSQL should return nil, got: %v", err)
	}
}
