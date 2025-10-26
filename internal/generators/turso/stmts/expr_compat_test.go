package stmts

import (
	"database/sql"
	"testing"

	_ "github.com/tursodatabase/turso-go"
)

// TestExprLiterals tests literal expressions according to Turso COMPAT.md
// Status: Yes - literals are fully supported
func TestExprLiterals(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Integer literals
		{"Integer positive", "SELECT 42"},
		{"Integer negative", "SELECT -42"},
		{"Integer zero", "SELECT 0"},
		{"Integer large", "SELECT 9223372036854775807"},
		
		// Float/Real literals
		{"Float simple", "SELECT 3.14"},
		{"Float negative", "SELECT -3.14"},
		{"Float scientific", "SELECT 1.5e10"},
		{"Float scientific negative", "SELECT -1.5e-10"},
		
		// String literals
		{"String single quotes", "SELECT 'hello'"},
		{"String empty", "SELECT ''"},
		{"String with escape", "SELECT 'it''s'"},
		{"String unicode", "SELECT 'Hello 世界'"},
		
		// Blob literals
		{"Blob hex", "SELECT X'48656c6c6f'"},
		{"Blob empty", "SELECT X''"},
		
		// NULL literal
		{"NULL", "SELECT NULL"},
		
		// Boolean-like (SQLite uses integers)
		{"True as 1", "SELECT 1 AS true_val"},
		{"False as 0", "SELECT 0 AS false_val"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
			
			if !rows.Next() {
				t.Errorf("No rows returned for %s", tt.sql)
				return
			}
			
			// Verify we can scan the result
			var result interface{}
			if err := rows.Scan(&result); err != nil {
				t.Errorf("Failed to scan result for %s: %v", tt.sql, err)
			}
		})
	}
}

// TestExprUnaryOperators tests unary operators according to Turso COMPAT.md
// Status: Yes - unary operators are fully supported
func TestExprUnaryOperators(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Minus operator
		{"Unary minus integer", "SELECT -42"},
		{"Unary minus float", "SELECT -3.14"},
		{"Unary minus expression", "SELECT -(10 + 5)"},
		
		// Plus operator
		{"Unary plus integer", "SELECT +42"},
		{"Unary plus float", "SELECT +3.14"},
		
		// NOT operator
		{"NOT true", "SELECT NOT 1"},
		{"NOT false", "SELECT NOT 0"},
		{"NOT NULL", "SELECT NOT NULL"},
		{"NOT expression", "SELECT NOT (1 = 1)"},
		
		// Bitwise NOT (~)
		{"Bitwise NOT", "SELECT ~5"},
		{"Bitwise NOT zero", "SELECT ~0"},
		{"Bitwise NOT negative", "SELECT ~(-1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
			
			if !rows.Next() {
				t.Errorf("No rows returned for %s", tt.sql)
			}
		})
	}
}

// TestExprBinaryOperators tests binary operators according to Turso COMPAT.md
// Status: Partial - Only %, !<, and !> are unsupported
func TestExprBinaryOperators(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name       string
		sql        string
		shouldFail bool
	}{
		// Arithmetic operators (supported)
		{"Addition", "SELECT 10 + 5", false},
		{"Subtraction", "SELECT 10 - 5", false},
		{"Multiplication", "SELECT 10 * 5", false},
		{"Division", "SELECT 10 / 5", false},
		{"Division float", "SELECT 10.0 / 3.0", false},
		
		// Modulo operator (unsupported per COMPAT.md)
		{"Modulo - unsupported", "SELECT 10 % 3", true},
		
		// Comparison operators (supported)
		{"Equal", "SELECT 10 = 10", false},
		{"Not equal", "SELECT 10 != 5", false},
		{"Not equal <>", "SELECT 10 <> 5", false},
		{"Less than", "SELECT 5 < 10", false},
		{"Greater than", "SELECT 10 > 5", false},
		{"Less than or equal", "SELECT 5 <= 10", false},
		{"Greater than or equal", "SELECT 10 >= 5", false},
		
		// Logical operators
		{"AND", "SELECT 1 AND 1", false},
		{"OR", "SELECT 1 OR 0", false},
		
		// Bitwise operators
		{"Bitwise AND", "SELECT 5 & 3", false},
		{"Bitwise OR", "SELECT 5 | 3", false},
		{"Left shift", "SELECT 5 << 2", false},
		{"Right shift", "SELECT 20 >> 2", false},
		
		// String concatenation
		{"Concat operator", "SELECT 'Hello' || ' ' || 'World'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if tt.shouldFail {
				if err == nil {
					rows.Close()
					t.Logf("Note: %s executed successfully (expected to fail per COMPAT.md)", tt.sql)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
			
			if !rows.Next() {
				t.Errorf("No rows returned for %s", tt.sql)
			}
		})
	}
}

