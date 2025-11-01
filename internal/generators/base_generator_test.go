package generators

import (
	"database/sql"
	"sqlfuse/internal/stmts/stmts"
	"testing"
)

func TestNewBaseGenerator(t *testing.T) {
	gen := NewBaseGenerator(42)
	
	if gen == nil {
		t.Fatal("NewBaseGenerator should not return nil")
	}
	
	if gen.lcg == nil {
		t.Error("LCG should be initialized")
	}
	
	if gen.maxRecursionDepth != 4 {
		t.Errorf("Expected default max recursion depth 4, got %d", gen.maxRecursionDepth)
	}
	
	if gen.impedanceMatcher == nil {
		t.Error("ImpedanceMatcher should be initialized")
	}
	
	if gen.stats == nil {
		t.Error("GenerationStats should be initialized")
	}
}

func TestBaseGenerator_SetWeights(t *testing.T) {
	gen := NewBaseGenerator(1)
	
	weights := map[stmts.StmtType]uint64{
		stmts.StmtInsert:      100,
		stmts.StmtSelectBasic: 200,
		stmts.StmtUpdate:      50,
	}
	
	gen.SetWeights(weights)
	
	retrieved := gen.GetWeights()
	
	if len(retrieved) != 3 {
		t.Errorf("Expected 3 weights, got %d", len(retrieved))
	}
	
	if retrieved[stmts.StmtInsert] != 100 {
		t.Errorf("Expected Insert weight 100, got %d", retrieved[stmts.StmtInsert])
	}
	
	if retrieved[stmts.StmtSelectBasic] != 200 {
		t.Errorf("Expected Select weight 200, got %d", retrieved[stmts.StmtSelectBasic])
	}
	
	if retrieved[stmts.StmtUpdate] != 50 {
		t.Errorf("Expected Update weight 50, got %d", retrieved[stmts.StmtUpdate])
	}
}

func TestBaseGenerator_SetWeight(t *testing.T) {
	gen := NewBaseGenerator(1)
	
	gen.SetWeight(stmts.StmtInsert, 150)
	gen.SetWeight(stmts.StmtSelectBasic, 250)
	
	weights := gen.GetWeights()
	
	if weights[stmts.StmtInsert] != 150 {
		t.Errorf("Expected Insert weight 150, got %d", weights[stmts.StmtInsert])
	}
	
	if weights[stmts.StmtSelectBasic] != 250 {
		t.Errorf("Expected Select weight 250, got %d", weights[stmts.StmtSelectBasic])
	}
}

func TestBaseGenerator_SetMaxRecursionDepth(t *testing.T) {
	gen := NewBaseGenerator(1)
	
	gen.SetMaxRecursionDepth(10)
	
	if gen.GetMaxRecursionDepth() != 10 {
		t.Errorf("Expected max recursion depth 10, got %d", gen.GetMaxRecursionDepth())
	}
}

func TestBaseGenerator_SetMaxRecursionDepth_Negative(t *testing.T) {
	gen := NewBaseGenerator(1)
	
	gen.SetMaxRecursionDepth(-5)
	
	if gen.GetMaxRecursionDepth() != 0 {
		t.Errorf("Expected negative depth to be clamped to 0, got %d", gen.GetMaxRecursionDepth())
	}
}

func TestBaseGenerator_GetLCG(t *testing.T) {
	gen := NewBaseGenerator(42)
	
	lcg := gen.GetLCG()
	
	if lcg == nil {
		t.Error("GetLCG should return non-nil LCG")
	}
	
	// Verify it's the same LCG instance
	v1 := lcg.Uint64()
	v2 := gen.GetLCG().Uint64()
	
	// Since we're using the same LCG, values should be different (sequential)
	if v1 == v2 {
		t.Error("LCG should generate different sequential values")
	}
}

func TestBaseGenerator_SetGenMap(t *testing.T) {
	gen := NewBaseGenerator(1)
	
	genMap := map[stmts.StmtType]func(db *sql.DB) (string, error){
		stmts.StmtInsert: func(db *sql.DB) (string, error) {
			return "INSERT INTO test VALUES (1)", nil
		},
	}
	
	// This should not panic
	gen.SetGenMap(genMap)
	
	if gen.genMap == nil {
		t.Error("GenMap should be set")
	}
}

func TestBaseGenerator_GetWeights_Copy(t *testing.T) {
	gen := NewBaseGenerator(1)
	
	weights := map[stmts.StmtType]uint64{
		stmts.StmtInsert: 100,
	}
	
	gen.SetWeights(weights)
	
	// Get a copy
	retrieved := gen.GetWeights()
	
	// Modify the retrieved copy
	retrieved[stmts.StmtInsert] = 999
	
	// Original should be unchanged
	original := gen.GetWeights()
	if original[stmts.StmtInsert] != 100 {
		t.Error("GetWeights should return a copy, not the original map")
	}
}

func TestBaseGenerator_SetWeights_Multiple(t *testing.T) {
	gen := NewBaseGenerator(1)
	
	// Set first batch of weights
	weights1 := map[stmts.StmtType]uint64{
		stmts.StmtInsert:      100,
		stmts.StmtSelectBasic: 200,
	}
	gen.SetWeights(weights1)
	
	// Set second batch (should replace)
	weights2 := map[stmts.StmtType]uint64{
		stmts.StmtUpdate: 300,
		stmts.StmtDelete: 400,
	}
	gen.SetWeights(weights2)
	
	// Should have all four statement types
	retrieved := gen.GetWeights()
	
	if len(retrieved) != 4 {
		t.Errorf("Expected 4 weights after multiple SetWeights, got %d", len(retrieved))
	}
	
	if retrieved[stmts.StmtInsert] != 100 {
		t.Errorf("Expected Insert weight 100, got %d", retrieved[stmts.StmtInsert])
	}
	
	if retrieved[stmts.StmtUpdate] != 300 {
		t.Errorf("Expected Update weight 300, got %d", retrieved[stmts.StmtUpdate])
	}
}

func TestBaseGenerator_TokensUsed(t *testing.T) {
	gen := NewBaseGenerator(1)
	lcg := gen.GetLCG()
	
	initialTokens := lcg.TokensUsed()
	
	// Use some tokens
	_ = lcg.Uint64()
	_ = lcg.Uint64()
	
	finalTokens := lcg.TokensUsed()
	
	if finalTokens != initialTokens+2 {
		t.Errorf("Expected %d tokens used, got %d", initialTokens+2, finalTokens)
	}
}
