package helper

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

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