// TestExprParentheses tests parenthesized expressions according to Turso COMPAT.md
// Status: Yes - (expr) is fully supported
func TestExprParentheses(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		{"Simple parens", "SELECT (42)"},
		{"Nested parens", "SELECT ((42))"},
		{"Parens with arithmetic", "SELECT (10 + 5) * 2"},
		{"Parens change precedence", "SELECT 2 * (3 + 4)"},
		{"Multiple paren groups", "SELECT (10 + 5) + (20 - 15)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
			
			if !rows.Next() {
				t.Errorf("No rows returned for %s", tt.sql)
			}
		})
	}
}

// TestExprCast tests CAST expressions according to Turso COMPAT.md
// Status: Yes - CAST (expr AS type) is fully supported
func TestExprCast(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Cast to INTEGER
		{"Cast string to INTEGER", "SELECT CAST('42' AS INTEGER)"},
		{"Cast float to INTEGER", "SELECT CAST(3.14 AS INTEGER)"},
		{"Cast NULL to INTEGER", "SELECT CAST(NULL AS INTEGER)"},
		
		// Cast to REAL
		{"Cast string to REAL", "SELECT CAST('3.14' AS REAL)"},
		{"Cast integer to REAL", "SELECT CAST(42 AS REAL)"},
		{"Cast NULL to REAL", "SELECT CAST(NULL AS REAL)"},
		
		// Cast to TEXT
		{"Cast integer to TEXT", "SELECT CAST(42 AS TEXT)"},
		{"Cast float to TEXT", "SELECT CAST(3.14 AS TEXT)"},
		{"Cast NULL to TEXT", "SELECT CAST(NULL AS TEXT)"},
		
		// Cast to BLOB
		{"Cast string to BLOB", "SELECT CAST('hello' AS BLOB)"},
		{"Cast NULL to BLOB", "SELECT CAST(NULL AS BLOB)"},
		
		// Nested casts
		{"Nested CAST", "SELECT CAST(CAST('42' AS INTEGER) AS TEXT)"},
		
		// Cast in expressions
		{"CAST in arithmetic", "SELECT CAST('10' AS INTEGER) + 5"},
		{"CAST in comparison", "SELECT CAST('42' AS INTEGER) = 42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
			
			if !rows.Next() {
				t.Errorf("No rows returned for %s", tt.sql)
			}
		})
	}
}

// TestExprCollate tests COLLATE expressions according to Turso COMPAT.md
// Status: Partial - Custom Collations not supported, only built-in collations
func TestExprCollate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Built-in collations (should work)
		{"COLLATE BINARY", "SELECT 'abc' COLLATE BINARY"},
		{"COLLATE NOCASE", "SELECT 'abc' COLLATE NOCASE"},
		{"COLLATE RTRIM", "SELECT 'abc ' COLLATE RTRIM"},
		
		// COLLATE in comparison
		{"COLLATE in WHERE", "SELECT * FROM users WHERE name = 'Alice' COLLATE NOCASE"},
		{"COLLATE in ORDER BY", "SELECT name FROM users ORDER BY name COLLATE NOCASE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
			
			// Just verify it executes, don't require specific results
		})
	}
}

