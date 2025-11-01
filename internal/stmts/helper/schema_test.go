package helper

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// MockFlavorConfig implements FlavorConfig for testing
type MockFlavorConfig struct {
	name string
}

func (m *MockFlavorConfig) Name() string {
	return m.name
}

// TestGetAllTablesAndColsWithDBType verifies that GetAllTablesAndCols
// correctly uses the passed database type instead of detecting it.
func TestGetAllTablesAndColsWithDBType(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			value REAL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	testCases := []struct {
		name     string
		dbType   string
		wantErr  bool
		wantTbls int
	}{
		{
			name:     "SQLite type",
			dbType:   "sqlite",
			wantErr:  false,
			wantTbls: 1,
		},
		{
			name:     "go-sqlite3 type",
			dbType:   "go-sqlite3",
			wantErr:  false,
			wantTbls: 1,
		},
		{
			name:     "turso type",
			dbType:   "turso",
			wantErr:  false,
			wantTbls: 1,
		},
		{
			name:     "empty string defaults to sqlite",
			dbType:   "",
			wantErr:  false,
			wantTbls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tables, err := GetAllTablesAndCols(db, tc.dbType)
			if tc.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(tables) != tc.wantTbls {
				t.Errorf("Expected %d tables, got %d", tc.wantTbls, len(tables))
			}
			if len(tables) > 0 {
				if tables[0].Name != "test_table" {
					t.Errorf("Expected table name 'test_table', got '%s'", tables[0].Name)
				}
				if len(tables[0].Cols) != 3 {
					t.Errorf("Expected 3 columns, got %d", len(tables[0].Cols))
				}
			}
		})
	}
}

// TestGetAllTablesAndColsNoAutoDetection verifies that the function
// no longer attempts to auto-detect the database type by probing.
func TestGetAllTablesAndColsNoAutoDetection(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER, name TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Call with explicit sqlite type - should work without probing
	tables, err := GetAllTablesAndCols(db, "sqlite")
	if err != nil {
		t.Fatalf("GetAllTablesAndCols failed: %v", err)
	}

	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}

	// Verify that passing "duckdb" for an SQLite database will try DuckDB introspection
	// This should fail because SQLite doesn't have information_schema
	_, err = GetAllTablesAndCols(db, "duckdb")
	if err == nil {
		t.Error("Expected error when using duckdb type on SQLite database, but got nil")
	}
}

// TestGetAllTablesAndColsNilDB verifies that passing a nil database
// returns an error instead of panicking.
func TestGetAllTablesAndColsNilDB(t *testing.T) {
	// Test with SQLite
	_, err := GetAllTablesAndCols(nil, "sqlite")
	if err == nil {
		t.Error("Expected error when passing nil database for sqlite, but got nil")
	}
	if err != nil && err.Error() != "database connection is nil" {
		t.Errorf("Expected 'database connection is nil' error, got: %v", err)
	}

	// Test with DuckDB
	_, err = GetAllTablesAndCols(nil, "duckdb")
	if err == nil {
		t.Error("Expected error when passing nil database for duckdb, but got nil")
	}
	if err != nil && err.Error() != "database connection is nil" {
		t.Errorf("Expected 'database connection is nil' error, got: %v", err)
	}

	// Test with empty string (defaults to SQLite)
	_, err = GetAllTablesAndCols(nil, "")
	if err == nil {
		t.Error("Expected error when passing nil database with empty dbType, but got nil")
	}
	if err != nil && err.Error() != "database connection is nil" {
		t.Errorf("Expected 'database connection is nil' error, got: %v", err)
	}
}

// TestGetAllTablesAndColsWithFlavor_NilFlavor tests nil flavor handling
func TestGetAllTablesAndColsWithFlavor_NilFlavor(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test (id INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Nil flavor should default to SQLite behavior
	tables, err := GetAllTablesAndColsWithFlavor(db, nil)
	if err != nil {
		t.Errorf("Expected no error with nil flavor, got: %v", err)
	}

	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}
}

// TestGetAllTablesAndColsWithFlavor_SQLiteFlavors tests SQLite flavor variants
func TestGetAllTablesAndColsWithFlavor_SQLiteFlavors(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE products (id INTEGER, name TEXT, price REAL)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	flavors := []string{"sqlite", "go-sqlite3", "turso"}

	for _, flavorName := range flavors {
		t.Run(flavorName, func(t *testing.T) {
			flavor := &MockFlavorConfig{name: flavorName}
			tables, err := GetAllTablesAndColsWithFlavor(db, flavor)
			if err != nil {
				t.Errorf("Unexpected error for %s flavor: %v", flavorName, err)
			}

			if len(tables) != 1 {
				t.Errorf("Expected 1 table, got %d", len(tables))
			}

			if len(tables) > 0 {
				if tables[0].Name != "products" {
					t.Errorf("Expected table 'products', got '%s'", tables[0].Name)
				}
				if len(tables[0].Cols) != 3 {
					t.Errorf("Expected 3 columns, got %d", len(tables[0].Cols))
				}
			}
		})
	}
}

// TestGetAllTablesAndCols_SpecialCharacters tests table names with special characters
func TestGetAllTablesAndCols_SpecialCharacters(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create table with single quote in name
	_, err = db.Exec(`CREATE TABLE "test'table" (id INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tables, err := GetAllTablesAndCols(db, "sqlite")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}

	if len(tables) > 0 && tables[0].Name != "test'table" {
		t.Errorf("Expected table name with quote, got '%s'", tables[0].Name)
	}
}

// TestGetAllTablesAndCols_MultipleTables tests handling of multiple tables
func TestGetAllTablesAndCols_MultipleTables(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER, name TEXT);
		CREATE TABLE orders (id INTEGER, user_id INTEGER, amount REAL);
		CREATE TABLE products (id INTEGER, title TEXT);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	tables, err := GetAllTablesAndCols(db, "sqlite")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(tables) != 3 {
		t.Errorf("Expected 3 tables, got %d", len(tables))
	}

	// Verify each table has columns
	for _, table := range tables {
		if len(table.Cols) == 0 {
			t.Errorf("Table %s has no columns", table.Name)
		}
	}
}

// TestGetAllTablesAndCols_EmptyDatabase tests database with no user tables
func TestGetAllTablesAndCols_EmptyDatabase(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = GetAllTablesAndCols(db, "sqlite")
	if err == nil {
		t.Error("Expected error for empty database, got nil")
	}
	
	if err != nil && err.Error() != "no user tables found in database" {
		t.Errorf("Expected 'no user tables found' error, got: %v", err)
	}
}

// TestColumnInfo_StructFields tests ColumnInfo struct fields
func TestColumnInfo_StructFields(t *testing.T) {
	col := ColumnInfo{
		Name: "test_col",
		Type: "INTEGER",
	}

	if col.Name != "test_col" {
		t.Errorf("Expected column name 'test_col', got '%s'", col.Name)
	}

	if col.Type != "INTEGER" {
		t.Errorf("Expected column type 'INTEGER', got '%s'", col.Type)
	}
}

// TestTableInfo_StructFields tests TableInfo struct fields
func TestTableInfo_StructFields(t *testing.T) {
	table := TableInfo{
		Name: "test_table",
		Cols: []ColumnInfo{
			{Name: "id", Type: "INTEGER"},
			{Name: "name", Type: "TEXT"},
		},
	}

	if table.Name != "test_table" {
		t.Errorf("Expected table name 'test_table', got '%s'", table.Name)
	}

	if len(table.Cols) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(table.Cols))
	}
}

