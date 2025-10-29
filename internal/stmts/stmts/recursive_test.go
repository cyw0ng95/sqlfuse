package stmts

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/tursodatabase/turso-go"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
)

// TestGenSelectRecursive tests the recursive SELECT generation.
func TestGenSelectRecursive(t *testing.T) {
	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory DB: %v", err)
	}
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE orders (id INTEGER PRIMARY KEY, user_id INTEGER, amount REAL)`)
	if err != nil {
		t.Fatalf("Failed to create orders table: %v", err)
	}

	// Insert some test data
	_, err = db.Exec(`INSERT INTO users (name, age) VALUES ('Alice', 30), ('Bob', 25), ('Charlie', 35)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	_, err = db.Exec(`INSERT INTO orders (user_id, amount) VALUES (1, 100.5), (1, 200.0), (2, 150.75)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test with different recursion depths
	depths := []int{0, 1, 2, 3}
	for _, depth := range depths {
		t.Run(t.Name()+"_depth_"+string(rune('0'+depth)), func(t *testing.T) {
			lcg := common.NewLCG(42)
			stmt, err := GenSelectRecursive(db, lcg, depth)
			if err != nil {
				t.Errorf("GenSelectRecursive(depth=%d) failed: %v", depth, err)
				return
			}

			sql := stmt.SQL()
			if sql == "" {
				t.Errorf("GenSelectRecursive(depth=%d) returned empty SQL", depth)
				return
			}

			// Verify SQL is valid by attempting to execute it
			rows, err := db.Query(sql)
			if err != nil {
				t.Logf("Generated SQL: %s", sql)
				t.Errorf("Generated SQL (depth=%d) failed to execute: %v", depth, err)
				return
			}
			rows.Close()

			t.Logf("Depth %d SQL: %s", depth, sql)
		})
	}
}

// TestGenSelectWithNestedCase tests nested CASE expression generation.
func TestGenSelectWithNestedCase(t *testing.T) {
	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory DB: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(`CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, price REAL, stock INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create products table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO products (name, price, stock) VALUES ('Widget', 10.5, 100), ('Gadget', 20.0, 50)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test with different recursion depths
	for depth := 0; depth <= 3; depth++ {
		lcg := common.NewLCG(42)
		stmt, err := GenSelectWithNestedCase(db, lcg, depth)
		if err != nil {
			t.Errorf("GenSelectWithNestedCase(depth=%d) failed: %v", depth, err)
			continue
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Errorf("GenSelectWithNestedCase(depth=%d) returned empty SQL", depth)
			continue
		}

		// Verify it contains CASE
		if !strings.Contains(strings.ToUpper(sql), "CASE") {
			t.Errorf("GenSelectWithNestedCase(depth=%d) did not generate CASE: %s", depth, sql)
			continue
		}

		// Verify SQL is valid
		rows, err := db.Query(sql)
		if err != nil {
			t.Logf("Generated SQL: %s", sql)
			t.Errorf("Generated SQL (depth=%d) failed to execute: %v", depth, err)
			continue
		}
		rows.Close()

		t.Logf("Depth %d SQL: %s", depth, sql)
	}
}