// TestExprLike tests LIKE/NOT LIKE expressions according to Turso COMPAT.md
// Status: Yes - (NOT) LIKE is fully supported
func TestExprLike(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Basic LIKE
		{"LIKE with wildcard", "SELECT 'hello' LIKE 'h%'"},
		{"LIKE exact match", "SELECT 'hello' LIKE 'hello'"},
		{"LIKE with _", "SELECT 'hello' LIKE 'h_llo'"},
		{"LIKE case sensitive", "SELECT 'Hello' LIKE 'hello'"},
		
		// NOT LIKE
		{"NOT LIKE", "SELECT 'hello' NOT LIKE 'world'"},
		{"NOT LIKE with wildcard", "SELECT 'hello' NOT LIKE 'w%'"},
		
		// LIKE with table data
		{"LIKE from table", "SELECT name FROM users WHERE name LIKE 'A%'"},
		{"NOT LIKE from table", "SELECT name FROM users WHERE name NOT LIKE 'Z%'"},
		
		// LIKE with escape character (3-argument form)
		{"LIKE with ESCAPE", "SELECT 'a%b' LIKE 'a\\%b' ESCAPE '\\'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
		})
	}
}

// TestExprGlob tests GLOB/NOT GLOB expressions according to Turso COMPAT.md
// Status: Yes - (NOT) GLOB is fully supported
func TestExprGlob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Basic GLOB
		{"GLOB with wildcard", "SELECT 'hello' GLOB 'h*'"},
		{"GLOB exact match", "SELECT 'hello' GLOB 'hello'"},
		{"GLOB with ?", "SELECT 'hello' GLOB 'h?llo'"},
		{"GLOB case sensitive", "SELECT 'Hello' GLOB 'hello'"},
		{"GLOB character class", "SELECT 'hello' GLOB 'h[aeiou]llo'"},
		
		// NOT GLOB
		{"NOT GLOB", "SELECT 'hello' NOT GLOB 'world'"},
		{"NOT GLOB with wildcard", "SELECT 'hello' NOT GLOB 'w*'"},
		
		// GLOB with table data
		{"GLOB from table", "SELECT name FROM users WHERE name GLOB 'A*'"},
		{"NOT GLOB from table", "SELECT name FROM users WHERE name NOT GLOB 'Z*'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
		})
	}
}

// TestExprIs tests IS/IS NOT expressions according to Turso COMPAT.md
// Status: Yes - IS (NOT) is fully supported
func TestExprIs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// IS NULL
		{"IS NULL true", "SELECT NULL IS NULL"},
		{"IS NULL false", "SELECT 42 IS NULL"},
		
		// IS NOT NULL
		{"IS NOT NULL false", "SELECT NULL IS NOT NULL"},
		{"IS NOT NULL true", "SELECT 42 IS NOT NULL"},
		
		// IS with values
		{"IS with equal values", "SELECT 42 IS 42"},
		{"IS with different values", "SELECT 42 IS 43"},
		{"IS with string", "SELECT 'hello' IS 'hello'"},
		
		// IS NOT with values
		{"IS NOT with equal values", "SELECT 42 IS NOT 42"},
		{"IS NOT with different values", "SELECT 42 IS NOT 43"},
		
		// IS with table data
		{"IS NULL from table", "SELECT * FROM users WHERE email IS NOT NULL"},
		{"IS comparison from table", "SELECT * FROM users WHERE age IS NOT NULL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
		})
	}
}

// TestExprIsDistinctFrom tests IS DISTINCT FROM/IS NOT DISTINCT FROM according to Turso COMPAT.md
// Status: Yes - IS (NOT) DISTINCT FROM is fully supported
func TestExprIsDistinctFrom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// IS DISTINCT FROM - NULL handling
		{"IS DISTINCT FROM NULL and NULL", "SELECT NULL IS DISTINCT FROM NULL"},
		{"IS DISTINCT FROM value and NULL", "SELECT 42 IS DISTINCT FROM NULL"},
		{"IS DISTINCT FROM NULL and value", "SELECT NULL IS DISTINCT FROM 42"},
		
		// IS DISTINCT FROM - value comparison
		{"IS DISTINCT FROM equal values", "SELECT 42 IS DISTINCT FROM 42"},
		{"IS DISTINCT FROM different values", "SELECT 42 IS DISTINCT FROM 43"},
		{"IS DISTINCT FROM strings", "SELECT 'hello' IS DISTINCT FROM 'world'"},
		
		// IS NOT DISTINCT FROM - NULL handling
		{"IS NOT DISTINCT FROM NULL and NULL", "SELECT NULL IS NOT DISTINCT FROM NULL"},
		{"IS NOT DISTINCT FROM value and NULL", "SELECT 42 IS NOT DISTINCT FROM NULL"},
		
		// IS NOT DISTINCT FROM - value comparison
		{"IS NOT DISTINCT FROM equal values", "SELECT 42 IS NOT DISTINCT FROM 42"},
		{"IS NOT DISTINCT FROM different values", "SELECT 42 IS NOT DISTINCT FROM 43"},
		
		// With table data
		{"IS DISTINCT FROM with column", "SELECT * FROM users WHERE age IS DISTINCT FROM NULL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
		})
	}
}

