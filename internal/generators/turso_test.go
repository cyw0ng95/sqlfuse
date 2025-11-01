package generators

import (
	"sqlfuse/internal/stmts/stmts"
	"testing"
)

func TestNewTursoGenerator(t *testing.T) {
	gen := NewTursoGenerator(42)
	
	if gen == nil {
		t.Fatal("NewTursoGenerator should not return nil")
	}
	
	if gen.BaseGenerator == nil {
		t.Error("BaseGenerator should be initialized")
	}
	
	if gen.flavorConfig == nil {
		t.Error("FlavorConfig should be initialized")
	}
}

func TestTursoGenerator_Name(t *testing.T) {
	gen := NewTursoGenerator(1)
	
	if gen.Name() != "turso" {
		t.Errorf("Expected name 'turso', got '%s'", gen.Name())
	}
}

func TestTursoGenerator_SupportedStmts(t *testing.T) {
	gen := NewTursoGenerator(1)
	
	stmts := gen.SupportedStmts()
	
	if len(stmts) == 0 {
		t.Error("SupportedStmts should return non-empty map")
	}
	
	// Check for some expected statement types
	expectedStmts := []string{
		"insert",
		"select_basic",
		"update",
		"delete",
		"create_table",
		"pragma",
	}
	
	for _, stmt := range expectedStmts {
		if _, ok := stmts[stmt]; !ok {
			t.Errorf("Expected statement type '%s' in supported statements", stmt)
		}
	}
}

func TestDefaultTursoStmtWeights(t *testing.T) {
	weights := DefaultTursoStmtWeights()
	
	if len(weights) == 0 {
		t.Error("DefaultTursoStmtWeights should return non-empty map")
	}
	
	// Verify some key statement weights are set
	if weights[stmts.StmtInsert] == 0 {
		t.Error("StmtInsert should have non-zero weight")
	}
	
	if weights[stmts.StmtSelectBasic] == 0 {
		t.Error("StmtSelectBasic should have non-zero weight")
	}
	
	if weights[stmts.StmtUpdate] == 0 {
		t.Error("StmtUpdate should have non-zero weight")
	}
	
	if weights[stmts.StmtDelete] == 0 {
		t.Error("StmtDelete should have non-zero weight")
	}
}

func TestTursoGenerator_WeightsInitialized(t *testing.T) {
	gen := NewTursoGenerator(1)
	
	weights := gen.GetWeights()
	
	if len(weights) == 0 {
		t.Error("Generator should have weights initialized by default")
	}
	
	// Verify weights match default weights
	defaultWeights := DefaultTursoStmtWeights()
	
	for stmtType, expectedWeight := range defaultWeights {
		if weights[stmtType] != expectedWeight {
			t.Errorf("Weight for %s: expected %d, got %d", stmtType, expectedWeight, weights[stmtType])
		}
	}
}

func TestTursoGenerator_CustomWeights(t *testing.T) {
	gen := NewTursoGenerator(1)
	
	// Set custom weights
	customWeights := map[stmts.StmtType]uint64{
		stmts.StmtInsert:      1000,
		stmts.StmtSelectBasic: 2000,
	}
	
	gen.SetWeights(customWeights)
	
	weights := gen.GetWeights()
	
	// Should have custom weights plus defaults
	if weights[stmts.StmtInsert] != 1000 {
		t.Errorf("Expected custom Insert weight 1000, got %d", weights[stmts.StmtInsert])
	}
	
	if weights[stmts.StmtSelectBasic] != 2000 {
		t.Errorf("Expected custom Select weight 2000, got %d", weights[stmts.StmtSelectBasic])
	}
}

func TestTursoGenerator_FlavorConfig(t *testing.T) {
	gen := NewTursoGenerator(1)
	
	if gen.flavorConfig.Name() != "turso" {
		t.Errorf("Expected flavor 'turso', got '%s'", gen.flavorConfig.Name())
	}
}
