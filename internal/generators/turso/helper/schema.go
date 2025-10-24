package helper

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

type ColumnInfo struct {
	Name string
	Type string
}

type TableInfo struct {
	Name string
	Cols []ColumnInfo
}

// GetAllTablesAndCols returns all user tables and their columns in the current database.
func GetAllTablesAndCols(db *sql.DB) ([]TableInfo, error) {
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
			var notnull, pk int
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