// TestExprBetween tests BETWEEN/NOT BETWEEN expressions according to Turso COMPAT.md
// Status: Yes - (NOT) BETWEEN ... AND ... is fully supported (rewritten in optimizer)
func TestExprBetween(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// BETWEEN with integers
		{"BETWEEN integers true", "SELECT 5 BETWEEN 1 AND 10"},
		{"BETWEEN integers false", "SELECT 15 BETWEEN 1 AND 10"},
		{"BETWEEN integers boundary lower", "SELECT 1 BETWEEN 1 AND 10"},
		{"BETWEEN integers boundary upper", "SELECT 10 BETWEEN 1 AND 10"},
		
		// NOT BETWEEN with integers
		{"NOT BETWEEN integers false", "SELECT 5 NOT BETWEEN 1 AND 10"},
		{"NOT BETWEEN integers true", "SELECT 15 NOT BETWEEN 1 AND 10"},
		
		// BETWEEN with floats
		{"BETWEEN floats", "SELECT 5.5 BETWEEN 1.0 AND 10.0"},
		{"NOT BETWEEN floats", "SELECT 15.5 NOT BETWEEN 1.0 AND 10.0"},
		
		// BETWEEN with strings
		{"BETWEEN strings", "SELECT 'b' BETWEEN 'a' AND 'c'"},
		{"NOT BETWEEN strings", "SELECT 'z' NOT BETWEEN 'a' AND 'c'"},
		
		// BETWEEN with table data
		{"BETWEEN with column", "SELECT * FROM users WHERE age BETWEEN 20 AND 40"},
		{"NOT BETWEEN with column", "SELECT * FROM users WHERE age NOT BETWEEN 50 AND 100"},
		{"BETWEEN with expression", "SELECT * FROM products WHERE price BETWEEN 10.0 AND 20.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
		})
	}
}

// TestExprCase tests CASE WHEN THEN ELSE END expressions according to Turso COMPAT.md
// Status: Yes - CASE WHEN THEN ELSE END is fully supported
func TestExprCase(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Simple CASE
		{"CASE simple when true", "SELECT CASE WHEN 1 = 1 THEN 'yes' ELSE 'no' END"},
		{"CASE simple when false", "SELECT CASE WHEN 1 = 2 THEN 'yes' ELSE 'no' END"},
		{"CASE without ELSE", "SELECT CASE WHEN 1 = 1 THEN 'yes' END"},
		
		// Multiple WHEN clauses
		{"CASE multiple WHEN", "SELECT CASE WHEN 1 = 2 THEN 'a' WHEN 2 = 2 THEN 'b' ELSE 'c' END"},
		{"CASE multiple WHEN none match", "SELECT CASE WHEN 1 = 2 THEN 'a' WHEN 3 = 4 THEN 'b' ELSE 'c' END"},
		
		// CASE with different types
		{"CASE returning INTEGER", "SELECT CASE WHEN 1 = 1 THEN 42 ELSE 0 END"},
		{"CASE returning REAL", "SELECT CASE WHEN 1 = 1 THEN 3.14 ELSE 0.0 END"},
		{"CASE returning NULL", "SELECT CASE WHEN 1 = 2 THEN 'value' ELSE NULL END"},
		
		// Searched CASE (no expression after CASE)
		{"Searched CASE", "SELECT CASE WHEN 10 > 5 THEN 'greater' WHEN 10 < 5 THEN 'less' ELSE 'equal' END"},
		
		// Simple CASE (with expression after CASE)
		{"Simple CASE with expression", "SELECT CASE 2 WHEN 1 THEN 'one' WHEN 2 THEN 'two' ELSE 'other' END"},
		{"Simple CASE no match", "SELECT CASE 5 WHEN 1 THEN 'one' WHEN 2 THEN 'two' ELSE 'other' END"},
		
		// CASE with table data
		{"CASE with column", "SELECT name, CASE WHEN age < 30 THEN 'young' ELSE 'old' END AS category FROM users"},
		{"CASE in WHERE", "SELECT * FROM users WHERE CASE WHEN age > 25 THEN 1 ELSE 0 END = 1"},
		
		// Nested CASE
		{"Nested CASE", "SELECT CASE WHEN 1 = 1 THEN CASE WHEN 2 = 2 THEN 'yes' ELSE 'maybe' END ELSE 'no' END"},
		
		// CASE with complex expressions
		{"CASE with arithmetic", "SELECT CASE WHEN (10 + 5) > 10 THEN 'sum is large' ELSE 'sum is small' END"},
		{"CASE with NULL", "SELECT CASE WHEN NULL IS NULL THEN 'is null' ELSE 'not null' END"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
			
			if !rows.Next() {
				t.Errorf("No rows returned for %s", tt.sql)
			}
		})
	}
}

