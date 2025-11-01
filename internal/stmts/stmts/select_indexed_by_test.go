package stmts

import (
	"database/sql"
	"sqlfuse/internal/common"
	"strings"
	"testing"
	
	_ "github.com/tursodatabase/turso-go"
)

func getTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory DB: %v", err)
	}
	
	// Create a test table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS test_table (
		id INTEGER PRIMARY KEY,
		name TEXT,
		value INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	
	return db
}

func TestSelectIndexedBy(t *testing.T) {
	lcg := common.NewLCG(42)
	flavor := GetDefaultFlavor()
	
	// Test with nil DB (should fallback)
	stmt, err := genSelectIndexedBy(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("genSelectIndexedBy failed: %v", err)
	}
	if stmt == nil {
		t.Fatal("Statement is nil")
	}
	sql := stmt.SQL()
	if sql == "" {
		t.Fatal("Empty SQL")
	}
	t.Logf("Generated SQL (nil DB): %s", sql)

	// With test DB
	db := getTestDB(t)
	defer db.Close()
	
	stmt, err = genSelectIndexedBy(db, lcg, flavor)
	if err != nil {
		t.Fatalf("genSelectIndexedBy with DB failed: %v", err)
	}
	if stmt == nil {
		t.Fatal("Statement is nil")
	}
	sql = stmt.SQL()
	if sql == "" {
		t.Fatal("Empty SQL")
	}
	
	// Verify INDEXED BY clause is present
	if !strings.Contains(strings.ToUpper(sql), "INDEXED BY") {
		t.Errorf("SQL does not contain INDEXED BY clause: %s", sql)
	}
	
	// Verify it's a SELECT statement
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT") {
		t.Errorf("SQL is not a SELECT statement: %s", sql)
	}
	
	t.Logf("Generated SQL (with DB): %s", sql)
}

func TestSelectNotIndexed(t *testing.T) {
	lcg := common.NewLCG(42)
	flavor := GetDefaultFlavor()
	
	// Test with nil DB (should fallback)
	stmt, err := genSelectNotIndexed(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("genSelectNotIndexed failed: %v", err)
	}
	if stmt == nil {
		t.Fatal("Statement is nil")
	}
	sql := stmt.SQL()
	if sql == "" {
		t.Fatal("Empty SQL")
	}
	t.Logf("Generated SQL (nil DB): %s", sql)

	// With test DB
	db := getTestDB(t)
	defer db.Close()
	
	stmt, err = genSelectNotIndexed(db, lcg, flavor)
	if err != nil {
		t.Fatalf("genSelectNotIndexed with DB failed: %v", err)
	}
	if stmt == nil {
		t.Fatal("Statement is nil")
	}
	sql = stmt.SQL()
	if sql == "" {
		t.Fatal("Empty SQL")
	}
	
	// Verify NOT INDEXED clause is present
	if !strings.Contains(strings.ToUpper(sql), "NOT INDEXED") {
		t.Errorf("SQL does not contain NOT INDEXED clause: %s", sql)
	}
	
	// Verify it's a SELECT statement
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT") {
		t.Errorf("SQL is not a SELECT statement: %s", sql)
	}
	
	t.Logf("Generated SQL (with DB): %s", sql)
}

func TestSelectIndexedByGenerator(t *testing.T) {
	lcg := common.NewLCG(42)
	db := getTestDB(t)
	defer db.Close()
	
	gen := NewSelectIndexedByGenerator()
	ctx := NewGenContext(db, lcg, 10)
	
	// Test CanGenerate
	if !gen.CanGenerate(ctx) {
		t.Fatal("CanGenerate returned false with valid DB")
	}
	
	// Test Generate
	stmt, err := gen.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if stmt == nil {
		t.Fatal("Statement is nil")
	}
	
	sql := stmt.SQL()
	if sql == "" {
		t.Fatal("Empty SQL")
	}
	
	if !strings.Contains(strings.ToUpper(sql), "INDEXED BY") {
		t.Errorf("SQL does not contain INDEXED BY clause: %s", sql)
	}
	
	t.Logf("Generated SQL: %s", sql)
}

func TestSelectNotIndexedGenerator(t *testing.T) {
	lcg := common.NewLCG(42)
	db := getTestDB(t)
	defer db.Close()
	
	gen := NewSelectNotIndexedGenerator()
	ctx := NewGenContext(db, lcg, 10)
	
	// Test CanGenerate
	if !gen.CanGenerate(ctx) {
		t.Fatal("CanGenerate returned false with valid DB")
	}
	
	// Test Generate
	stmt, err := gen.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if stmt == nil {
		t.Fatal("Statement is nil")
	}
	
	sql := stmt.SQL()
	if sql == "" {
		t.Fatal("Empty SQL")
	}
	
	if !strings.Contains(strings.ToUpper(sql), "NOT INDEXED") {
		t.Errorf("SQL does not contain NOT INDEXED clause: %s", sql)
	}
	
	t.Logf("Generated SQL: %s", sql)
}

func TestIndexedByFactoryIntegration(t *testing.T) {
	lcg := common.NewLCG(42)
	flavor := GetDefaultFlavor()
	db := getTestDB(t)
	defer db.Close()
	
	factory := NewStmtGeneratorFactory(lcg, 10, flavor)
	
	// Test StmtSelectIndexedBy
	genIndexedBy := factory.CreateGenerator(StmtSelectIndexedBy)
	if genIndexedBy == nil {
		t.Fatal("Factory returned nil for StmtSelectIndexedBy")
	}
	
	stmtIndexedBy, err := factory.GenerateStmt(db, StmtSelectIndexedBy)
	if err != nil {
		t.Fatalf("GenerateStmt failed for StmtSelectIndexedBy: %v", err)
	}
	if stmtIndexedBy == nil {
		t.Fatal("Statement is nil for StmtSelectIndexedBy")
	}
	
	sql := stmtIndexedBy.SQL()
	if !strings.Contains(strings.ToUpper(sql), "INDEXED BY") {
		t.Errorf("SQL does not contain INDEXED BY: %s", sql)
	}
	t.Logf("StmtSelectIndexedBy SQL: %s", sql)
	
	// Test StmtSelectNotIndexed
	genNotIndexed := factory.CreateGenerator(StmtSelectNotIndexed)
	if genNotIndexed == nil {
		t.Fatal("Factory returned nil for StmtSelectNotIndexed")
	}
	
	stmtNotIndexed, err := factory.GenerateStmt(db, StmtSelectNotIndexed)
	if err != nil {
		t.Fatalf("GenerateStmt failed for StmtSelectNotIndexed: %v", err)
	}
	if stmtNotIndexed == nil {
		t.Fatal("Statement is nil for StmtSelectNotIndexed")
	}
	
	sql = stmtNotIndexed.SQL()
	if !strings.Contains(strings.ToUpper(sql), "NOT INDEXED") {
		t.Errorf("SQL does not contain NOT INDEXED: %s", sql)
	}
	t.Logf("StmtSelectNotIndexed SQL: %s", sql)
}
