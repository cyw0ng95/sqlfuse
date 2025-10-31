package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
)

// GenSelectOrderBy generates a SELECT with ORDER BY on one or two columns.
func GenSelectOrderBy(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
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

	// pick columns
	cols := []helper.ColumnInfo{}
	cols = append(cols, tbl.Cols[rnd(len(tbl.Cols))])
	if len(tbl.Cols) > 1 && rnd(2) == 0 {
		cols = append(cols, tbl.Cols[rnd(len(tbl.Cols))])
	}

	// build order by
	order := stmts.QuoteIdent(cols[0].Name)
	if len(cols) > 1 {
		order += ", " + stmts.QuoteIdent(cols[1].Name)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s;", joinCols(cols), stmts.QuoteIdent(tbl.Name), order)
	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}