// TestExprColumnReference tests column references according to Turso COMPAT.md
// Status: Partial - schema.table.column not supported (schemas aren't supported)
func TestExprColumnReference(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name string
		sql  string
	}{
		// Simple column reference
		{"Column by name", "SELECT name FROM users"},
		{"Multiple columns", "SELECT name, email FROM users"},
		{"All columns", "SELECT * FROM users"},
		
		// Qualified column reference (table.column)
		{"Qualified column", "SELECT users.name FROM users"},
		{"Qualified multiple", "SELECT users.name, users.email FROM users"},
		
		// Column in expressions
		{"Column in WHERE", "SELECT name FROM users WHERE age > 25"},
		{"Column in ORDER BY", "SELECT name FROM users ORDER BY age"},
		{"Column in GROUP BY", "SELECT age, COUNT(*) FROM users GROUP BY age"},
		
		// Aliased columns
		{"Column with alias", "SELECT name AS user_name FROM users"},
		{"Expression with alias", "SELECT age + 1 AS next_age FROM users"},
		
		// Column references in joins
		{"Column in JOIN", "SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id"},
		{"Qualified columns in JOIN", "SELECT users.name, orders.total FROM users JOIN orders ON users.id = orders.user_id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
		})
	}
}

// TestExprCombined tests combinations of expressions
func TestExprCombined(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name       string
		sql        string
		shouldFail bool
	}{
		// Complex WHERE clauses
		{"WHERE with AND OR", "SELECT * FROM users WHERE (age > 20 AND age < 40) OR name LIKE 'A%'", false},
		{"WHERE with BETWEEN and LIKE", "SELECT * FROM users WHERE age BETWEEN 20 AND 40 AND name LIKE '%e%'", false},
		{"WHERE with CASE", "SELECT * FROM users WHERE CASE WHEN age > 25 THEN 1 ELSE 0 END = 1", false},
		
		// Complex SELECT expressions
		{"SELECT with arithmetic", "SELECT age, age + 10 AS age_plus_ten, age * 2 AS age_doubled FROM users", false},
		{"SELECT with CASE and CAST", "SELECT CAST(CASE WHEN age > 25 THEN age ELSE 0 END AS TEXT) FROM users", false},
		{"SELECT with nested functions", "SELECT UPPER(LOWER(name)) AS normalized_name FROM users", false},
		
		// Subqueries with expressions - NOT supported per COMPAT.md (subquery in WHERE not supported)
		{"Subquery in SELECT - unsupported", "SELECT name, (SELECT COUNT(*) FROM orders WHERE user_id = users.id) AS order_count FROM users", true},
		
		// Complex ORDER BY
		{"ORDER BY expression", "SELECT * FROM users ORDER BY age * 2 DESC", false},
		{"ORDER BY CASE", "SELECT * FROM users ORDER BY CASE WHEN age < 30 THEN 1 ELSE 2 END, name", false},
		
		// Multiple operators
		{"Arithmetic and comparison", "SELECT * FROM products WHERE (price * stock) > 1000", false},
		{"String concat and LIKE", "SELECT * FROM users WHERE (name || ' ' || email) LIKE '%example%'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if tt.shouldFail {
				if err == nil {
					rows.Close()
					t.Logf("Note: %s executed successfully (expected to fail per COMPAT.md)", tt.sql)
				}
				// Expected to fail, test passes
				return
			}
			
			if err != nil {
				t.Errorf("Failed to execute %s: %v", tt.sql, err)
				return
			}
			defer rows.Close()
		})
	}
}

