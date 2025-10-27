package stmts

import (
	"testing"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
)

// TestSelectBuilder tests the basic SELECT builder functionality
func TestSelectBuilder(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	// Test basic SELECT
	builder := NewSelectBuilder(ctx).
		Select("id", "name").
		From("users").
		Limit(10)

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "SELECT id, name FROM \"users\" LIMIT 10;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestSelectBuilderWithWhere tests SELECT with WHERE clause
func TestSelectBuilderWithWhere(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewSelectBuilder(ctx).
		Select("id", "name").
		From("users").
		Where("age > 18").
		Where("active = 1").
		Limit(5)

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "SELECT id, name FROM \"users\" WHERE age > 18 AND active = 1 LIMIT 5;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestSelectBuilderWithGroupBy tests SELECT with GROUP BY and HAVING
func TestSelectBuilderWithGroupBy(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewSelectBuilder(ctx).
		Select("department", "COUNT(*)").
		From("employees").
		GroupBy("department").
		Having("COUNT(*) > 5").
		Limit(10)

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "SELECT department, COUNT(*) FROM \"employees\" GROUP BY department HAVING COUNT(*) > 5 LIMIT 10;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestSelectBuilderWithOrderBy tests SELECT with ORDER BY
func TestSelectBuilderWithOrderBy(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewSelectBuilder(ctx).
		Select("name", "age").
		From("users").
		OrderBy("age DESC", "name ASC").
		Limit(20)

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "SELECT name, age FROM \"users\" ORDER BY age DESC, name ASC LIMIT 20;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestSelectBuilderWithSubquery tests SELECT with subquery in FROM
func TestSelectBuilderWithSubquery(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewSelectBuilder(ctx).
		Select("*").
		FromSubquery("SELECT id, name FROM users WHERE active = 1", "active_users").
		Limit(10)

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "SELECT * FROM (SELECT id, name FROM users WHERE active = 1) AS \"active_users\" LIMIT 10;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestSelectBuilderWithOffset tests SELECT with LIMIT and OFFSET
func TestSelectBuilderWithOffset(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewSelectBuilder(ctx).
		Select("*").
		From("products").
		Limit(10).
		Offset(20)

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "SELECT * FROM \"products\" LIMIT 10 OFFSET 20;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestSelectBuilderNoFrom tests that Build fails without FROM clause
func TestSelectBuilderNoFrom(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewSelectBuilder(ctx).
		Select("1")

	_, err := builder.Build()
	if err == nil {
		t.Error("Build should fail without FROM clause")
	}
}

// TestSelectBuilderSelectColumns tests SelectColumns method
func TestSelectBuilderSelectColumns(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	table := helper.TableInfo{
		Name: "users",
		Cols: []helper.ColumnInfo{
			{Name: "id", Type: "INTEGER"},
			{Name: "name", Type: "TEXT"},
			{Name: "email", Type: "TEXT"},
			{Name: "age", Type: "INTEGER"},
		},
	}

	builder := NewSelectBuilder(ctx).
		SelectColumns(table, 2).
		FromTable(table).
		Limit(5)

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should select 2 columns
	sql := stmt.SQL()
	if sql == "" {
		t.Error("BuildSQL returned empty string")
	}
	// Verify it contains FROM users
	if !contains(sql, "FROM \"users\"") {
		t.Errorf("SQL should contain FROM \"users\", got: %s", sql)
	}
}

// TestInsertBuilder tests the basic INSERT builder functionality
func TestInsertBuilder(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewInsertBuilder(ctx).
		Into("users").
		Columns("name", "email").
		Values("'Alice'", "'alice@example.com'")

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "INSERT INTO \"users\" (\"name\", \"email\") VALUES ('Alice', 'alice@example.com');"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestInsertBuilderMultipleRows tests INSERT with multiple rows
func TestInsertBuilderMultipleRows(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewInsertBuilder(ctx).
		Into("users").
		Columns("name", "email").
		Values("'Alice'", "'alice@example.com'").
		Values("'Bob'", "'bob@example.com'")

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "INSERT INTO \"users\" (\"name\", \"email\") VALUES ('Alice', 'alice@example.com'), ('Bob', 'bob@example.com');"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestInsertBuilderDefaultValues tests INSERT with DEFAULT VALUES
func TestInsertBuilderDefaultValues(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewInsertBuilder(ctx).
		Into("users")

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "INSERT INTO \"users\" DEFAULT VALUES;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestInsertBuilderOnConflict tests INSERT with ON CONFLICT
func TestInsertBuilderOnConflict(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewInsertBuilder(ctx).
		Into("users").
		Columns("email", "name").
		Values("'alice@example.com'", "'Alice'").
		OnConflict("ON CONFLICT(email) DO UPDATE SET name=excluded.name")

	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := "INSERT INTO \"users\" (\"email\", \"name\") VALUES ('alice@example.com', 'Alice') ON CONFLICT(email) DO UPDATE SET name=excluded.name;"
	if stmt.SQL() != expected {
		t.Errorf("Expected %q, got %q", expected, stmt.SQL())
	}
}

// TestInsertBuilderNoTable tests that Build fails without table name
func TestInsertBuilderNoTable(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	builder := NewInsertBuilder(ctx).
		Columns("name").
		Values("'Alice'")

	_, err := builder.Build()
	if err == nil {
		t.Error("Build should fail without table name")
	}
}

// TestRandomSelectBuilder tests the random SELECT builder
func TestRandomSelectBuilder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(123)
	ctx := NewGenContext(db, lcg, 2)

	builder := NewRandomSelectBuilder(ctx, db)
	stmt, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if stmt == nil {
		t.Error("Build returned nil statement")
	}

	sql := stmt.SQL()
	if sql == "" {
		t.Error("Build returned empty SQL")
	}

	// Verify SQL is valid
	valid, errors := ValidateSQL(sql)
	if !valid {
		t.Errorf("Generated invalid SQL: %s\nErrors: %v", sql, errors)
	}

	// Verify it's executable
	rows, err := db.Query(sql)
	if err != nil {
		t.Errorf("Failed to execute generated SQL: %v\nSQL: %s", err, sql)
	}
	if rows != nil {
		rows.Close()
	}
}

// TestBuilderChaining tests method chaining works correctly
func TestBuilderChaining(t *testing.T) {
	lcg := common.NewLCG(42)
	ctx := NewGenContext(nil, lcg, 2)

	// Test that we can chain methods fluently
	sql, err := NewSelectBuilder(ctx).
		Select("id", "name", "email").
		From("users").
		Where("active = 1").
		Where("age > 18").
		OrderBy("name ASC").
		Limit(10).
		BuildSQL()

	if err != nil {
		t.Fatalf("BuildSQL failed: %v", err)
	}

	expected := "SELECT id, name, email FROM \"users\" WHERE active = 1 AND age > 18 ORDER BY name ASC LIMIT 10;"
	if sql != expected {
		t.Errorf("Expected %q, got %q", expected, sql)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
