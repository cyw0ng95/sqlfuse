package generators

import (
	"sqlfuse/internal/stmts/stmts"
	"testing"
)

func TestNewGoSQLite3Generator(t *testing.T) {
	gen := NewGoSQLite3Generator(42)
	
	if gen == nil {
		t.Fatal("NewGoSQLite3Generator should not return nil")
	}
	
	if gen.BaseGenerator == nil {
		t.Error("BaseGenerator should be initialized")
	}
	
	if gen.flavorConfig == nil {
		t.Error("FlavorConfig should be initialized")
	}
}

func TestGoSQLite3Generator_Name(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	
	if gen.Name() != "go-sqlite3" {
		t.Errorf("Expected name 'go-sqlite3', got '%s'", gen.Name())
	}
}

func TestGoSQLite3Generator_SupportedStmts(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	
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
		"select_window",    // go-sqlite3 supports window functions
		"select_cte",       // go-sqlite3 supports CTEs
	}
	
	for _, stmt := range expectedStmts {
		if _, ok := stmts[stmt]; !ok {
			t.Errorf("Expected statement type '%s' in supported statements", stmt)
		}
	}
}

func TestDefaultGoSQLite3StmtWeights(t *testing.T) {
	weights := DefaultGoSQLite3StmtWeights()
	
	if len(weights) == 0 {
		t.Error("DefaultGoSQLite3StmtWeights should return non-empty map")
	}
	
	// Verify some key statement weights are set
	if weights[stmts.StmtInsert] == 0 {
		t.Error("StmtInsert should have non-zero weight")
	}
	
	if weights[stmts.StmtSelectBasic] == 0 {
		t.Error("StmtSelectBasic should have non-zero weight")
	}
	
	// Verify go-sqlite3 has higher weights for advanced features
	// compared to Turso (which has restrictions)
	if weights[stmts.StmtSelectWindow] == 0 {
		t.Error("StmtSelectWindow should have non-zero weight (full support)")
	}
	
	if weights[stmts.StmtSelectCTE] == 0 {
		t.Error("StmtSelectCTE should have non-zero weight (full support)")
	}
	
	if weights[stmts.StmtSelectRecursiveCTE] == 0 {
		t.Error("StmtSelectRecursiveCTE should have non-zero weight (full support)")
	}
}

func TestGoSQLite3Generator_WeightsInitialized(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	
	weights := gen.GetWeights()
	
	if len(weights) == 0 {
		t.Error("Generator should have weights initialized by default")
	}
	
	// Verify weights match default weights
	defaultWeights := DefaultGoSQLite3StmtWeights()
	
	for stmtType, expectedWeight := range defaultWeights {
		if weights[stmtType] != expectedWeight {
			t.Errorf("Weight for %s: expected %d, got %d", stmtType, expectedWeight, weights[stmtType])
		}
	}
}

func TestGoSQLite3Generator_CustomWeights(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	
	// Set custom weights
	customWeights := map[stmts.StmtType]uint64{
		stmts.StmtInsert:       1000,
		stmts.StmtSelectWindow: 2000,
	}
	
	gen.SetWeights(customWeights)
	
	weights := gen.GetWeights()
	
	// Should have custom weights plus defaults
	if weights[stmts.StmtInsert] != 1000 {
		t.Errorf("Expected custom Insert weight 1000, got %d", weights[stmts.StmtInsert])
	}
	
	if weights[stmts.StmtSelectWindow] != 2000 {
		t.Errorf("Expected custom SelectWindow weight 2000, got %d", weights[stmts.StmtSelectWindow])
	}
}

func TestGoSQLite3Generator_FlavorConfig(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	
	if gen.flavorConfig.Name() != "go-sqlite3" {
		t.Errorf("Expected flavor 'go-sqlite3', got '%s'", gen.flavorConfig.Name())
	}
}

func TestGoSQLite3Generator_WindowFunctionsWeight(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	weights := gen.GetWeights()
	
	// go-sqlite3 should have higher weight for window functions than basic selects
	// since it fully supports them
	windowWeight := weights[stmts.StmtSelectWindow]
	
	if windowWeight == 0 {
		t.Error("Window functions should have non-zero weight in go-sqlite3")
	}
	
	// Should be at least 50 (higher than Turso's 35)
	if windowWeight < 50 {
		t.Errorf("Expected window functions weight >= 50, got %d", windowWeight)
	}
}

func TestGoSQLite3Generator_CTEWeight(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	weights := gen.GetWeights()
	
	// go-sqlite3 should have higher weight for CTEs
	cteWeight := weights[stmts.StmtSelectCTE]
	
	if cteWeight == 0 {
		t.Error("CTEs should have non-zero weight in go-sqlite3")
	}
	
	// Should be at least 50
	if cteWeight < 50 {
		t.Errorf("Expected CTE weight >= 50, got %d", cteWeight)
	}
}

func TestGoSQLite3Generator_PragmaWeight(t *testing.T) {
	gen := NewGoSQLite3Generator(1)
	weights := gen.GetWeights()
	
	// go-sqlite3 should have higher pragma weight (supports all pragmas)
	pragmaWeight := weights[stmts.StmtPragma]
	
	if pragmaWeight == 0 {
		t.Error("PRAGMA should have non-zero weight in go-sqlite3")
	}
	
	// Should be at least 50 (higher than Turso's 30)
	if pragmaWeight < 50 {
		t.Errorf("Expected PRAGMA weight >= 50, got %d", pragmaWeight)
	}
}
