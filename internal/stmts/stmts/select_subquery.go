package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
	"sqlsmith-go/internal/stmts/types"
)

// GenSelectSubquery generates a SELECT with a subquery in the FROM clause (derived table).
// NOTE: Turso does NOT support EXISTS (subquery) or IN (subquery) per COMPAT.md.
// This generator uses subqueries in FROM clause as derived tables, which IS supported.
func GenSelectSubquery(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	// Pick a table for the subquery
	tbl := tbls[rnd(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name))}, nil
	}

	// Select columns from subquery (pick 1-3 columns)
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

	// Build a subquery in FROM clause (derived table)
	// This is supported: SELECT * FROM (SELECT ... FROM table) AS subq
	innerLimit := 1 + rnd(20)
	subquery := fmt.Sprintf("(SELECT %s FROM %s LIMIT %d) AS subq",
		joinCols(cols),
		quoteIdent(tbl.Name),
		innerLimit)

	// Optionally add a WHERE clause to the outer query
	var where string
	if rnd(2) == 0 && len(cols) > 0 {
		col := cols[rnd(len(cols))]
		val := types.ValueForType(col.Type, lcg, col.Name)
		where = fmt.Sprintf(" WHERE %s IS NOT NULL OR %s = %s", quoteIdent(col.Name), quoteIdent(col.Name), val)
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d;", subquery, where, limit)
	return SelectStmt{sql: sql}, nil
}
