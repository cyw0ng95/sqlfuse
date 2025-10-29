package stmts

import (
	"database/sql"
	"sqlsmith-go/internal/common"
	"testing"
)

func TestStmtGeneratorFactory_CreateGenerator(t *testing.T) {
	lcg := common.NewLCG(42)
	factory := NewStmtGeneratorFactory(lcg, 3, GetDefaultFlavor())

	tests := []struct {
		name     string
		stmtType StmtType
		wantNil  bool
	}{
		{"PRAGMA", StmtPragma, false},
		{"INSERT", StmtInsert, false},
		{"UPDATE", StmtUpdate, false},
		{"DELETE", StmtDelete, false},
		{"SELECT", StmtSelectBasic, false},
		{"CREATE TABLE", StmtCreateTable, false},
		{"DROP TABLE", StmtDropTable, false},
		{"ALTER TABLE", StmtAlterTable, false},
		{"INSERT MULTIPLE", StmtInsertMultiple, false},
		{"SELECT WHERE", StmtSelectWhere, false},
		{"SELECT JOIN", StmtSelectJoin, false},
		{"SELECT CTE", StmtSelectCTE, false},
		{"Invalid type", StmtType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := factory.CreateGenerator(tt.stmtType)
			if tt.wantNil && gen != nil {
				t.Errorf("CreateGenerator(%v) = %v, want nil", tt.stmtType, gen)
			}
			if !tt.wantNil && gen == nil {
				t.Errorf("CreateGenerator(%v) = nil, want non-nil", tt.stmtType)
			}
		})
	}
}

func TestStmtGeneratorFactory_GenerateStmt(t *testing.T) {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("sqlite3 driver not available")
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	lcg := common.NewLCG(42)
	factory := NewStmtGeneratorFactory(lcg, 3, GetDefaultFlavor())

	tests := []struct {
		name     string
		stmtType StmtType
		wantErr  bool
	}{
		{"PRAGMA", StmtPragma, false},
		{"SELECT", StmtSelectBasic, false},
		{"INSERT", StmtInsert, false},
		{"UPDATE", StmtUpdate, false},
		{"DELETE", StmtDelete, false},
		{"Invalid type", StmtType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := factory.GenerateStmt(db, tt.stmtType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateStmt(%v) error = %v, wantErr %v", tt.stmtType, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if stmt == nil {
					t.Errorf("GenerateStmt(%v) returned nil statement", tt.stmtType)
					return
				}
				sql := stmt.SQL()
				if sql == "" {
					t.Errorf("GenerateStmt(%v) returned empty SQL", tt.stmtType)
				}
			}
		})
	}
}

func TestStmtGeneratorFactory_CreateContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("sqlite3 driver not available")
	}
	defer db.Close()

	lcg := common.NewLCG(42)
	maxDepth := 5
	flavor := GetDefaultFlavor()
	factory := NewStmtGeneratorFactory(lcg, maxDepth, flavor)

	ctx := factory.CreateContext(db)
	if ctx == nil {
		t.Fatal("CreateContext returned nil")
	}
	if ctx.DB != db {
		t.Error("Context DB mismatch")
	}
	if ctx.LCG != lcg {
		t.Error("Context LCG mismatch")
	}
	if ctx.MaxDepth != maxDepth {
		t.Errorf("Context MaxDepth = %d, want %d", ctx.MaxDepth, maxDepth)
	}
	if ctx.Flavor != flavor {
		t.Error("Context Flavor mismatch")
	}
}

func TestInsertVariantGenerator(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("sqlite3 driver not available")
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	lcg := common.NewLCG(42)
	ctx := NewGenContext(db, lcg, 3)

	variants := []StmtType{
		StmtInsert,
		StmtInsertMultiple,
		StmtInsertBulk,
		StmtInsertOrReplace,
		StmtInsertOrIgnore,
		StmtInsertOrAbort,
		StmtInsertOrRollback,
		StmtInsertOrFail,
	}

	for _, variant := range variants {
		t.Run(string(variant), func(t *testing.T) {
			gen := &InsertVariantGenerator{variant: variant}
			if !gen.CanGenerate(ctx) {
				t.Skip("Generator cannot generate with current context")
			}
			stmt, err := gen.Generate(ctx)
			if err != nil {
				t.Errorf("Generate(%v) error = %v", variant, err)
				return
			}
			if stmt == nil {
				t.Errorf("Generate(%v) returned nil statement", variant)
				return
			}
			sql := stmt.SQL()
			if sql == "" {
				t.Errorf("Generate(%v) returned empty SQL", variant)
			}
		})
	}
}

func TestSelectVariantGenerator(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("sqlite3 driver not available")
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	lcg := common.NewLCG(42)
	ctx := NewGenContext(db, lcg, 3)

	variants := []StmtType{
		StmtSelectBasic,
		StmtSelectWhere,
		StmtSelectLimit,
		StmtSelectOrder,
		StmtSelectGroup,
		StmtSelectJoin,
	}

	for _, variant := range variants {
		t.Run(string(variant), func(t *testing.T) {
			gen := &SelectVariantGenerator{variant: variant, maxDepth: 3}
			if !gen.CanGenerate(ctx) {
				t.Skip("Generator cannot generate with current context")
			}
			stmt, err := gen.Generate(ctx)
			if err != nil {
				t.Errorf("Generate(%v) error = %v", variant, err)
				return
			}
			if stmt == nil {
				t.Errorf("Generate(%v) returned nil statement", variant)
				return
			}
			sql := stmt.SQL()
			if sql == "" {
				t.Errorf("Generate(%v) returned empty SQL", variant)
			}
		})
	}
}

func TestUnsupportedStmtTypeError(t *testing.T) {
	err := &UnsupportedStmtTypeError{StmtType: StmtType("invalid")}
	expected := "unsupported statement type: invalid"
	if err.Error() != expected {
		t.Errorf("Error message = %q, want %q", err.Error(), expected)
	}
}

func TestCannotGenerateError(t *testing.T) {
	err := &CannotGenerateError{StmtType: StmtInsert, Reason: "no tables"}
	expected := "cannot generate insert: no tables"
	if err.Error() != expected {
		t.Errorf("Error message = %q, want %q", err.Error(), expected)
	}
}
