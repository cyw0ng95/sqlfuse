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

// GenSelectWhere generates a SELECT with a WHERE clause (numeric or equality)
// lcg should be *common.LCG.
func GenSelectWhere(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	// reuse basic select selection logic
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

	// pick up to 3 columns
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

	// Try to find a numeric column among table columns
	numericCol := -1
	for i, c := range tbl.Cols {
		if isNumericType(c.Type) || containsTypeHintSimple(c.Name, "id", "num", "count", "amount") {
			numericCol = i
			break
		}
	}

	where := ""
	if numericCol >= 0 {
		val := types.ValueForType(tbl.Cols[numericCol].Type, lcg, tbl.Cols[numericCol].Name)
		where = fmt.Sprintf(" WHERE %s > %s", stmts.QuoteIdent(tbl.Cols[numericCol].Name), val)
	} else {
		// fallback to equality on first selected column
		val := types.ValueForType(cols[0].Type, lcg, cols[0].Name)
		where = fmt.Sprintf(" WHERE %s = %s", stmts.QuoteIdent(cols[0].Name), val)
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d;", joinCols(cols), stmts.QuoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}

func isNumericType(t string) bool {
	if t == "" {
		return false
	}
	up := strings.ToUpper(t)
	return strings.Contains(up, "INT") || strings.Contains(up, "REAL") || strings.Contains(up, "NUM") || strings.Contains(up, "FLOAT") || strings.Contains(up, "DOUBLE") || strings.Contains(up, "DEC")
}

func containsTypeHintSimple(name string, hints ...string) bool {
	nu := strings.ToUpper(name)
	for _, h := range hints {
		if strings.Contains(nu, strings.ToUpper(h)) {
			return true
		}
	}
	return false
}
