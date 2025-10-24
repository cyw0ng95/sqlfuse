package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
	"sqlsmith-go/internal/generators/turso/types"
	"strings"
)

// InsertStmt represents an INSERT statement.
type InsertStmt struct {
	sql string
}

func (s *InsertStmt) SQL() string  { return s.sql }
func (s *InsertStmt) Type() string { return "insert" }

// GenInsert generates a type-aware INSERT for a random user table.
// lcg should be *common.LCG.
func GenInsert(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables available for INSERT: %v", err)
	}

	// rnd and u64 helpers
	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	// pick a table
	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		// no columns -> use DEFAULT VALUES
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", quoteIdent(tbl.Name))}, nil
	}

	cols := []string{}
	vals := []string{}
	for _, c := range tbl.Cols {
		// Skip common autoincrement/id columns by name
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		cols = append(cols, quoteIdent(c.Name))
		val := types.ValueForType(c.Type, lcg, c.Name)
		// If ValueForType returns an unquoted numeric, keep as is; it returns quoted strings already
		vals = append(vals, val)
	}

	if len(cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", quoteIdent(tbl.Name))}, nil
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", quoteIdent(tbl.Name), strings.Join(cols, ", "), strings.Join(vals, ", "))
	return &InsertStmt{sql: sql}, nil
}

func quoteIdent(s string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(s, "\"", "\"\""))
}

// GenInsertMultiple generates an INSERT with multiple VALUES rows (2..N rows).
func GenInsertMultiple(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	// pick number of rows 2..5
	n := 2
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		n = 2 + r.Intn(4) // 2..5
	}
	return genInsertInternal(db, lcgOrRand, n)
}

// GenUpsert generates an INSERT ... ON CONFLICT(...) DO UPDATE statement when possible.
func GenUpsert(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for upsert: %v", err)
	}

	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", quoteIdent(tbl.Name))}, nil
	}

	// choose a subset of non-PK columns for insert (skip id if present)
	cols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		cols = append(cols, c)
	}
	if len(cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", quoteIdent(tbl.Name))}, nil
	}

	// pick conflict target: try to use first UNIQUE-like column name (email, key) else use id
	conflictCol := ""
	for _, c := range tbl.Cols {
		n := strings.ToLower(c.Name)
		if n == "email" || n == "key" || n == "name" {
			conflictCol = c.Name
			break
		}
	}
	if conflictCol == "" {
		// if id exists use id, else use first column
		for _, c := range tbl.Cols {
			if strings.EqualFold(c.Name, "id") {
				conflictCol = c.Name
				break
			}
		}
		if conflictCol == "" {
			conflictCol = tbl.Cols[0].Name
		}
	}

	// single row values
	vals, err := buildValuesRow(cols, lcgOrRand)
	if err != nil {
		return nil, err
	}

	// build set clause updating non-conflict columns
	setParts := []string{}
	for _, c := range cols {
		if c.Name == conflictCol {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s=excluded.%s", quoteIdent(c.Name), quoteIdent(c.Name)))
	}
	setClause := ""
	if len(setParts) > 0 {
		setClause = " DO UPDATE SET " + strings.Join(setParts, ", ")
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT(%s)%s;", quoteIdent(tbl.Name), strings.Join(colsNames(cols), ", "), vals, quoteIdent(conflictCol), setClause)
	return &InsertStmt{sql: sql}, nil
}

// GenInsertFromSelect generates INSERT INTO t(cols) SELECT ... FROM other_table LIMIT n
func GenInsertFromSelect(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) < 1 {
		return nil, fmt.Errorf("no tables for insert-select: %v", err)
	}

	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		rnd = func(n int) int { return 0 }
	}

	target := tables[rnd(len(tables))]
	source := tables[rnd(len(tables))]
	// ensure at least one column
	if len(target.Cols) == 0 || len(source.Cols) == 0 {
		return nil, fmt.Errorf("tables lack columns for insert-select")
	}

	// pick up to 3 columns present in both (by position/name)
	n := 1
	if len(target.Cols) > 1 {
		n = 1 + rnd(min(3, len(target.Cols)))
	}
	cols := target.Cols[:min(n, len(target.Cols))]
	// build SELECT projection from source: reuse names but quote
	selectCols := []string{}
	for _, c := range cols {
		selectCols = append(selectCols, quoteIdent(c.Name))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s LIMIT %d;", quoteIdent(target.Name), strings.Join(colsNames(cols), ", "), strings.Join(selectCols, ", "), quoteIdent(source.Name), 1+rnd(10))
	return &InsertStmt{sql: sql}, nil
}

// genInsertInternal generates n rows inserted into a chosen table.
func genInsertInternal(db *sql.DB, lcgOrRand interface{}, rowsCount int) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for insert: %v", err)
	}

	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", quoteIdent(tbl.Name))}, nil
	}

	// choose columns excluding common autoincrement 'id'
	cols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		cols = append(cols, c)
	}
	if len(cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", quoteIdent(tbl.Name))}, nil
	}

	rowsVals := []string{}
	for rIdx := 0; rIdx < rowsCount; rIdx++ {
		vals, err := buildValuesRow(cols, lcgOrRand)
		if err != nil {
			return nil, err
		}
		rowsVals = append(rowsVals, fmt.Sprintf("(%s)", vals))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;", quoteIdent(tbl.Name), strings.Join(colsNames(cols), ", "), strings.Join(rowsVals, ", "))
	return &InsertStmt{sql: sql}, nil
}

// buildValuesRow returns comma-joined values for the provided columns using lcgOrRand
func buildValuesRow(cols []helper.ColumnInfo, lcgOrRand interface{}) (string, error) {
	// obtain a *common.LCG to pass into types.ValueForType
	var lcg *common.LCG
	switch v := lcgOrRand.(type) {
	case *common.LCG:
		lcg = v
	case interface{ Uint64() uint64 }:
		// seed a new LCG from the provided Uint64 source
		lcg = common.NewLCG(v.Uint64())
	default:
		lcg = common.NewLCG(1)
	}

	vals := []string{}
	for _, c := range cols {
		val := types.ValueForType(c.Type, lcg, c.Name)
		vals = append(vals, val)
	}
	return strings.Join(vals, ", "), nil
}

func colsNames(cols []helper.ColumnInfo) []string {
	n := make([]string, len(cols))
	for i, c := range cols {
		n[i] = quoteIdent(c.Name)
	}
	return n
}

func escapeSingle(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
