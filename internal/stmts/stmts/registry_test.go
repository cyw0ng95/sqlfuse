package stmts

import (
	"testing"

	"sqlfuse/internal/common"
)

// TestGeneratorRegistry tests the generator registry functionality
func TestGeneratorRegistry(t *testing.T) {
	reg := NewGeneratorRegistry()

	// Test empty registry
	if len(reg.Names()) != 0 {
		t.Errorf("New registry should be empty, got %d generators", len(reg.Names()))
	}

	// Test registration
	reg.Register("test", &PragmaGenerator{})
	if !reg.Has("test") {
		t.Error("Registry should have 'test' generator after registration")
	}

	// Test retrieval
	gen := reg.Get("test")
	if gen == nil {
		t.Error("Get should return non-nil generator")
	}
	if _, ok := gen.(*PragmaGenerator); !ok {
		t.Error("Retrieved generator should be PragmaGenerator")
	}

	// Test missing generator
	if reg.Get("nonexistent") != nil {
		t.Error("Get should return nil for non-existent generator")
	}

	// Test Names
	names := reg.Names()
	if len(names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(names))
	}
	if names[0] != "test" {
		t.Errorf("Expected name 'test', got %s", names[0])
	}
}

// TestDefaultRegistry tests that the default registry has all standard generators
func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()

	expectedGenerators := []string{
		"pragma",
		"insert",
		"select",
		"update",
		"delete",
		"create_table",
		"drop_table",
		"alter_table",
		"create_view",
		"drop_view",
	}

	for _, name := range expectedGenerators {
		if !reg.Has(name) {
			t.Errorf("Default registry should have %s generator", name)
		}
		gen := reg.Get(name)
		if gen == nil {
			t.Errorf("Default registry should return non-nil generator for %s", name)
		}
	}

	names := reg.Names()
	if len(names) != len(expectedGenerators) {
		t.Errorf("Expected %d generators, got %d", len(expectedGenerators), len(names))
	}
}

// TestGenerateStmt tests the convenience function for generating statements
func TestGenerateStmt(t *testing.T) {
	// Setup a test database for generators that need it
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(42)
	ctx := NewGenContext(db, lcg, 2)

	// Test generating a pragma statement
	stmt, err := GenerateStmt(ctx, "pragma")
	if err != nil {
		t.Fatalf("GenerateStmt failed: %v", err)
	}
	if stmt == nil {
		t.Error("GenerateStmt should return non-nil statement")
	}
	if stmt.Type() != "pragma" {
		t.Errorf("Expected pragma statement, got %s", stmt.Type())
	}

	// Test generating a create table statement
	stmt, err = GenerateStmt(ctx, "create_table")
	if err != nil {
		t.Fatalf("GenerateStmt failed: %v", err)
	}
	if stmt == nil {
		t.Error("GenerateStmt should return non-nil statement")
	}
	if stmt.Type() != "create_table" {
		t.Errorf("Expected create_table statement, got %s", stmt.Type())
	}

	// Test with non-existent generator
	_, err = GenerateStmt(ctx, "nonexistent")
	if err == nil {
		t.Error("GenerateStmt should return error for non-existent generator")
	}
}

// TestStmtGeneratorInterface tests that all standard generators implement StmtGenerator
func TestStmtGeneratorInterface(t *testing.T) {
	// Setup a test database for generators that need it
	db := setupTestDB(t)
	defer db.Close()

	generators := []struct {
		name string
		gen  StmtGenerator
	}{
		{"PragmaGenerator", &PragmaGenerator{}},
		{"InsertGenerator", &InsertGenerator{}},
		{"SelectGenerator", &SelectGenerator{}},
		{"UpdateGenerator", &UpdateGenerator{}},
		{"DeleteGenerator", &DeleteGenerator{}},
		{"CreateTableGenerator", &CreateTableGenerator{}},
		{"DropTableGenerator", &DropTableGenerator{}},
		{"AlterTableGenerator", &AlterTableGenerator{}},
		{"CreateViewGenerator", &CreateViewGenerator{}},
		{"DropViewGenerator", &DropViewGenerator{}},
	}

	lcg := common.NewLCG(123)
	ctx := NewGenContext(db, lcg, 2)

	for _, tc := range generators {
		t.Run(tc.name, func(t *testing.T) {
			// Test CanGenerate
			canGen := tc.gen.CanGenerate(ctx)
			t.Logf("%s.CanGenerate() = %v", tc.name, canGen)

			// If CanGenerate returns true, test Generate
			if canGen {
				stmt, err := tc.gen.Generate(ctx)
				if err != nil {
					t.Errorf("%s.Generate() failed: %v", tc.name, err)
				}
				if stmt == nil {
					t.Errorf("%s.Generate() returned nil statement", tc.name)
				}
				if stmt != nil && stmt.SQL() == "" {
					t.Errorf("%s.Generate() returned empty SQL", tc.name)
				}
			}
		})
	}
}

// TestGeneratorFuncWrapper tests the GeneratorFunc wrapper
func TestGeneratorFuncWrapper(t *testing.T) {
	// Create a simple generator function
	genFunc := GeneratorFunc(func(ctx *GenContext) (Stmt, error) {
		return &PragmaStmt{sql: "PRAGMA test;"}, nil
	})

	lcg := common.NewLCG(456)
	ctx := NewGenContext(nil, lcg, 2)

	// Test CanGenerate (should always return true for GeneratorFunc)
	if !genFunc.CanGenerate(ctx) {
		t.Error("GeneratorFunc.CanGenerate() should return true")
	}

	// Test Generate
	stmt, err := genFunc.Generate(ctx)
	if err != nil {
		t.Fatalf("GeneratorFunc.Generate() failed: %v", err)
	}
	if stmt == nil {
		t.Error("GeneratorFunc.Generate() returned nil statement")
	}
	if stmt.SQL() != "PRAGMA test;" {
		t.Errorf("Expected 'PRAGMA test;', got %s", stmt.SQL())
	}
}

// TestStmtGeneratorWithCheck tests the StmtGeneratorWithCheck wrapper
func TestStmtGeneratorWithCheck(t *testing.T) {
	// Create a generator with a custom CanGenerate check
	callCount := 0
	gen := &StmtGeneratorWithCheck{
		GenerateFn: func(ctx *GenContext) (Stmt, error) {
			callCount++
			return &PragmaStmt{sql: "PRAGMA custom;"}, nil
		},
		CanGenerateFn: func(ctx *GenContext) bool {
			// Only allow generation if depth is 0
			return ctx.Depth == 0
		},
	}

	lcg := common.NewLCG(789)

	// Test with depth 0 (should succeed)
	ctx1 := NewGenContext(nil, lcg, 2)
	if !gen.CanGenerate(ctx1) {
		t.Error("CanGenerate should return true for depth 0")
	}
	stmt1, err := gen.Generate(ctx1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if stmt1 == nil {
		t.Error("Generate returned nil statement")
	}

	// Test with depth 1 (should fail CanGenerate)
	ctx2 := ctx1.Descend()
	if gen.CanGenerate(ctx2) {
		t.Error("CanGenerate should return false for depth > 0")
	}

	// Verify callCount
	if callCount != 1 {
		t.Errorf("Expected Generate to be called once, got %d", callCount)
	}
}
