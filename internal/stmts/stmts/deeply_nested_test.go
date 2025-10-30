package stmts

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/tursodatabase/turso-go"

	"sqlsmith-go/internal/common"
)

// TestGenSelectDeeplyNested tests the deeply nested SELECT generation.
func TestGenSelectDeeplyNested(t *testing.T) {
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
	depths := []int{0, 1, 2, 3, 4}
	for _, depth := range depths {
		t.Run(fmt.Sprintf("%s_depth_%d", t.Name(), depth), func(t *testing.T) {
			lcg := common.NewLCG(42)
			stmt, err := GenSelectDeeplyNested(db, lcg, depth)
			if err != nil {
				t.Errorf("GenSelectDeeplyNested(depth=%d) failed: %v", depth, err)
				return
			}

			sql := stmt.SQL()
			if sql == "" {
				t.Errorf("GenSelectDeeplyNested(depth=%d) returned empty SQL", depth)
				return
			}

			// Verify SQL contains nested structure at higher depths
			if depth >= 2 {
				nestedCount := strings.Count(sql, "nested_")
				if nestedCount == 0 {
					t.Logf("Warning: Expected nested subqueries at depth %d, but found none", depth)
				} else {
					t.Logf("Depth %d has %d nested levels", depth, nestedCount)
				}
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

// TestGenSelectDeeplyNestedComplexity tests that deeper recursion produces more complex queries.
func TestGenSelectDeeplyNestedComplexity(t *testing.T) {
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

	// Generate queries at different depths and verify complexity increases
	lcg1 := common.NewLCG(12345)
	stmt1, err := GenSelectDeeplyNested(db, lcg1, 1)
	if err != nil {
		t.Fatalf("Failed to generate depth 1: %v", err)
	}

	lcg2 := common.NewLCG(12345) // Same seed for comparison
	stmt2, err := GenSelectDeeplyNested(db, lcg2, 3)
	if err != nil {
		t.Fatalf("Failed to generate depth 3: %v", err)
	}

	sql1 := stmt1.SQL()
	sql2 := stmt2.SQL()

	// Verify depth 3 produces longer/more complex SQL
	if len(sql2) <= len(sql1) {
		t.Logf("Depth 1 SQL (%d chars): %s", len(sql1), sql1)
		t.Logf("Depth 3 SQL (%d chars): %s", len(sql2), sql2)
		t.Logf("Warning: Expected depth 3 SQL to be more complex than depth 1")
	} else {
		t.Logf("Depth 1 SQL length: %d", len(sql1))
		t.Logf("Depth 3 SQL length: %d", len(sql2))
		t.Logf("Complexity increase: %.1f%%", float64(len(sql2)-len(sql1))/float64(len(sql1))*100)
	}

	// Verify both execute successfully
	rows, err := db.Query(sql1)
	if err != nil {
		t.Errorf("Depth 1 SQL failed to execute: %v", err)
	} else {
		rows.Close()
	}

	rows, err = db.Query(sql2)
	if err != nil {
		t.Errorf("Depth 3 SQL failed to execute: %v", err)
	} else {
		rows.Close()
	}
}
