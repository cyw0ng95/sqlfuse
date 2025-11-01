package main

// Example demonstrating oracle-based validation inspired by SQLRight.
// Oracles generate functionally equivalent queries and compare results
// to detect logical bugs in database systems.
//
// Note: This example uses internal packages which are normally not accessible
// from outside the module. To run this example, you would need to:
// 1. Move it to an internal directory, or
// 2. Export the necessary APIs from public packages
//
// This is a demonstration of the oracle API design - in production, these features
// would be integrated into the executors.

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"sqlfuse/internal/oracles"
)

// This example demonstrates the oracle-based validation feature inspired by SQLRight.
// Oracles generate functionally equivalent queries and compare results to detect logical bugs.

func main() {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create sample schema
	setupSchema(db)

	// Insert test data
	insertTestData(db)

	fmt.Println("Oracle-Based Validation Example")
	fmt.Println("=================================\n")

	// Example 1: NOREC Oracle - Verify COUNT(*) correctness
	fmt.Println("1. NOREC Oracle (No Empty Result Check)")
	fmt.Println("   Verifies that COUNT(*) matches actual row count\n")

	norecOracle := oracles.NewNoRecOracle()
	testQuery := "SELECT name, age FROM users WHERE age > 25"

	fmt.Printf("   Base Query: %s\n", testQuery)

	if norecOracle.IsApplicable(testQuery) {
		transformedQueries := norecOracle.TransformQuery(testQuery)
		fmt.Printf("   Transformed: %s\n", transformedQueries[0])

		result := oracles.ValidateWithOracle(db, norecOracle, testQuery)
		fmt.Printf("   Result: %s\n\n", result)

		if result == oracles.Fail {
			fmt.Println("   ⚠️  Logical bug detected!")
		} else if result == oracles.Pass {
			fmt.Println("   ✓  Query passed validation")
		}
	}

	// Example 2: TLP Oracle - Verify ternary logic partitioning
	fmt.Println("\n2. TLP Oracle (Ternary Logic Partitioning)")
	fmt.Println("   Verifies WHERE clause evaluation with TRUE/FALSE/NULL partitioning\n")

	tlpOracle := oracles.NewTLPOracle()
	testQuery2 := "SELECT name FROM users WHERE age > 25"

	fmt.Printf("   Base Query: %s\n", testQuery2)

	if tlpOracle.IsApplicable(testQuery2) {
		transformedQueries := tlpOracle.TransformQuery(testQuery2)
		if len(transformedQueries) > 0 {
			fmt.Printf("   Transformed: %s\n", transformedQueries[0])

			result := oracles.ValidateWithOracle(db, tlpOracle, testQuery2)
			fmt.Printf("   Result: %s\n\n", result)

			if result == oracles.Fail {
				fmt.Println("   ⚠️  Logical bug detected!")
			} else if result == oracles.Pass {
				fmt.Println("   ✓  Query passed validation")
			} else {
				fmt.Println("   ℹ️  Execution error (expected for simplified TLP)")
			}
		}
	}

	// Example 3: Test with NULL values
	fmt.Println("\n3. Testing with NULL values")
	fmt.Println("   Demonstrates oracle behavior with NULL data\n")

	testQuery3 := "SELECT name FROM users WHERE age IS NOT NULL"
	fmt.Printf("   Query: %s\n", testQuery3)

	result := oracles.ValidateWithOracle(db, norecOracle, testQuery3)
	fmt.Printf("   NOREC Result: %s\n", result)

	// Example 4: Multiple oracles on same query
	fmt.Println("\n4. Using multiple oracles on the same query")
	fmt.Println("   Comprehensive validation with different oracle types\n")

	allOracles := []oracles.Oracle{
		oracles.NewNoRecOracle(),
		oracles.NewTLPOracle(),
	}

	testQuery4 := "SELECT * FROM users WHERE age > 20"
	fmt.Printf("   Query: %s\n\n", testQuery4)

	for _, oracle := range allOracles {
		if oracle.IsApplicable(testQuery4) {
			result := oracles.ValidateWithOracle(db, oracle, testQuery4)
			fmt.Printf("   %s: %s\n", oracle.Name(), result)
		} else {
			fmt.Printf("   %s: Not applicable\n", oracle.Name())
		}
	}

	// Example 5: Demonstrate bug detection (simulated)
	fmt.Println("\n5. Simulated Bug Detection")
	fmt.Println("   If COUNT(*) returned wrong value, oracle would detect it\n")

	// Execute base query
	rows, _ := db.Query("SELECT name FROM users WHERE age > 25")
	var actualCount int
	for rows.Next() {
		actualCount++
	}
	rows.Close()

	// Execute count query
	countQuery := "SELECT COUNT(*) FROM (SELECT name FROM users WHERE age > 25)"
	var countResult int
	db.QueryRow(countQuery).Scan(&countResult)

	fmt.Printf("   Actual rows returned: %d\n", actualCount)
	fmt.Printf("   COUNT(*) result: %d\n", countResult)

	if actualCount == countResult {
		fmt.Println("   ✓  COUNT(*) matches actual count - No bug")
	} else {
		fmt.Println("   ⚠️  COUNT(*) mismatch - Bug detected!")
	}

	fmt.Println("\n=================================")
	fmt.Println("For more information, see docs/SQLRIGHT_INTEGRATION.md")
}

func setupSchema(db *sql.DB) {
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			age INTEGER,
			email TEXT
		);

		CREATE TABLE orders (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			amount REAL,
			status TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`

	_, err := db.Exec(schema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create schema: %v\n", err)
		os.Exit(1)
	}
}

func insertTestData(db *sql.DB) {
	data := `
		INSERT INTO users (id, name, age, email) VALUES
			(1, 'Alice', 30, 'alice@example.com'),
			(2, 'Bob', 25, 'bob@example.com'),
			(3, 'Charlie', 35, 'charlie@example.com'),
			(4, 'David', 28, 'david@example.com'),
			(5, 'Eve', NULL, 'eve@example.com');

		INSERT INTO orders (id, user_id, amount, status) VALUES
			(1, 1, 100.50, 'completed'),
			(2, 1, 250.75, 'pending'),
			(3, 2, 75.00, 'completed'),
			(4, 3, 500.00, 'completed'),
			(5, 4, 125.25, 'cancelled');
	`

	_, err := db.Exec(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to insert data: %v\n", err)
		os.Exit(1)
	}
}
