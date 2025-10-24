package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/generators/turso/helper"
)

type SelectStmt struct {
	sql string
}

func (s *SelectStmt) SQL() string  { return s.sql }
func (s *SelectStmt) Type() string { return "select" }

// GenSelect generates a simple SELECT statement using available tables/columns.
// lcgOrRand should be an object implementing Intn(n int) int (e.g. *common.LCG).
func GenSelect(db *sql.DB, lcgOrRand interface{}) (SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		// propagate error so caller can decide fallback
		return SelectStmt{}, err
	}
	if len(tables) == 0 {
		// No real user tables available — return a harmless no-op select
		return SelectStmt{sql: "SELECT 3;"}, nil
	}

	// choose rnd function from provided generator
	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		// fallback to deterministic choice
		rnd = func(n int) int { return 0 }
	}

	// pick a table
	tbl := tables[rnd(len(tables))]
	// if no columns known, select all
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name))}, nil
	}

	// pick 1..min(3,len(cols)) columns using rnd
	maxCols := len(tbl.Cols)
	if maxCols > 3 {
		maxCols = 3
	}
	nCols := 1 + rnd(maxCols)
	if nCols > len(tbl.Cols) {
		nCols = len(tbl.Cols)
	}

	selected := make(map[int]struct{}, nCols)
	cols := make([]helper.ColumnInfo, 0, nCols)
	for len(cols) < nCols {
		idx := rnd(len(tbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		cols = append(cols, tbl.Cols[idx])
	}

	// Prefer a numeric column for WHERE if available
	numericIdx := -1
	for i, c := range cols {
		t := c.Type
		if t == "" {
			// unknown type -- check name for common hints
			if containsTypeHint(c.Name, "id", "count", "num", "amount") {
				numericIdx = i
				break
			}
			continue
		}
		up := stringsToUpper(t)
		if containsAny(up, "INT", "REAL", "NUM", "FLOAT", "DOUBLE", "DEC") {
			numericIdx = i
			break
		}
	}

	where := ""
	if numericIdx >= 0 {
		v := 1 + rnd(100)
		where = fmt.Sprintf(" WHERE %s > %d", quoteIdent(cols[numericIdx].Name), v)
	}

	limit := 1 + rnd(50)

	sql := fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d;", joinCols(cols), quoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql}, nil
}

func joinCols(cols []helper.ColumnInfo) string {
	q := ""
	for i, c := range cols {
		if i > 0 {
			q += ", "
		}
		q += quoteIdent(c.Name)
	}
	return q
}

func quoteIdent(s string) string {
	// use double quotes for identifiers
	return fmt.Sprintf("\"%s\"", s)
}

func stringsToUpper(s string) string {
	b := []byte(s)
	for i := range b {
		if 'a' <= b[i] && b[i] <= 'z' {
			b[i] = b[i] - ('a' - 'A')
		}
	}
	return string(b)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if stringsIndex(s, sub) >= 0 {
			return true
		}
	}
	return false
}

func containsTypeHint(name string, hints ...string) bool {
	nu := stringsToUpper(name)
	for _, h := range hints {
		if stringsIndex(nu, stringsToUpper(h)) >= 0 {
			return true
		}
	}
	return false
}

func stringsIndex(s, sub string) int {
	// simple implementation of strings.Index to avoid importing strings
	n := len(s)
	sn := len(sub)
	if sn == 0 {
		return 0
	}
	for i := 0; i+sn <= n; i++ {
		if s[i:i+sn] == sub {
			return i
		}
	}
	return -1
}
