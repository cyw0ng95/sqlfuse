package stmts

import (
	"strings"
	"testing"

	"sqlfuse/internal/common"
)

// TestGenBeginTransaction tests BEGIN TRANSACTION statement generation
func TestGenBeginTransaction(t *testing.T) {
	lcg := common.NewLCG(42)

	for i := 0; i < testIterations; i++ {
		stmt := GenBeginTransaction(lcg)
		sql := stmt.SQL()

		if sql == "" {
			t.Error("GenBeginTransaction returned empty SQL")
		}

		if stmt.Type() != "transaction" {
			t.Errorf("Expected type 'transaction', got '%s'", stmt.Type())
		}

		// Verify it's a valid BEGIN variant
		validPrefixes := []string{"BEGIN", "BEGIN TRANSACTION", "BEGIN DEFERRED", "BEGIN IMMEDIATE", "BEGIN EXCLUSIVE"}
		hasValidPrefix := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(sql, prefix) {
				hasValidPrefix = true
				break
			}
		}
		if !hasValidPrefix {
			t.Errorf("Invalid BEGIN statement: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid BEGIN SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenCommitTransaction tests COMMIT TRANSACTION statement generation
func TestGenCommitTransaction(t *testing.T) {
	lcg := common.NewLCG(123)

	for i := 0; i < testIterations; i++ {
		stmt := GenCommitTransaction(lcg)
		sql := stmt.SQL()

		if sql == "" {
			t.Error("GenCommitTransaction returned empty SQL")
		}

		if stmt.Type() != "transaction" {
			t.Errorf("Expected type 'transaction', got '%s'", stmt.Type())
		}

		// Verify it's a valid COMMIT/END variant
		validPrefixes := []string{"COMMIT", "END"}
		hasValidPrefix := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(sql, prefix) {
				hasValidPrefix = true
				break
			}
		}
		if !hasValidPrefix {
			t.Errorf("Invalid COMMIT statement: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid COMMIT SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenRollbackTransaction tests ROLLBACK TRANSACTION statement generation
func TestGenRollbackTransaction(t *testing.T) {
	lcg := common.NewLCG(456)

	for i := 0; i < testIterations; i++ {
		stmt := GenRollbackTransaction(lcg)
		sql := stmt.SQL()

		if sql == "" {
			t.Error("GenRollbackTransaction returned empty SQL")
		}

		if stmt.Type() != "transaction" {
			t.Errorf("Expected type 'transaction', got '%s'", stmt.Type())
		}

		// Verify it's a ROLLBACK statement
		if !strings.HasPrefix(sql, "ROLLBACK") {
			t.Errorf("Invalid ROLLBACK statement: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid ROLLBACK SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenExplain tests EXPLAIN statement generation
func TestGenExplain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(789)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenExplain(db, lcg)
		if err != nil {
			t.Fatalf("GenExplain failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenExplain returned empty SQL")
		}

		if stmt.Type() != "explain" {
			t.Errorf("Expected type 'explain', got '%s'", stmt.Type())
		}

		// Verify it starts with EXPLAIN
		if !strings.HasPrefix(sql, "EXPLAIN") {
			t.Errorf("EXPLAIN statement doesn't start with EXPLAIN: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid EXPLAIN SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenExplainQueryPlan tests EXPLAIN QUERY PLAN statement generation
func TestGenExplainQueryPlan(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(101112)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenExplainQueryPlan(db, lcg)
		if err != nil {
			t.Fatalf("GenExplainQueryPlan failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenExplainQueryPlan returned empty SQL")
		}

		// Verify it starts with EXPLAIN QUERY PLAN
		if !strings.HasPrefix(sql, "EXPLAIN QUERY PLAN") {
			t.Errorf("EXPLAIN QUERY PLAN statement doesn't have correct prefix: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid EXPLAIN QUERY PLAN SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenCreateIndex tests CREATE INDEX statement generation
func TestGenCreateIndex(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(131415)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenCreateIndex(db, lcg)
		if err != nil {
			t.Fatalf("GenCreateIndex failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenCreateIndex returned empty SQL")
		}

		if stmt.Type() != "create_index" {
			t.Errorf("Expected type 'create_index', got '%s'", stmt.Type())
		}

		// Verify it contains CREATE INDEX
		if !strings.Contains(sql, "CREATE") || !strings.Contains(sql, "INDEX") {
			t.Errorf("Invalid CREATE INDEX statement: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid CREATE INDEX SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenDropIndex tests DROP INDEX statement generation
func TestGenDropIndex(t *testing.T) {
	lcg := common.NewLCG(161718)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenDropIndex(lcg)
		if err != nil {
			t.Fatalf("GenDropIndex failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenDropIndex returned empty SQL")
		}

		if stmt.Type() != "drop_index" {
			t.Errorf("Expected type 'drop_index', got '%s'", stmt.Type())
		}

		// Verify it's a DROP INDEX statement
		if !strings.HasPrefix(sql, "DROP INDEX") {
			t.Errorf("Invalid DROP INDEX statement: %s", sql)
		}

		// Should always have IF EXISTS for safety
		if !strings.Contains(sql, "IF EXISTS") {
			t.Errorf("DROP INDEX should use IF EXISTS: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid DROP INDEX SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenAttachDatabase tests ATTACH DATABASE statement generation
func TestGenAttachDatabase(t *testing.T) {
	lcg := common.NewLCG(192021)

	for i := 0; i < testIterations; i++ {
		stmt := GenAttachDatabase(lcg)
		sql := stmt.SQL()

		if sql == "" {
			t.Error("GenAttachDatabase returned empty SQL")
		}

		if stmt.Type() != "attach" {
			t.Errorf("Expected type 'attach', got '%s'", stmt.Type())
		}

		// Verify it's an ATTACH DATABASE statement
		if !strings.HasPrefix(sql, "ATTACH DATABASE") {
			t.Errorf("Invalid ATTACH DATABASE statement: %s", sql)
		}

		// Should contain AS clause
		if !strings.Contains(sql, " AS ") {
			t.Errorf("ATTACH DATABASE should have AS clause: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid ATTACH DATABASE SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenDetachDatabase tests DETACH DATABASE statement generation
func TestGenDetachDatabase(t *testing.T) {
	lcg := common.NewLCG(222324)

	for i := 0; i < testIterations; i++ {
		stmt := GenDetachDatabase(lcg)
		sql := stmt.SQL()

		if sql == "" {
			t.Error("GenDetachDatabase returned empty SQL")
		}

		if stmt.Type() != "detach" {
			t.Errorf("Expected type 'detach', got '%s'", stmt.Type())
		}

		// Verify it's a DETACH statement
		if !strings.HasPrefix(sql, "DETACH") {
			t.Errorf("Invalid DETACH statement: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid DETACH SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenCreateVirtualTable tests CREATE VIRTUAL TABLE statement generation
func TestGenCreateVirtualTable(t *testing.T) {
	lcg := common.NewLCG(252627)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenCreateVirtualTable(lcg)
		if err != nil {
			t.Fatalf("GenCreateVirtualTable failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenCreateVirtualTable returned empty SQL")
		}

		if stmt.Type() != "create_virtual_table" {
			t.Errorf("Expected type 'create_virtual_table', got '%s'", stmt.Type())
		}

		// Verify it contains CREATE VIRTUAL TABLE
		if !strings.Contains(sql, "CREATE VIRTUAL TABLE") {
			t.Errorf("Invalid CREATE VIRTUAL TABLE statement: %s", sql)
		}

		// Should use USING clause
		if !strings.Contains(sql, " USING ") {
			t.Errorf("CREATE VIRTUAL TABLE should have USING clause: %s", sql)
		}

		// Validate SQL syntax
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid CREATE VIRTUAL TABLE SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenInsertOrReplace tests INSERT OR REPLACE statement generation
func TestGenInsertOrReplace(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(1000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenInsertOrReplace(db, lcg)
		if err != nil {
			t.Fatalf("GenInsertOrReplace failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenInsertOrReplace returned empty SQL")
		}

		if stmt.Type() != "insert" {
			t.Errorf("Expected type 'insert', got '%s'", stmt.Type())
		}

		// Verify it contains INSERT OR REPLACE
		if !strings.Contains(sql, "INSERT OR REPLACE") {
			t.Errorf("Statement should contain INSERT OR REPLACE: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid INSERT OR REPLACE SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}

// TestGenInsertOrIgnore tests INSERT OR IGNORE statement generation
func TestGenInsertOrIgnore(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(2000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenInsertOrIgnore(db, lcg)
		if err != nil {
			t.Fatalf("GenInsertOrIgnore failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenInsertOrIgnore returned empty SQL")
		}

		if stmt.Type() != "insert" {
			t.Errorf("Expected type 'insert', got '%s'", stmt.Type())
		}

		// Verify it contains INSERT OR IGNORE
		if !strings.Contains(sql, "INSERT OR IGNORE") {
			t.Errorf("Statement should contain INSERT OR IGNORE: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid INSERT OR IGNORE SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}

// TestGenInsertOrAbort tests INSERT OR ABORT statement generation
func TestGenInsertOrAbort(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(3000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenInsertOrAbort(db, lcg)
		if err != nil {
			t.Fatalf("GenInsertOrAbort failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenInsertOrAbort returned empty SQL")
		}

		if stmt.Type() != "insert" {
			t.Errorf("Expected type 'insert', got '%s'", stmt.Type())
		}

		// Verify it contains INSERT OR ABORT
		if !strings.Contains(sql, "INSERT OR ABORT") {
			t.Errorf("Statement should contain INSERT OR ABORT: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid INSERT OR ABORT SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}

// TestGenInsertOrRollback tests INSERT OR ROLLBACK statement generation
func TestGenInsertOrRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(4000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenInsertOrRollback(db, lcg)
		if err != nil {
			t.Fatalf("GenInsertOrRollback failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenInsertOrRollback returned empty SQL")
		}

		if stmt.Type() != "insert" {
			t.Errorf("Expected type 'insert', got '%s'", stmt.Type())
		}

		// Verify it contains INSERT OR ROLLBACK
		if !strings.Contains(sql, "INSERT OR ROLLBACK") {
			t.Errorf("Statement should contain INSERT OR ROLLBACK: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid INSERT OR ROLLBACK SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}

// TestGenInsertOrFail tests INSERT OR FAIL statement generation
func TestGenInsertOrFail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(5000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenInsertOrFail(db, lcg)
		if err != nil {
			t.Fatalf("GenInsertOrFail failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenInsertOrFail returned empty SQL")
		}

		if stmt.Type() != "insert" {
			t.Errorf("Expected type 'insert', got '%s'", stmt.Type())
		}

		// Verify it contains INSERT OR FAIL
		if !strings.Contains(sql, "INSERT OR FAIL") {
			t.Errorf("Statement should contain INSERT OR FAIL: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid INSERT OR FAIL SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}

// TestGenSelectWithCTE tests SELECT statement with Common Table Expression
func TestGenSelectWithCTE(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(6000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithCTE(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithCTE failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithCTE returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		// Verify it contains WITH clause
		if !strings.Contains(sql, "WITH ") {
			t.Errorf("Statement should contain WITH clause: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WITH CTE SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}

// TestGenSelectWithMultipleCTE tests SELECT statement with multiple CTEs
func TestGenSelectWithMultipleCTE(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(7000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithMultipleCTE(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithMultipleCTE failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithMultipleCTE returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		// Verify it contains WITH clause
		if !strings.Contains(sql, "WITH ") {
			t.Errorf("Statement should contain WITH clause: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WITH MULTIPLE CTE SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}

// TestGenSelectWithRecursiveCTE tests SELECT statement with recursive CTE
func TestGenSelectWithRecursiveCTE(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(8000)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenSelectWithRecursiveCTE(db, lcg)
		if err != nil {
			t.Fatalf("GenSelectWithRecursiveCTE failed on iteration %d: %v", i, err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSelectWithRecursiveCTE returned empty SQL")
		}

		if stmt.Type() != "select" {
			t.Errorf("Expected type 'select', got '%s'", stmt.Type())
		}

		// Verify it contains WITH RECURSIVE
		if !strings.Contains(sql, "WITH RECURSIVE") {
			t.Errorf("Statement should contain WITH RECURSIVE: %s", sql)
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid SELECT WITH RECURSIVE CTE SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
		}
	}
}