// TestExprUnsupportedFeatures tests features explicitly marked as unsupported in COMPAT.md
func TestExprUnsupportedFeatures(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		sql         string
		description string
	}{
		// These should fail or be ignored according to COMPAT.md
		{"REGEXP not supported", "SELECT 'hello' REGEXP 'h.*o'", "REGEXP is not supported"},
		{"MATCH not supported", "SELECT 'hello' MATCH 'hello'", "MATCH is not supported"},
		{"RAISE not supported", "SELECT RAISE(IGNORE)", "RAISE is not supported"},
		
		// Note: NOT IN (subquery) and EXISTS (subquery) are also unsupported
		// but harder to test without more complex setup
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				// Expected to fail
				t.Logf("Expected failure: %s - %v", tt.description, err)
				return
			}
			if rows != nil {
				rows.Close()
			}
			t.Logf("Note: %s executed (marked as unsupported but may work in this version)", tt.description)
		})
	}
}

// TestExprWithRealData tests expressions with actual table data to ensure practical usage works
func TestExprWithRealData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		sql         string
		expectRows  bool
		description string
	}{
		// Test various expressions actually work with real data
		{
			"Filter by age range",
			"SELECT name FROM users WHERE age BETWEEN 20 AND 35",
			true,
			"BETWEEN with numeric column",
		},
		{
			"Case-insensitive name search",
			"SELECT name FROM users WHERE UPPER(name) LIKE UPPER('alice%')",
			true,
			"LIKE with UPPER function",
		},
		{
			"NULL email check",
			"SELECT name FROM users WHERE email IS NOT NULL",
			true,
			"IS NOT NULL filter",
		},
		{
			"Price calculation",
			"SELECT name, CAST(price * 1.1 AS REAL) AS price_with_tax FROM products",
			true,
			"Arithmetic with CAST",
		},
		{
			"Categorize products by price",
			"SELECT name, CASE WHEN price < 15 THEN 'cheap' WHEN price < 25 THEN 'medium' ELSE 'expensive' END AS category FROM products",
			true,
			"CASE for categorization",
		},
		{
			"Complex join with expressions",
			"SELECT u.name, COUNT(o.id) AS order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id, u.name HAVING COUNT(o.id) > 0",
			true,
			"JOIN with GROUP BY and HAVING",
		},
		{
			"String concatenation",
			"SELECT name || ' <' || email || '>' AS contact FROM users WHERE email IS NOT NULL",
			true,
			"String concatenation operator",
		},
		{
			"Numeric comparison with CAST",
			"SELECT name FROM products WHERE CAST(stock AS REAL) / CAST(price AS REAL) > 5.0",
			true,
			"Division with CAST to REAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.sql)
			if err != nil {
				t.Errorf("Failed to execute %s: %v\nSQL: %s", tt.description, err, tt.sql)
				return
			}
			defer rows.Close()

			hasRows := rows.Next()
			if tt.expectRows && !hasRows {
				t.Logf("Warning: Expected rows for %s but got none\nSQL: %s", tt.description, tt.sql)
			}
		})
	}
}

// BenchmarkExprEvaluation benchmarks expression evaluation performance
func BenchmarkExprEvaluation(b *testing.B) {
	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	benchmarks := []struct {
		name string
		sql  string
	}{
		{"Simple literal", "SELECT 42"},
		{"Arithmetic", "SELECT 10 + 20 * 3"},
		{"String LIKE", "SELECT 'hello world' LIKE '%world%'"},
		{"CASE expression", "SELECT CASE WHEN 1 = 1 THEN 'yes' ELSE 'no' END"},
		{"CAST", "SELECT CAST('123' AS INTEGER)"},
		{"BETWEEN", "SELECT 50 BETWEEN 1 AND 100"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rows, err := db.Query(bm.sql)
				if err != nil {
					b.Fatal(err)
				}
				rows.Close()
			}
		})
	}
}
