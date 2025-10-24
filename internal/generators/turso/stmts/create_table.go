package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
	"strings"
)

// CreateTableStmt represents a CREATE TABLE statement.
type CreateTableStmt struct {
	sql string
}

func (s *CreateTableStmt) SQL() string  { return s.sql }
func (s *CreateTableStmt) Type() string { return "create_table" }

// GenCreateTable generates a simple CREATE TABLE statement using the provided LCG.
// It produces 1..4 columns with common SQLite-compatible types.
func GenCreateTable(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// choose number of columns 1..4
	n := 1 + lcg.Intn(4)
	cols := make([]string, 0, n)
	types := []string{"INTEGER", "TEXT", "REAL", "BLOB"}
	for i := 0; i < n; i++ {
		colName := fmt.Sprintf("col%d", i+1)
		typeIdx := lcg.Intn(len(types))
		cols = append(cols, fmt.Sprintf("\"%s\" %s", colName, types[typeIdx]))
	}

	// generate a lightweight unique-ish table name
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s);", tbl, strings.Join(cols, ", "))
	return &CreateTableStmt{sql: sql}, nil
}