// TestGenSelectWithComplexJoin tests complex join generation.
func TestGenSelectWithComplexJoin(t *testing.T) {
	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory DB: %v", err)
	}
	defer db.Close()

	// Create test tables with common column
	_, err = db.Exec(`CREATE TABLE departments (id INTEGER PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create departments table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE employees (id INTEGER PRIMARY KEY, name TEXT, department_id INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create employees table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO departments (id, name) VALUES (1, 'Engineering'), (2, 'Sales')`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	_, err = db.Exec(`INSERT INTO employees (name, department_id) VALUES ('Alice', 1), ('Bob', 2)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test with different recursion depths
	for depth := 0; depth <= 2; depth++ {
		lcg := common.NewLCG(42)
		stmt, err := GenSelectWithComplexJoin(db, lcg, depth)
		if err != nil {
			t.Errorf("GenSelectWithComplexJoin(depth=%d) failed: %v", depth, err)
			continue
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Errorf("GenSelectWithComplexJoin(depth=%d) returned empty SQL", depth)
			continue
		}

		// Verify it contains JOIN
		if !strings.Contains(strings.ToUpper(sql), "JOIN") {
			t.Errorf("GenSelectWithComplexJoin(depth=%d) did not generate JOIN: %s", depth, sql)
			continue
		}

		// Verify SQL is valid
		rows, err := db.Query(sql)
		if err != nil {
			t.Logf("Generated SQL: %s", sql)
			t.Errorf("Generated SQL (depth=%d) failed to execute: %v", depth, err)
			continue
		}
		rows.Close()

		t.Logf("Depth %d SQL: %s", depth, sql)
	}
}

// TestExprGeneratorWhereExpr tests recursive WHERE expression generation.
func TestExprGeneratorWhereExpr(t *testing.T) {
	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory DB: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test_table (id INTEGER, value TEXT, score REAL)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO test_table VALUES (1, 'test', 100.0)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Get table info
	tbls, err := getTestTables(db)
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}

	// Test with different depths
	for depth := 0; depth <= 3; depth++ {
		lcg := common.NewLCG(42)
		ctx := NewGenContext(db, lcg, depth)
		exprGen := NewExprGenerator(ctx)

		whereExpr := exprGen.GenWhereExpr(tbls)
		if whereExpr == "" && depth > 0 {
			// Allow empty expressions at depth 0, but expect something at higher depths
			t.Logf("Warning: GenWhereExpr(depth=%d) returned empty expression", depth)
		}

		// Try to use the expression in a query
		if whereExpr != "" {
			sql := "SELECT * FROM test_table WHERE " + whereExpr
			rows, err := db.Query(sql)
			if err != nil {
				t.Logf("Generated WHERE: %s", whereExpr)
				t.Errorf("WHERE expression (depth=%d) failed to execute: %v", depth, err)
				continue
			}
			rows.Close()

			t.Logf("Depth %d WHERE: %s", depth, whereExpr)
		}
	}
}

// TestGenContextDepthControl tests depth tracking and control.
func TestGenContextDepthControl(t *testing.T) {
	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory DB: %v", err)
	}
	defer db.Close()

	lcg := common.NewLCG(42)

	// Test basic context creation
	ctx := NewGenContext(db, lcg, 3)
	if ctx.Depth != 0 {
		t.Errorf("New context should have depth 0, got %d", ctx.Depth)
	}
	if ctx.MaxDepth != 3 {
		t.Errorf("New context should have max depth 3, got %d", ctx.MaxDepth)
	}
	if !ctx.CanRecurse() {
		t.Error("New context with max depth 3 should be able to recurse")
	}

	// Test descending
	ctx2 := ctx.Descend()
	if ctx2.Depth != 1 {
		t.Errorf("Descended context should have depth 1, got %d", ctx2.Depth)
	}
	if !ctx2.CanRecurse() {
		t.Error("Descended context (depth 1, max 3) should be able to recurse")
	}

	// Test reaching max depth
	ctx3 := ctx2.Descend().Descend()
	if ctx3.Depth != 3 {
		t.Errorf("Twice-descended context should have depth 3, got %d", ctx3.Depth)
	}
	if ctx3.CanRecurse() {
		t.Error("Context at max depth should not be able to recurse")
	}
}

// Helper function to get table info for testing
func getTestTables(db *sql.DB) ([]helper.TableInfo, error) {
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []helper.TableInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}

		// Simplified: just get the table name without columns
		tables = append(tables, helper.TableInfo{Name: name, Cols: []helper.ColumnInfo{{Name: "id", Type: "INTEGER"}}})
	}
	return tables, nil
}
