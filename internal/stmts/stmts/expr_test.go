package stmts

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"

	_ "github.com/tursodatabase/turso-go"
)

// setupTestDBForExpr creates an in-memory database with test tables for expression testing
func setupTestDBForExpr(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("turso", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	schema := []string{
		`CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT,
			value REAL,
			count INTEGER,
			description TEXT
		)`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			t.Fatalf("Failed to create test table: %v\nSQL: %s", err, stmt)
		}
	}

	// Insert test data
	testData := `INSERT INTO test_table (id, name, value, count, description) VALUES 
		(1, 'test1', 1.5, 10, 'description1'),
		(2, 'test2', 2.5, 20, 'description2')`

	if _, err := db.Exec(testData); err != nil {
		db.Close()
		t.Fatalf("Failed to insert test data: %v", err)
	}

	return db
}

// TestGenCastExpr tests CAST expression generation
func TestGenCastExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(42)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	for i := 0; i < testIterations; i++ {
		expr := eg.GenCastExpr(tbls)
		if expr == "" {
			t.Error("GenCastExpr returned empty expression")
		}

		if !strings.Contains(expr, "CAST") || !strings.Contains(expr, "AS") {
			t.Errorf("Expected CAST expression, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT %s FROM test_table;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid CAST expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute CAST expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenBetweenExpr tests BETWEEN expression generation
func TestGenBetweenExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(43)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test BETWEEN
	for i := 0; i < testIterations; i++ {
		expr := eg.GenBetweenExpr(tbls, false)
		if expr == "" {
			t.Error("GenBetweenExpr returned empty expression")
		}

		if !strings.Contains(expr, "BETWEEN") || !strings.Contains(expr, "AND") {
			t.Errorf("Expected BETWEEN expression, got: %s", expr)
		}

		if strings.Contains(expr, "NOT") {
			t.Errorf("Expected BETWEEN without NOT, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT * FROM test_table WHERE %s;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid BETWEEN expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute BETWEEN expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}

	// Test NOT BETWEEN
	for i := 0; i < testIterations; i++ {
		expr := eg.GenBetweenExpr(tbls, true)
		if expr == "" {
			t.Error("GenBetweenExpr returned empty expression")
		}

		if !strings.Contains(expr, "NOT BETWEEN") || !strings.Contains(expr, "AND") {
			t.Errorf("Expected NOT BETWEEN expression, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT * FROM test_table WHERE %s;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid NOT BETWEEN expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute NOT BETWEEN expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenGlobExpr tests GLOB expression generation
func TestGenGlobExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(44)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test GLOB
	for i := 0; i < testIterations; i++ {
		expr := eg.GenGlobExpr(tbls, false)
		if expr == "" {
			t.Error("GenGlobExpr returned empty expression")
		}

		if !strings.Contains(expr, "GLOB") {
			t.Errorf("Expected GLOB expression, got: %s", expr)
		}

		if strings.Contains(expr, "NOT GLOB") {
			t.Errorf("Expected GLOB without NOT, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT * FROM test_table WHERE %s;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid GLOB expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute GLOB expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}

	// Test NOT GLOB
	for i := 0; i < testIterations; i++ {
		expr := eg.GenGlobExpr(tbls, true)
		if expr == "" {
			t.Error("GenGlobExpr returned empty expression")
		}

		if !strings.Contains(expr, "NOT GLOB") {
			t.Errorf("Expected NOT GLOB expression, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT * FROM test_table WHERE %s;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid NOT GLOB expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute NOT GLOB expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenIsDistinctFromExpr tests IS (NOT) DISTINCT FROM expression generation
// Note: IS DISTINCT FROM is listed in Turso COMPAT.md as supported, but the SQLite ANTLR parser
// doesn't recognize it, so we skip validation for now. This function is kept for documentation.
func TestGenIsDistinctFromExpr(t *testing.T) {
	t.Skip("IS DISTINCT FROM is not supported by SQLite ANTLR parser")

	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(45)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test IS DISTINCT FROM
	for i := 0; i < testIterations; i++ {
		expr := eg.GenIsDistinctFromExpr(tbls, false)
		if expr == "" {
			t.Error("GenIsDistinctFromExpr returned empty expression")
		}

		if !strings.Contains(expr, "IS DISTINCT FROM") {
			t.Errorf("Expected IS DISTINCT FROM expression, got: %s", expr)
		}

		if strings.Contains(expr, "IS NOT DISTINCT FROM") {
			t.Errorf("Expected IS DISTINCT FROM without NOT, got: %s", expr)
		}
	}

	// Test IS NOT DISTINCT FROM
	for i := 0; i < testIterations; i++ {
		expr := eg.GenIsDistinctFromExpr(tbls, true)
		if expr == "" {
			t.Error("GenIsDistinctFromExpr returned empty expression")
		}

		if !strings.Contains(expr, "IS NOT DISTINCT FROM") {
			t.Errorf("Expected IS NOT DISTINCT FROM expression, got: %s", expr)
		}
	}
}

// TestGenCollateExpr tests COLLATE expression generation
func TestGenCollateExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(46)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	for i := 0; i < testIterations; i++ {
		expr := eg.GenCollateExpr(tbls)
		if expr == "" {
			t.Error("GenCollateExpr returned empty expression")
		}

		if !strings.Contains(expr, "COLLATE") {
			t.Errorf("Expected COLLATE expression, got: %s", expr)
		}

		// Check for valid collations
		hasValidCollation := strings.Contains(expr, "BINARY") ||
			strings.Contains(expr, "NOCASE") ||
			strings.Contains(expr, "RTRIM")
		if !hasValidCollation {
			t.Errorf("Expected valid COLLATE expression, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT %s FROM test_table;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid COLLATE expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute COLLATE expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenUnaryExpr tests unary operator expression generation
func TestGenUnaryExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(47)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	for i := 0; i < testIterations; i++ {
		expr := eg.GenUnaryExpr(tbls)
		if expr == "" {
			t.Error("GenUnaryExpr returned empty expression")
		}

		// Check for unary operators
		hasUnaryOp := strings.HasPrefix(expr, "+") ||
			strings.HasPrefix(expr, "-") ||
			strings.HasPrefix(expr, "~") ||
			strings.HasPrefix(expr, "NOT")
		if !hasUnaryOp {
			t.Errorf("Expected unary operator expression, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT %s FROM test_table;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid unary operator expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute unary operator expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenBinaryExpr tests binary operator expression generation
func TestGenBinaryExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(48)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	for i := 0; i < testIterations; i++ {
		expr := eg.GenBinaryExpr(tbls)
		if expr == "" {
			t.Error("GenBinaryExpr returned empty expression")
		}

		// Ensure unsupported operators are not used
		if strings.Contains(expr, " % ") || strings.Contains(expr, "!<") || strings.Contains(expr, "!>") {
			t.Errorf("Unsupported binary operator found in expression: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT %s FROM test_table;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid binary operator expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute binary operator expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenParenthesizedExpr tests parenthesized expression generation
func TestGenParenthesizedExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(49)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	for i := 0; i < testIterations; i++ {
		expr := eg.GenParenthesizedExpr(tbls)
		if expr == "" {
			t.Error("GenParenthesizedExpr returned empty expression")
		}

		if !strings.HasPrefix(expr, "(") || !strings.HasSuffix(expr, ")") {
			t.Errorf("Expected parenthesized expression, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT * FROM test_table WHERE %s;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid parenthesized expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute parenthesized expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenIsNullExpr tests IS (NOT) NULL expression generation
func TestGenIsNullExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(50)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test IS NULL
	for i := 0; i < testIterations; i++ {
		expr := eg.GenIsNullExpr(tbls, false)
		if expr == "" {
			t.Error("GenIsNullExpr returned empty expression")
		}

		if !strings.Contains(expr, "IS NULL") {
			t.Errorf("Expected IS NULL expression, got: %s", expr)
		}

		if strings.Contains(expr, "IS NOT NULL") {
			t.Errorf("Expected IS NULL without NOT, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT * FROM test_table WHERE %s;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid IS NULL expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute IS NULL expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}

	// Test IS NOT NULL
	for i := 0; i < testIterations; i++ {
		expr := eg.GenIsNullExpr(tbls, true)
		if expr == "" {
			t.Error("GenIsNullExpr returned empty expression")
		}

		if !strings.Contains(expr, "IS NOT NULL") {
			t.Errorf("Expected IS NOT NULL expression, got: %s", expr)
		}

		// Validate SQL syntax
		sql := fmt.Sprintf("SELECT * FROM test_table WHERE %s;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid IS NOT NULL expression SQL: %s\nErrors: %v", sql, errors)
		}

		// Try to execute
		rows, err := db.Query(sql)
		if err != nil {
			t.Errorf("Failed to execute IS NOT NULL expression: %v\nSQL: %s", err, sql)
		}
		if rows != nil {
			rows.Close()
		}
	}
}

// TestGenRandomExpr tests random expression generation
func TestGenRandomExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(51)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Generate many random expressions to test variety
	for i := 0; i < testIterations*10; i++ {
		expr := eg.GenRandomExpr(tbls)
		if expr == "" {
			t.Error("GenRandomExpr returned empty expression")
		}

		// Try to use it in a SELECT
		sql := fmt.Sprintf("SELECT %s FROM test_table;", expr)
		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid random expression SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenRegexpExpr tests REGEXP expression generation
// Note: REGEXP is not supported by Turso according to COMPAT.md
func TestGenRegexpExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(52)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test REGEXP
	for i := 0; i < testIterations; i++ {
		expr := eg.GenRegexpExpr(tbls, false)
		if expr == "" {
			t.Error("GenRegexpExpr returned empty expression")
		}

		if !strings.Contains(expr, "REGEXP") {
			t.Errorf("Expected REGEXP expression, got: %s", expr)
		}

		if strings.Contains(expr, "NOT REGEXP") {
			t.Errorf("Expected REGEXP without NOT, got: %s", expr)
		}
	}

	// Test NOT REGEXP
	for i := 0; i < testIterations; i++ {
		expr := eg.GenRegexpExpr(tbls, true)
		if expr == "" {
			t.Error("GenRegexpExpr returned empty expression")
		}

		if !strings.Contains(expr, "NOT REGEXP") {
			t.Errorf("Expected NOT REGEXP expression, got: %s", expr)
		}
	}
}

// TestGenMatchExpr tests MATCH expression generation
// Note: MATCH is not supported by Turso according to COMPAT.md
func TestGenMatchExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(53)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test MATCH
	for i := 0; i < testIterations; i++ {
		expr := eg.GenMatchExpr(tbls, false)
		if expr == "" {
			t.Error("GenMatchExpr returned empty expression")
		}

		if !strings.Contains(expr, "MATCH") {
			t.Errorf("Expected MATCH expression, got: %s", expr)
		}

		if strings.Contains(expr, "NOT MATCH") {
			t.Errorf("Expected MATCH without NOT, got: %s", expr)
		}
	}

	// Test NOT MATCH
	for i := 0; i < testIterations; i++ {
		expr := eg.GenMatchExpr(tbls, true)
		if expr == "" {
			t.Error("GenMatchExpr returned empty expression")
		}

		if !strings.Contains(expr, "NOT MATCH") {
			t.Errorf("Expected NOT MATCH expression, got: %s", expr)
		}
	}
}

// TestGenInSubqueryExpr tests IN (subquery) expression generation
// Note: IN (subquery) is not supported by Turso according to COMPAT.md
func TestGenInSubqueryExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(54)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test IN (subquery)
	for i := 0; i < testIterations; i++ {
		expr := eg.GenInSubqueryExpr(tbls, false)
		if expr == "" {
			t.Error("GenInSubqueryExpr returned empty expression")
		}

		if !strings.Contains(expr, " IN (") || !strings.Contains(expr, "SELECT") {
			t.Errorf("Expected IN (subquery) expression, got: %s", expr)
		}

		if strings.Contains(expr, "NOT IN") {
			t.Errorf("Expected IN without NOT, got: %s", expr)
		}
	}

	// Test NOT IN (subquery)
	for i := 0; i < testIterations; i++ {
		expr := eg.GenInSubqueryExpr(tbls, true)
		if expr == "" {
			t.Error("GenInSubqueryExpr returned empty expression")
		}

		if !strings.Contains(expr, "NOT IN (") || !strings.Contains(expr, "SELECT") {
			t.Errorf("Expected NOT IN (subquery) expression, got: %s", expr)
		}
	}
}

// TestGenExistsSubqueryExpr tests EXISTS (subquery) expression generation
// Note: EXISTS (subquery) is not supported by Turso according to COMPAT.md
func TestGenExistsSubqueryExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(55)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	// Test EXISTS (subquery)
	for i := 0; i < testIterations; i++ {
		expr := eg.GenExistsSubqueryExpr(tbls, false)
		if expr == "" {
			t.Error("GenExistsSubqueryExpr returned empty expression")
		}

		if !strings.Contains(expr, "EXISTS (") || !strings.Contains(expr, "SELECT") {
			t.Errorf("Expected EXISTS (subquery) expression, got: %s", expr)
		}

		if strings.Contains(expr, "NOT EXISTS") {
			t.Errorf("Expected EXISTS without NOT, got: %s", expr)
		}
	}

	// Test NOT EXISTS (subquery)
	for i := 0; i < testIterations; i++ {
		expr := eg.GenExistsSubqueryExpr(tbls, true)
		if expr == "" {
			t.Error("GenExistsSubqueryExpr returned empty expression")
		}

		if !strings.Contains(expr, "NOT EXISTS (") || !strings.Contains(expr, "SELECT") {
			t.Errorf("Expected NOT EXISTS (subquery) expression, got: %s", expr)
		}
	}
}

// TestGenFilterExpr tests aggregate FILTER clause expression generation
// Note: FILTER is not supported by Turso according to COMPAT.md (incorrectly ignored)
func TestGenFilterExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(56)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	for i := 0; i < testIterations; i++ {
		expr := eg.GenFilterExpr(tbls)
		if expr == "" {
			t.Error("GenFilterExpr returned empty expression")
		}

		if !strings.Contains(expr, "FILTER") || !strings.Contains(expr, "WHERE") {
			t.Errorf("Expected FILTER clause expression, got: %s", expr)
		}

		// Check for valid aggregate functions
		hasValidAgg := strings.Contains(expr, "COUNT") ||
			strings.Contains(expr, "SUM") ||
			strings.Contains(expr, "AVG") ||
			strings.Contains(expr, "MIN") ||
			strings.Contains(expr, "MAX")
		if !hasValidAgg {
			t.Errorf("Expected valid aggregate function in FILTER expression, got: %s", expr)
		}
	}
}

// TestGenOverExpr tests window function OVER clause expression generation
// Note: OVER is not supported by Turso according to COMPAT.md (incorrectly ignored)
func TestGenOverExpr(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	lcg := common.NewLCG(57)
	ctx := NewGenContext(db, lcg, 3)
	eg := NewExprGenerator(ctx)

	for i := 0; i < testIterations; i++ {
		expr := eg.GenOverExpr(tbls)
		if expr == "" {
			t.Error("GenOverExpr returned empty expression")
		}

		if !strings.Contains(expr, "OVER (") {
			t.Errorf("Expected OVER clause expression, got: %s", expr)
		}

		// Check for valid window functions
		hasValidFunc := strings.Contains(expr, "ROW_NUMBER") ||
			strings.Contains(expr, "RANK") ||
			strings.Contains(expr, "DENSE_RANK") ||
			strings.Contains(expr, "NTILE")
		if !hasValidFunc {
			t.Errorf("Expected valid window function in OVER expression, got: %s", expr)
		}
	}
}

// TestExpressionDeterminism tests that same seed produces same expressions
func TestExpressionDeterminism(t *testing.T) {
	db := setupTestDBForExpr(t)
	defer db.Close()

	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}

	tests := []struct {
		name string
		gen  func(*ExprGenerator, []helper.TableInfo) string
	}{
		{"Cast", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenCastExpr(tbls) }},
		{"Between", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenBetweenExpr(tbls, false) }},
		{"Glob", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenGlobExpr(tbls, false) }},
		// IS DISTINCT FROM is skipped - not supported by SQLite ANTLR parser
		{"Collate", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenCollateExpr(tbls) }},
		{"Unary", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenUnaryExpr(tbls) }},
		{"Binary", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenBinaryExpr(tbls) }},
		{"Parenthesized", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenParenthesizedExpr(tbls) }},
		{"IsNull", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenIsNullExpr(tbls, false) }},
		{"Regexp", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenRegexpExpr(tbls, false) }},
		{"Match", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenMatchExpr(tbls, false) }},
		{"InSubquery", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenInSubqueryExpr(tbls, false) }},
		{"ExistsSubquery", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenExistsSubqueryExpr(tbls, false) }},
		{"Filter", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenFilterExpr(tbls) }},
		{"Over", func(eg *ExprGenerator, tbls []helper.TableInfo) string { return eg.GenOverExpr(tbls) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lcg1 := common.NewLCG(12345)
			ctx1 := NewGenContext(db, lcg1, 3)
			eg1 := NewExprGenerator(ctx1)
			expr1 := tt.gen(eg1, tbls)

			lcg2 := common.NewLCG(12345)
			ctx2 := NewGenContext(db, lcg2, 3)
			eg2 := NewExprGenerator(ctx2)
			expr2 := tt.gen(eg2, tbls)

			if expr1 != expr2 {
				t.Errorf("Same seed produced different expressions:\n  First:  %s\n  Second: %s", expr1, expr2)
			}
		})
	}
}
