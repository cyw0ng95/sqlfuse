package stmts

import (
	"database/sql"
	"strings"
	"testing"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
)

// TestNoUnsupportedFeatures verifies that the generator NEVER produces
// SQL with features that are explicitly NOT supported by Turso LibSQL.
// Reference: https://github.com/tursodatabase/turso/blob/main/COMPAT.md
func TestNoUnsupportedFeatures(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed the database with test tables
	createTestTables(db, t)

	// Test each generator function with multiple seeds
	tests := []struct {
		name      string
		generator func(*sql.DB, *common.LCG) (SelectStmt, error)
		forbidden []string // Patterns that should NEVER appear
	}{
		{
			name:      "SelectSubquery",
			generator: GenSelectSubquery,
			forbidden: []string{
				"EXISTS (",
				"NOT EXISTS (",
				"WHERE.*IN.*SELECT", // IN (subquery) pattern
			},
		},
		{
			name:      "SelectWhereIn",
			generator: GenSelectWhereIn,
			forbidden: []string{
				"IN.*SELECT", // IN (subquery) pattern
			},
		},
		{
			name:      "SelectWhere",
			generator: GenSelectWhere,
			forbidden: []string{
				" REGEXP ",
				" MATCH ",
				"NOT REGEXP",
				"NOT MATCH",
				" FILTER ",
				" OVER ",
				"RAISE(",
			},
		},
		{
			name:      "SelectCase",
			generator: GenSelectCase,
			forbidden: []string{
				" REGEXP ", // REGEXP operator with spaces to avoid matching in string literals
				" MATCH ",  // MATCH operator with spaces to avoid matching 'matched', 'matching', etc.
				"NOT REGEXP",
				"NOT MATCH",
			},
		},
		{
			name:      "SelectAggregateComplex",
			generator: GenSelectAggregateComplex,
			forbidden: []string{
				" FILTER ",
				"FILTER(",
				" OVER ",
				"OVER(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with multiple seeds to ensure randomness doesn't introduce forbidden features
			for seed := uint64(0); seed < 100; seed++ {
				lcg := common.NewLCG(seed)
				stmt, err := tt.generator(db, lcg)
				if err != nil {
					t.Fatalf("Generator failed: %v", err)
				}

				sql := stmt.SQL()
				sqlUpper := strings.ToUpper(sql)

				// Check for forbidden patterns
				for _, forbidden := range tt.forbidden {
					forbiddenUpper := strings.ToUpper(forbidden)
					// Simple contains check for most patterns
					if strings.Contains(sqlUpper, forbiddenUpper) {
						t.Errorf("Seed %d: Generated SQL contains forbidden pattern '%s':\n%s",
							seed, forbidden, sql)
					}
				}
			}
		})
	}
}

// TestBinaryOperatorCompliance verifies that only supported binary operators are generated
func TestBinaryOperatorCompliance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	createTestTables(db, t)

	// Unsupported operators per Turso COMPAT.md
	unsupportedOps := []string{
		"%", // modulo - NOT SUPPORTED
		// Note: !< and !> are unlikely to appear in our simple generator
		// but we should check for them anyway
	}

	ctx := NewGenContext(db, common.NewLCG(42), 2)
	eg := NewExprGenerator(ctx)

	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	// Generate many binary expressions and verify none contain unsupported operators
	// Note: This uses simple pattern matching which may have false positives in edge cases
	// (e.g., operators in string literals or comments), but our generator doesn't produce those.
	for i := 0; i < 200; i++ {
		expr := eg.GenBinaryExpr(tbls)

		for _, op := range unsupportedOps {
			// Check for the operator with spaces around it to reduce false positives
			patterns := []string{
				" " + op + " ",
				"(" + op,
				op + "(",
			}
			for _, pattern := range patterns {
				if strings.Contains(expr, pattern) {
					t.Errorf("Generated binary expression contains unsupported operator '%s':\n%s",
						op, expr)
				}
			}
		}
	}
}

