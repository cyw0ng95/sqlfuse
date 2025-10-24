package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/generators/turso/helper"
	"strings"
)

// InsertStmt represents an INSERT statement.
type InsertStmt struct {
	sql string
}

func (s *InsertStmt) SQL() string  { return s.sql }
func (s *InsertStmt) Type() string { return "insert" }

// GenInsert generates a single-row INSERT (existing behavior).
func GenInsert(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	return genInsertInternal(db, lcgOrRand, 1)
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
	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		rnd = func(n int) int { return 0 }
	}

	var u64 func() uint64
	if u, ok := lcgOrRand.(interface{ Uint64() uint64 }); ok {
		u64 = u.Uint64
	} else {
		u64 = func() uint64 { return uint64(rnd(1 << 30)) }
	}

	vals := []string{}
	for _, c := range cols {
		t := stringsToUpper(c.Type)
		var val string
		if t == "" {
			if containsTypeHint(c.Name, "id", "count", "num", "qty", "amount") {
				val = fmt.Sprintf("%d", 1+rnd(1000))
			} else {
				val = fmt.Sprintf("'%s'", escapeSingle(fmt.Sprintf("s%08x", u64())))
			}
		} else if containsAny(t, "INT") {
			val = fmt.Sprintf("%d", 1+rnd(100000))
		} else if containsAny(t, "REAL", "FLOA", "DOUB", "DEC", "NUM") {
			val = fmt.Sprintf("%f", float64(rnd(100000))/100.0)
		} else if containsAny(t, "CHAR", "CLOB", "TEXT") {
			val = fmt.Sprintf("'%s'", escapeSingle(fmt.Sprintf("s%08x", u64())))
		} else if containsAny(t, "BLOB") {
			val = fmt.Sprintf("X'%016x'", u64())
		} else {
			val = fmt.Sprintf("'%s'", escapeSingle(fmt.Sprintf("s%08x", u64())))
		}
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
