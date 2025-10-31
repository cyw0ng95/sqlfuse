package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
	"sqlfuse/internal/stmts/types"
	"strings"
)

// GenSelectCase generates a SELECT with CASE expressions in the projection.
func GenSelectCase(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tbls[rnd(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", stmts.QuoteIdent(tbl.Name)), flavor: stmts.GetDefaultFlavor()}, nil
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
			stmts.QuoteIdent(caseCol.Name), val1,
			stmts.QuoteIdent(caseCol.Name), val1, stmts.QuoteIdent(caseCol.Name), val2,
			caseCol.Name)
	} else {
		// Text/generic CASE expression
		val := types.ValueForType(caseCol.Type, lcg, caseCol.Name)
		caseExpr = fmt.Sprintf("CASE WHEN %s = %s THEN 'matched' WHEN %s IS NULL THEN 'null' ELSE 'other' END AS %s_status",
			stmts.QuoteIdent(caseCol.Name), val,
			stmts.QuoteIdent(caseCol.Name),
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
		selectCols = append(selectCols, stmts.QuoteIdent(tbl.Cols[idx].Name))
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", strings.Join(selectCols, ", "), stmts.QuoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}