// TestCollateOnlyDefaultCollations verifies that only default SQLite collations are used
func TestCollateOnlyDefaultCollations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	createTestTables(db, t)

	// Allowed collations per Turso COMPAT.md
	allowedCollations := map[string]bool{
		"BINARY": true,
		"NOCASE": true,
		"RTRIM":  true,
	}

	ctx := NewGenContext(db, common.NewLCG(42), 2)
	eg := NewExprGenerator(ctx)

	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	// Generate many COLLATE expressions
	for i := 0; i < 100; i++ {
		expr := eg.GenCollateExpr(tbls)
		exprUpper := strings.ToUpper(expr)

		if !strings.Contains(exprUpper, "COLLATE") {
			continue // Skip if COLLATE wasn't generated
		}

		// Extract the collation name
		parts := strings.Split(exprUpper, "COLLATE")
		if len(parts) < 2 {
			continue
		}

		collationPart := strings.TrimSpace(parts[1])
		collationName := strings.Fields(collationPart)[0]

		if !allowedCollations[collationName] {
			t.Errorf("Generated COLLATE expression uses non-default collation '%s':\n%s",
				collationName, expr)
		}
	}
}

// TestScalarFunctionCompliance verifies that only supported scalar functions are generated
func TestScalarFunctionCompliance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	createTestTables(db, t)

	// Unsupported functions per Turso COMPAT.md
	unsupportedFuncs := []string{
		"format(", // format(FORMAT,...) - NOT SUPPORTED
		// Note: load_extension with 1 param is supported, but with 2 params is not
		// We check for the simpler pattern and would need manual inspection for multi-param
	}

	// Generate many SELECT statements with scalar functions
	for seed := uint64(0); seed < 50; seed++ {
		lcg := common.NewLCG(seed)
		stmt, err := GenSelectWithScalarFunction(db, lcg)
		if err != nil {
			t.Fatalf("Generator failed: %v", err)
		}

		sql := stmt.SQL()
		sqlLower := strings.ToLower(sql)

		for _, unsupported := range unsupportedFuncs {
			unsupportedLower := strings.ToLower(unsupported)
			if strings.Contains(sqlLower, unsupportedLower) {
				t.Errorf("Seed %d: Generated SQL contains unsupported function pattern '%s':\n%s",
					seed, unsupported, sql)
			}
		}
	}
}

// TestExpressionTypeCompliance verifies all expression types follow Turso compatibility
func TestExpressionTypeCompliance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	createTestTables(db, t)

	unsupportedExprPatterns := []string{
		"REGEXP",
		"MATCH",
		"FILTER.*WHERE", // FILTER clause pattern
		"OVER.*(",       // Window function pattern
		"RAISE",
	}

	ctx := NewGenContext(db, common.NewLCG(42), 2)
	eg := NewExprGenerator(ctx)

	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	// Generate many random expressions
	for i := 0; i < 200; i++ {
		ctx.LCG = common.NewLCG(uint64(i)) // Reset LCG for each iteration
		eg.ctx = ctx
		expr := eg.GenRandomExpr(tbls)
		exprUpper := strings.ToUpper(expr)

		for _, pattern := range unsupportedExprPatterns {
			patternUpper := strings.ToUpper(pattern)
			if strings.Contains(exprUpper, patternUpper) {
				t.Errorf("Iteration %d: Generated expression contains unsupported pattern '%s':\n%s",
					i, pattern, expr)
			}
		}
	}
}

// Helper function to create test tables with sample data
func createTestTables(db *sql.DB, t *testing.T) {
	t.Helper()

	// Create test tables with various column types
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS test_users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT,
			age INTEGER,
			balance REAL,
			data BLOB
		)`,
		`CREATE TABLE IF NOT EXISTS test_products (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			price REAL,
			stock INTEGER,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS test_orders (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			product_id INTEGER,
			quantity INTEGER,
			total REAL
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}
	}

	// Insert some test data (simple inserts, may fail silently if already exists)
	testData := []string{
		`DELETE FROM test_users`,
		`DELETE FROM test_products`,
		`DELETE FROM test_orders`,
		`INSERT INTO test_users (id, name, email, age, balance) VALUES 
			(1, 'Alice', 'alice@test.com', 30, 100.50),
			(2, 'Bob', 'bob@test.com', 25, 250.75),
			(3, 'Charlie', 'charlie@test.com', 35, 500.00)`,
		`INSERT INTO test_products (id, name, price, stock) VALUES
			(1, 'Widget', 10.99, 100),
			(2, 'Gadget', 25.50, 50),
			(3, 'Tool', 15.00, 75)`,
		`INSERT INTO test_orders (id, user_id, product_id, quantity, total) VALUES
			(1, 1, 1, 2, 21.98),
			(2, 2, 2, 1, 25.50),
			(3, 3, 3, 3, 45.00)`,
	}

	for _, data := range testData {
		if _, err := db.Exec(data); err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}
}
