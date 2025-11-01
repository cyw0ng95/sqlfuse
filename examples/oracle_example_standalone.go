//go:build ignore
// +build ignore

// This example demonstrates SQLancer-inspired oracle-based testing
// Build with: go run examples/oracle_example_standalone.go

package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Simplified oracle demonstration (without internal packages)
func main() {
	// Open database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create test schema
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		age INTEGER,
		name TEXT,
		balance REAL
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create schema: %v\n", err)
		os.Exit(1)
	}

	// Insert test data
	for i := 0; i < 100; i++ {
		_, err = db.Exec("INSERT INTO users (id, age, name, balance) VALUES (?, ?, ?, ?)",
			i, 20+i%50, fmt.Sprintf("user%d", i), float64(i*100))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to insert user data: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("SQLancer Oracle Testing Example")
	fmt.Println("================================")
	fmt.Println()

	// Demonstrate TLP (Ternary Logic Partitioning)
	fmt.Println("1. TLP (Ternary Logic Partitioning) Oracle")
	fmt.Println("   Tests that WHERE p + WHERE NOT p + WHERE p IS NULL = original count")
	fmt.Println()

	predicate := "age > 30"
	
	var originalCount, trueCount, falseCount, nullCount int64
	
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&originalCount)
	db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", predicate)).Scan(&trueCount)
	db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM users WHERE NOT (%s)", predicate)).Scan(&falseCount)
	db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM users WHERE (%s) IS NULL", predicate)).Scan(&nullCount)
	
	sum := trueCount + falseCount + nullCount
	fmt.Printf("   Predicate: %s\n", predicate)
	fmt.Printf("   Original count: %d\n", originalCount)
	fmt.Printf("   WHERE %s: %d\n", predicate, trueCount)
	fmt.Printf("   WHERE NOT (%s): %d\n", predicate, falseCount)
	fmt.Printf("   WHERE (%s) IS NULL: %d\n", predicate, nullCount)
	fmt.Printf("   Sum: %d\n", sum)
	
	if originalCount == sum {
		fmt.Println("   ✓ TLP check passed!")
	} else {
		fmt.Println("   ✗ TLP check failed - potential bug!")
	}
	fmt.Println()

	// Demonstrate NoREC concept
	fmt.Println("2. NoREC (Non-Optimizing Reference Engine) Concept")
	fmt.Println("   Compares optimized vs unoptimized query execution")
	fmt.Println()

	query := "SELECT * FROM users WHERE age > 30 ORDER BY age"
	
	// Execute with different optimization levels
	rows1, _ := db.Query(query)
	var count1 int
	for rows1.Next() {
		count1++
	}
	rows1.Close()
	
	fmt.Printf("   Query: %s\n", query)
	fmt.Printf("   Result count: %d\n", count1)
	fmt.Println("   ✓ NoREC-style validation would compare results")
	fmt.Println()

	// Demonstrate PQS concept
	fmt.Println("3. PQS (Pivoted Query Synthesis) Concept")
	fmt.Println("   Verifies that specific rows are returned by generated queries")
	fmt.Println()

	// Get a pivot row
	var pivotID int64
	var pivotAge int
	db.QueryRow("SELECT id, age FROM users LIMIT 1").Scan(&pivotID, &pivotAge)
	
	// Query that should return the pivot row
	pivotQuery := fmt.Sprintf("SELECT id, age FROM users WHERE id = %d AND age = %d", pivotID, pivotAge)
	var foundID int64
	var foundAge int
	err = db.QueryRow(pivotQuery).Scan(&foundID, &foundAge)
	
	fmt.Printf("   Pivot row: id=%d, age=%d\n", pivotID, pivotAge)
	fmt.Printf("   Query: %s\n", pivotQuery)
	
	if err == nil && foundID == pivotID && foundAge == pivotAge {
		fmt.Println("   ✓ PQS check passed - pivot row found!")
	} else {
		fmt.Println("   ✗ PQS check failed - pivot row missing!")
	}
	fmt.Println()

	// Demonstrate QPG concept
	fmt.Println("4. QPG (Query Plan Guidance) Concept")
	fmt.Println("   Tracks unique query plans to guide test generation")
	fmt.Println()

	queries := []string{
		"SELECT * FROM users WHERE age > 30",
		"SELECT * FROM users WHERE age > 30 ORDER BY age",
		"SELECT * FROM users WHERE age > 30 LIMIT 10",
	}

	seenPlans := make(map[string]bool)
	
	for _, q := range queries {
		planQuery := fmt.Sprintf("EXPLAIN QUERY PLAN %s", q)
		rows, err := db.Query(planQuery)
		if err == nil {
			var plan string
			for rows.Next() {
				var id, parent, notused int
				var detail string
				rows.Scan(&id, &parent, &notused, &detail)
				plan += detail + ";"
			}
			rows.Close()
			seenPlans[plan] = true
		}
	}

	fmt.Printf("   Tested %d queries\n", len(queries))
	fmt.Printf("   Found %d unique query plans\n", len(seenPlans))
	fmt.Println("   ✓ QPG helps maximize query plan coverage")
	fmt.Println()

	fmt.Println("Summary")
	fmt.Println("=======")
	fmt.Println("SQLancer-inspired oracles provide powerful testing techniques:")
	fmt.Println("  - TLP: Detects logic bugs via ternary partitioning")
	fmt.Println("  - NoREC: Finds optimizer bugs by comparing execution modes")
	fmt.Println("  - PQS: Validates row-level query correctness")
	fmt.Println("  - QPG: Guides fuzzing to maximize code coverage")
}
