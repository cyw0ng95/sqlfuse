package helper

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestGetAllTablesAndCols_WithSQLite3 tests that schema discovery works with SQLite3
// which returns int values (0/1) for notnull and pk columns in PRAGMA table_info
func TestGetAllTablesAndCols_WithSQLite3(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test schema discovery
	tables, err := GetAllTablesAndCols(db)
	if err != nil {
		t.Fatalf("GetAllTablesAndCols failed: %v", err)
	}

	// Verify results
	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	if tables[0].Name != "users" {
		t.Errorf("Expected table name 'users', got '%s'", tables[0].Name)
	}

	if len(tables[0].Cols) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(tables[0].Cols))
	}

	// Verify column names
	expectedCols := map[string]bool{"id": false, "name": false, "email": false}
	for _, col := range tables[0].Cols {
		if _, ok := expectedCols[col.Name]; !ok {
			t.Errorf("Unexpected column name: %s", col.Name)
		}
		expectedCols[col.Name] = true
	}

	for col, found := range expectedCols {
		if !found {
			t.Errorf("Expected column %s not found", col)
		}
	}
}

// TestGetAllTablesAndCols_NoTables tests the error case when no tables exist
func TestGetAllTablesAndCols_NoTables(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// No tables created, should return error
	_, err = GetAllTablesAndCols(db)
	if err == nil {
		t.Error("Expected error when no tables exist, got nil")
	}
}
