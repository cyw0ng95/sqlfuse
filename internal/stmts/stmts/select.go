package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
	"sqlsmith-go/internal/stmts/types"
)

// SelectGenerator is a StmtGenerator for basic SELECT statements.
type SelectGenerator struct{}

// Generate implements StmtGenerator for SELECT statements.
func (g *SelectGenerator) Generate(ctx *GenContext) (Stmt, error) {
	stmt, err := genSelectInternal(ctx.DB, ctx.LCG)
	return &stmt, err
}

// CanGenerate implements StmtGenerator. SELECT can be generated if tables exist or falls back to SELECT 1.
func (g *SelectGenerator) CanGenerate(ctx *GenContext) bool {
	return true // Can always generate SELECT 1 as fallback
}

type SelectStmt struct {
	sql string
}

func (s *SelectStmt) SQL() string  { return s.sql }
func (s *SelectStmt) Type() string { return "select" }

// GenSelect generates a simple SELECT statement using available tables/columns.
// This function is kept for backward compatibility with existing code.
// lcg should be *common.LCG.
func GenSelect(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	return genSelectInternal(db, lcg)
}

// genSelectInternal is the internal implementation used by both old and new interfaces.
func genSelectInternal(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		// No real user tables available — return a harmless no-op select
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	// choose rnd function from provided generator
	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
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
		val := types.ValueForType(cols[numericIdx].Type, lcg, cols[numericIdx].Name)
		where = fmt.Sprintf(" WHERE %s > %s", quoteIdent(cols[numericIdx].Name), val)
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
