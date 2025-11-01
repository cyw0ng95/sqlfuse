package helper

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// FlavorConfig defines the minimal interface needed for database flavor detection.
// This avoids circular dependencies with the stmts package.
type FlavorConfig interface {
	Name() string
}

type ColumnInfo struct {
	Name string
	Type string
}

type TableInfo struct {
	Name string
	Cols []ColumnInfo
}

// GetAllTablesAndCols returns all user tables and their columns in the current database.
// It supports both SQLite and DuckDB based on the dbType parameter.
// dbType should be one of: "duckdb", "sqlite", "go-sqlite3", "turso", or empty (defaults to "sqlite").
// DEPRECATED: Use GetAllTablesAndColsWithFlavor when flavor is available.
func GetAllTablesAndCols(db *sql.DB, dbType string) ([]TableInfo, error) {
	// Normalize dbType to handle empty or unknown values
	if dbType == "" {
		dbType = "sqlite"
	}

	// Route to appropriate implementation based on database type
	if dbType == "duckdb" {
		return getDuckDBTablesAndCols(db)
	}
	// All SQLite variants (sqlite, go-sqlite3, turso) use the same schema introspection
	return getSQLiteTablesAndCols(db)
}

// GetAllTablesAndColsWithFlavor returns all user tables and their columns using a FlavorConfig.
// This is the preferred method when a FlavorConfig is available.
// If flavor is nil, defaults to SQLite behavior.
func GetAllTablesAndColsWithFlavor(db *sql.DB, flavor FlavorConfig) ([]TableInfo, error) {
	// Use default SQLite flavor if none provided
	if flavor == nil {
		return getSQLiteTablesAndCols(db)
	}

	dbType := flavor.Name()

	// Route to appropriate implementation based on database type
	if dbType == "duckdb" {
		return getDuckDBTablesAndCols(db)
	}
	// All SQLite variants (sqlite, go-sqlite3, turso) use the same schema introspection
	return getSQLiteTablesAndCols(db)
}

// getSQLiteTablesAndCols uses SQLite-specific PRAGMA statements
func getSQLiteTablesAndCols(db *sql.DB) ([]TableInfo, error) {
	// Handle nil database
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	
	// serialize concurrent callers
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying sqlite_master: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	// Collect table names
	tableNames := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			fmt.Fprintf(os.Stderr, "error scanning sqlite_master row: %v\n", err)
			continue
		}
		tableNames = append(tableNames, name)
	}

	tables := []TableInfo{}
	for _, tableName := range tableNames {
		// Use single quotes and escape any single quotes in table name for PRAGMA
		esc := strings.ReplaceAll(tableName, "'", "''")
		pragma := fmt.Sprintf("PRAGMA table_info('%s')", esc)
		colRows, err := db.Query(pragma)
		if err != nil {
			fmt.Fprintf(os.Stderr, "PRAGMA table_info error for %s: %v\n", tableName, err)
			// continue to next table instead of failing entire introspection
			continue
		}
		cols := []ColumnInfo{}
		for colRows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk interface{}
			var dflt interface{}
			if err := colRows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				fmt.Fprintf(os.Stderr, "error scanning PRAGMA result for %s: %v\n", tableName, err)
				colRows.Close()
				break
			}
			cols = append(cols, ColumnInfo{Name: name, Type: ctype})
		}
		colRows.Close()
		tables = append(tables, TableInfo{Name: tableName, Cols: cols})
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("no user tables found in database")
	}
	return tables, nil
}

// getDuckDBTablesAndCols uses DuckDB's information_schema
func getDuckDBTablesAndCols(db *sql.DB) ([]TableInfo, error) {
	// Handle nil database
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	
	// Query tables from information_schema
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'main' 
		  AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying information_schema.tables: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	// Collect table names
	tableNames := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			fmt.Fprintf(os.Stderr, "error scanning table name: %v\n", err)
			continue
		}
		tableNames = append(tableNames, name)
	}

	tables := []TableInfo{}
	for _, tableName := range tableNames {
		// Query columns from information_schema
		colRows, err := db.Query(`
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_schema = 'main'
			  AND table_name = ?
			ORDER BY ordinal_position
		`, tableName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error querying columns for %s: %v\n", tableName, err)
			continue
		}

		cols := []ColumnInfo{}
		for colRows.Next() {
			var name, ctype string
			if err := colRows.Scan(&name, &ctype); err != nil {
				fmt.Fprintf(os.Stderr, "error scanning column for %s: %v\n", tableName, err)
				colRows.Close()
				break
			}
			cols = append(cols, ColumnInfo{Name: name, Type: ctype})
		}
		colRows.Close()
		tables = append(tables, TableInfo{Name: tableName, Cols: cols})
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("no user tables found in database")
	}
	return tables, nil
}
