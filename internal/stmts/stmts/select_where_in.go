package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
	"sqlsmith-go/internal/stmts/types"
	"strings"
)

// GenSelectWhereIn generates a SELECT with WHERE IN clause.
// IMPORTANT: Uses IN with value list only, NOT subqueries.
// Turso does NOT support IN (subquery) per COMPAT.md.
func GenSelectWhereIn(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: GetDefaultFlavor()}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tbls[rnd(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// pick up to 3 columns for selection
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

	// Pick a column for the IN clause
	inCol := tbl.Cols[rnd(len(tbl.Cols))]

	// Generate 2-5 values for the IN list
	numValues := 2 + rnd(4) // 2..5
	values := []string{}
	for i := 0; i < numValues; i++ {
		val := types.ValueForType(inCol.Type, lcg, inCol.Name)
		if val != "" {
			values = append(values, val)
		}
	}

	// Fallback if no values generated
	if len(values) == 0 {
		values = append(values, "NULL")
	}

	where := fmt.Sprintf(" WHERE %s IN (%s)", quoteIdent(inCol.Name), strings.Join(values, ", "))
	limit := 1 + rnd(50)

	sql := fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d;", joinCols(cols), quoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
