package stmts

import (
	"testing"

	"sqlsmith-go/internal/common"
)

// TestGenSelectWithExpressions tests SELECT with various expressions
func TestGenSelectWithExpressions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2300)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithExpressions(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithExpressions failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithExpressions returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WITH EXPRESSIONS SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Logf("Warning: Failed to execute SELECT WITH EXPRESSIONS on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWhereCast tests SELECT with CAST in WHERE clause
func TestGenSelectWhereCast(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2400)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWhereCast(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWhereCast failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWhereCast returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WHERE CAST SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Verify the SQL executes without error
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute SELECT WHERE CAST on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWhereBetween tests SELECT with BETWEEN in WHERE clause
func TestGenSelectWhereBetween(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2500)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWhereBetween(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWhereBetween failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWhereBetween returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WHERE BETWEEN SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Verify the SQL executes without error
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute SELECT WHERE BETWEEN on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWhereGlob tests SELECT with GLOB in WHERE clause
func TestGenSelectWhereGlob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2600)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWhereGlob(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWhereGlob failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWhereGlob returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WHERE GLOB SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Verify the SQL executes without error
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute SELECT WHERE GLOB on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithCollate tests SELECT with COLLATE in ORDER BY
func TestGenSelectWithCollate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2700)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithCollate(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithCollate failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithCollate returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WITH COLLATE SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Verify the SQL executes without error
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute SELECT WITH COLLATE on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithUnaryOp tests SELECT with unary operators
func TestGenSelectWithUnaryOp(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2800)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithUnaryOp(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithUnaryOp failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithUnaryOp returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WITH UNARY OP SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Verify the SQL executes without error
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute SELECT WITH UNARY OP on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenSelectWithBinaryOp tests SELECT with binary operators
func TestGenSelectWithBinaryOp(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2900)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithBinaryOp(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithBinaryOp failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithBinaryOp returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WITH BINARY OP SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}

		// Verify the SQL executes without error
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute SELECT WITH BINARY OP on iteration %d: %v\nSQL: %s", i, err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}
