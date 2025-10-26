package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
	"sqlsmith-go/internal/generators/turso/types"
	"strings"
)

// GenSelectCase generates a SELECT with CASE expressions in the projection.
func GenSelectCase(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
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

	tbl := tbls[rnd(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name))}, nil
	}

	// Pick a column for the CASE expression
	caseCol := tbl.Cols[rnd(len(tbl.Cols))]

	// Build a CASE expression
	var caseExpr string
	if isNumericType(caseCol.Type) || containsTypeHintSimple(caseCol.Name, "id", "num", "count", "amount", "age") {
		// Numeric CASE expression
		val1 := types.ValueForType(caseCol.Type, lcg, caseCol.Name)
		val2 := types.ValueForType(caseCol.Type, lcg, caseCol.Name)
		caseExpr = fmt.Sprintf("CASE WHEN %s < %s THEN 'low' WHEN %s >= %s AND %s < %s THEN 'medium' ELSE 'high' END AS %s_category",
			quoteIdent(caseCol.Name), val1,
			quoteIdent(caseCol.Name), val1, quoteIdent(caseCol.Name), val2,
			caseCol.Name)
	} else {
		// Text/generic CASE expression
		val := types.ValueForType(caseCol.Type, lcg, caseCol.Name)
		caseExpr = fmt.Sprintf("CASE WHEN %s = %s THEN 'matched' WHEN %s IS NULL THEN 'null' ELSE 'other' END AS %s_status",
			quoteIdent(caseCol.Name), val,
			quoteIdent(caseCol.Name),
			caseCol.Name)
	}

	// Select a few other columns along with the CASE
	selectCols := []string{caseExpr}
	maxCols := len(tbl.Cols)
	if maxCols > 2 {
		maxCols = 2
	}
	nCols := 1 + rnd(maxCols)

	selected := make(map[int]struct{})
	for i := 0; i < nCols && len(selectCols) < 4; i++ {
		idx := rnd(len(tbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		selectCols = append(selectCols, quoteIdent(tbl.Cols[idx].Name))
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", strings.Join(selectCols, ", "), quoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql}, nil
}
